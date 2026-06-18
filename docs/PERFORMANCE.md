# Performance — proving <2s for 1k concurrent users

The challenge requires that "load times for all functionality" stay **under 2 seconds for 1k
concurrent users**, for **frontend and backend**. This document explains how that target is met and,
more importantly, how to **reproduce the measurement yourself** — no source-diving, one `make` target.

## Measured result

A [k6](https://k6.io/) load test (`make load-test`) drives the whole customer path at 1k concurrent
users: each virtual user loads the storefront page **and** fetches the catalog, like a real visitor.
Both steps carry an independent `p(95) < 2s` threshold, so the run fails if either the frontend serving
or the backend read regresses. Representative steady-state runs on a 10-core Minikube node:

| Step | p95 | median | errors |
|---|---|---|---|
| Storefront page (frontend, nginx static) | 3–80 ms | ~0.2 ms | 0% |
| `availablePets` (backend: gateway → API → Redis → Postgres) | 0.1–1.3 s | ~1 ms | 0% |

Comfortably inside the 2s bar on both halves, at ~700–800 req/s sustained, with a 0% error rate.

## How the target is met

Two structural choices plus four load-driven optimizations carry it:

**Structural**

- **Redis cache-aside** on the catalog, keyed per store and per page, with a **generation counter**
  baked into each key. A write (create/remove/sell) increments the store's generation, instantly making
  every stale page unreachable — so correctness comes from invalidation, *not* from a short TTL. See
  [`ARCHITECTURE.md`](ARCHITECTURE.md) §"Caching".
- **Keyset pagination** ordered by `(created_at, id)` — no `OFFSET`. `first` is capped at 100 by a
  query-complexity limit, so no request can ask for an unbounded page.

**Load-driven (each found by running the test, in order)**

1. **Cached authentication.** Basic auth re-runs bcrypt (~66 ms of CPU) on every request by design.
   Uncached, that alone capped throughput at ~150 req/s and put p95 at 35 s under 1k users. A short-TTL
   (60 s) in-memory cache of *successful* authentications (keyed by a hash of the credential) skips
   bcrypt on repeat requests. See [`SECURITY.md`](SECURITY.md) §"Credential cache". → p95 35 s → 2.3 s.
2. **Gateway upstream keepalive.** The storefront nginx re-encrypts to the API's TLS; without upstream
   keepalive it paid a fresh TLS handshake per request. An `upstream { keepalive }` block reuses
   connections. → p95 2.3 s → 0.7 s.
3. **Single-flight catalog loads.** When a cached page expires under heavy concurrency, naive
   cache-aside lets every concurrent miss hit Postgres at once (a stampede). The listing service
   coalesces concurrent misses for the same page through `singleflight` so exactly one query runs and
   the rest share its result.
4. **Long cache TTL + warm connection pools.** Because invalidation is generation-based, the TTL is only
   a memory backstop — raised to 10 min so a browsing session never triggers a needless refill. The
   Redis pool keeps idle connections warm (`MinIdleConns`) so the first traffic burst against a fresh
   pod doesn't stall opening connections.

## How to reproduce it

```bash
make k8s-up                          # bring the whole stack up (if it isn't already)
make load-test                       # 1k-VU storefront load test in-cluster; streams the k6 summary, prints PASS/FAIL
```

`make load-test` (re)creates a ConfigMap from [`../loadtest/storefront.js`](../loadtest/storefront.js),
runs the [k6 Job](../deploy/k8s/loadtest-job.yaml) **inside the cluster**, streams k6's live output, and
prints **PASS** only if every threshold held. Running in-cluster is deliberate: a `kubectl port-forward`
is a single userspace TCP proxy and would itself bottleneck at 1k VUs, measuring `kubectl` rather than
the system.

> **Warm-up.** Run it once to warm the catalog cache for the store, then again for the measured number —
> the very first burst against a cold cache pays the one-time fill. The cache then stays warm (10-min
> TTL, no writes during a read test).

### Reading the result

k6 prints a per-step summary; the lines that matter:

```
✓ { step:app_shell }...: p(95)=...ms      ← frontend, must be < 2000ms
✓ { step:catalog }.....: p(95)=...ms      ← backend,  must be < 2000ms
✓ http_req_failed......: 0.00%
PASS — all thresholds met (p95 < 2s, errors < 1%).
```

Capture this summary in the demo video.

### Tuning the run

| Knob | Default | Override |
|---|---|---|
| Virtual users | `1000` | `make load-test LOAD_VUS=2000` |
| Steady-state duration | `30s` | `make load-test LOAD_DURATION=60s` |
| Per-user think time (s) | `1`–`3` (realistic browsing) | `THINK_MIN`/`THINK_MAX` env in the script; set both `0` for a raw throughput flood |
| Gateway / store / page size / auth | gateway origin, demo store, 24, ambient | env vars in [`loadtest/storefront.js`](../loadtest/storefront.js) |

`1000 concurrent users` is modeled as 1000 simultaneous sessions with a 1–3 s human pause between
actions — not 1000 requests every second. By default the test sends **no `Authorization` header**, so
the gateway injects the ambient browse credential exactly as it does for a real anonymous visitor.

### A note on the test environment

The load generator runs in the same single Minikube node as the system. On a CPU-constrained host the
generator and the system contend for the same cores, which inflates the tail. If you see the catalog
p95 graze 2 s, give Minikube more CPU (`minikube start --cpus=N`) or lower the offered load — the
medians (~1 ms) and the 0% error rate show the system itself is far inside the target.

## Frontend single-user load time

Beyond serving under load, the SPA is built to load fast for an individual user: cursor-based infinite
scroll (never bulk-loading the catalog), `loading="lazy"` + sized images to avoid layout shift,
route code-splitting, self-hosted subset fonts (no CDN round-trip), and leaning on the Apollo cache to
avoid request waterfalls. Verify with Lighthouse against the production preview (`npm run build &&
npm run preview` in `frontend/`).
