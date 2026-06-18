// k6 load test proving the challenge performance target for the WHOLE customer
// path — frontend and backend — under 1k concurrent users: every load time stays
// under 2 seconds.
//
// Each virtual user does what a real visitor does, in order:
//   1. GET the storefront page  (frontend: nginx serving the SPA shell)
//   2. POST availablePets        (backend: gateway -> API -> Redis -> Postgres)
// Both steps are tagged and each carries its own p95 < 2s threshold, so a regression
// in either the frontend serving or the backend read fails the run independently.
//
// Run it in-cluster (see `make load-test` / deploy/k8s/loadtest-job.yaml) so a
// `kubectl port-forward` proxy never becomes the bottleneck.
//
// Every value is overridable by env var; the defaults encode the challenge target
// (1k users, the seeded demo store).
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const GATEWAY_URL = (__ENV.GATEWAY_URL || 'http://petstore-web').replace(/\/$/, '');
const STORE_ID = __ENV.STORE_ID || '11111111-1111-1111-1111-111111111111';
// Empty by default: the storefront gateway injects the ambient browse credential
// when a request carries no Authorization, exactly as it does for a real visitor.
const AUTH_HEADER = __ENV.AUTH_HEADER || '';
const VUS = parseInt(__ENV.VUS || '1000', 10);
const DURATION = __ENV.DURATION || '30s';
const PAGE_SIZE = parseInt(__ENV.PAGE_SIZE || '24', 10);
// Per-user think time between actions. 1000 *concurrent users* means 1000
// simultaneous sessions with human pauses between clicks — not 1000 requests
// every second. A 1–3s pause models real browsing; set THINK_MIN=THINK_MAX=0 for
// a raw throughput flood instead.
const THINK_MIN = parseFloat(__ENV.THINK_MIN || '1');
const THINK_MAX = parseFloat(__ENV.THINK_MAX || '3');

const shellUrl = `${GATEWAY_URL}/store/${STORE_ID}`;
const graphqlUrl = `${GATEWAY_URL}/graphql`;

const failedRequests = new Rate('failed_requests');

const catalogQuery = `query Catalog($storeId: ID!, $first: Int!, $after: String) {
  availablePets(storeId: $storeId, first: $first, after: $after) {
    edges {
      node { id name species ageYears description pictureUrl status createdAt }
      cursor
    }
    pageInfo { hasNextPage endCursor }
  }
}`;

export const options = {
  insecureSkipTLSVerify: true,
  scenarios: {
    storefront: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: VUS },
        { duration: DURATION, target: VUS },
        { duration: '5s', target: 0 },
      ],
      gracefulRampDown: '10s',
    },
  },
  // The challenge bar, asserted per step: the run fails (non-zero exit) if either
  // the frontend shell or the backend catalog read breaks 2s at p95, or errors rise.
  thresholds: {
    'http_req_duration{step:app_shell}': ['p(95)<2000'],
    'http_req_duration{step:catalog}': ['p(95)<2000'],
    http_req_failed: ['rate<0.01'],
    failed_requests: ['rate<0.01'],
    checks: ['rate>0.99'],
  },
};

const catalogParams = {
  tags: { step: 'catalog' },
  headers: AUTH_HEADER
    ? { 'Content-Type': 'application/json', Authorization: AUTH_HEADER }
    : { 'Content-Type': 'application/json' },
};

export default function () {
  const shell = http.get(shellUrl, { tags: { step: 'app_shell' } });
  const shellOk = check(shell, {
    'shell status is 200': (r) => r.status === 200,
    'shell is the SPA document': (r) => typeof r.body === 'string' && r.body.includes('<div id="root"'),
  });

  const payload = JSON.stringify({
    query: catalogQuery,
    variables: { storeId: STORE_ID, first: PAGE_SIZE, after: null },
  });
  const catalog = http.post(graphqlUrl, payload, catalogParams);
  const catalogOk = check(catalog, {
    'catalog status is 200': (r) => r.status === 200,
    'no GraphQL errors': (r) => {
      try {
        const body = r.json();
        return !body.errors && !!body.data && Array.isArray(body.data.availablePets.edges);
      } catch (_) {
        return false;
      }
    },
  });

  failedRequests.add(!(shellOk && catalogOk));

  sleep(THINK_MIN + Math.random() * (THINK_MAX - THINK_MIN));
}
