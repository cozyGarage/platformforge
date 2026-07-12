# Follow the packet

An API container cannot reach PostgreSQL. Its connection string points to `localhost`, which means the API container itself—not the database container.

Repair `compose.yaml` using Compose service discovery and make the startup relationship explicit. A dependency is not a readiness check, but it documents topology and controls creation order.

## Diagnostic model

Always identify the network namespace from which a name or address is resolved. Loopback, published host ports, and container service names solve different problems.
