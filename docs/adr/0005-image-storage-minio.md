# 0005 — MinIO object storage for pet pictures

- Status: Accepted
- Date: 2026-06-13

## Context

Each pet has a picture. The options are storing the bytes in Postgres (`bytea`), on a filesystem volume, or in an object store. The challenge forbids externally-hosted services, so whatever we pick must run locally.

## Decision

Store pictures in self-hosted **MinIO** (S3-compatible). Postgres stores only the object key; images are served via presigned or proxied URLs and never streamed through GraphQL resolvers.

## Consequences

- Large blobs stay out of the database — smaller DB, faster backups, and better read performance under the 1k-concurrent-user target.
- Adds one self-hosted component to the local stack (docker-compose + Minikube), with no external dependency.
- Object storage sits behind a port/interface, so it stays swappable.
