# 0007 — Serve pet pictures through an application proxy path

- Status: Accepted
- Date: 2026-06-17

## Context

ADR-0005 stores pictures in MinIO and leaves open whether clients fetch them via
**presigned URLs** or a **proxied path**. The customer frontend (React, separate repo) now forces
the choice. A presigned MinIO URL is signed against a specific host, and in our local Minikube setup
that host is the in-cluster `minio:9000` over plain HTTP — which a browser on the host cannot resolve
and which would be mixed content against an HTTPS single-page app. Making presigned URLs work in a
browser would mean exposing MinIO at a browser-reachable TLS address (extra ingress/host, cert SANs)
and depends on the ingress being routable (on the macOS Docker driver that needs `minikube tunnel`).

## Decision

Serve pet pictures through the API at **`GET /pictures/{objectKey...}`**, which streams the object
from MinIO. `pictureUrl` resolves to this same-origin path; clients never receive a signed URL and the
object-storage bucket and host are never exposed (the opaque object key is the only storage detail in
the path). The route is unauthenticated: pet pictures are public catalog content addressed by an
opaque, unguessable key, so this matches the access model a presigned URL would give while keeping
the bucket private. The store only serves keys under the `pets/` prefix, so the route cannot reach
unrelated objects. Bytes are streamed (never buffered through a GraphQL resolver, per the §7 rule)
with a short `Cache-Control` since object keys are immutable.

## Consequences

- The frontend works over the same TLS and origin as the API (the existing `kubectl port-forward` is
  enough) — no second exposed service, no extra certificate, no CORS, and no `http`/`https`
  mixed-content problem.
- Storage topology is encapsulated: clients see `/pictures/<key>`, not the bucket, host, or a
  signature. The picture path can switch storage backends without a client change.
- Image bytes flow through the API tier rather than direct-to-MinIO. At this challenge's scale that is
  immaterial, and browser/proxy caching (`Cache-Control`) absorbs repeat reads; if it ever mattered,
  reintroducing presigned URLs is a contained change behind the same `PictureStore` interface.
- `PictureStore.PresignedURL` is removed in favour of `Get`; no dead code is left behind.
