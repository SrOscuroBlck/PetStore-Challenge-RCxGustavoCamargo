# Architecture Decision Records

Each record captures one significant decision: its context, the choice made, and the consequences. Format is lightweight [MADR](https://adr.github.io/madr/). Records are immutable once accepted; a reversed decision gets a new record that supersedes the old one.

| # | Decision | Status |
|---|---|---|
| [0001](0001-backend-stack.md) | Go for the backend, and the overall stack | Accepted |
| [0002](0002-graphql-over-grpc.md) | GraphQL as the protocol; no gRPC | Accepted |
| [0003](0003-sqlc-and-atlas.md) | sqlc for data access, Atlas for migrations | Accepted |
| [0004](0004-race-condition-strategy.md) | Database-enforced race-condition strategy | Accepted |
| [0005](0005-image-storage-minio.md) | MinIO object storage for pet pictures | Accepted |
