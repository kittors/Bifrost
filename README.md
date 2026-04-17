# Bifrost

Bifrost is an enterprise private web service access gateway.

The first phase focuses on controlled access to internal `HTTP/HTTPS` services through:

- a Go gateway server
- an Electron desktop client
- a React admin web console
- a shared design system and API contract
- a local Docker-based multi-container test environment

The desktop client is intentionally small: it is a card-like access launcher, not a VPN, not a system proxy, and not an admin console.

## Documentation

Start with [docs/00-overview/README.md](./docs/00-overview/README.md).

The implementation checklist is maintained in [docs/08-roadmap/development-checklist.md](./docs/08-roadmap/development-checklist.md). Completed items must include a completion timestamp.

## Development Status

Implementation has started with Phase 0: repository and engineering skeleton.

