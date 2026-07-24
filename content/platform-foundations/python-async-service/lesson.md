# Build an async Python service

Async services keep many network conversations moving without one thread per request. This lab uses Python's standard-library `asyncio` APIs to serve the same platform contract as earlier HTTP labs.

## Tasks

1. Create `/workspace/server.py`
2. Use `asyncio.start_server` on port `8080`
3. Return `ok` for `/health`
4. Return `[{"id":1,"amount":25}]` for `/payments`
5. Set `Content-Type: application/json`
6. Run the server in the background
