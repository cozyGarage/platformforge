# Serve a Python health endpoint

Same platform contract as the Go lab — probes need `/health` — implemented with Python's standard library.

## Tasks

1. Write `/workspace/server.py` using `http.server`
2. Handle `/health` with body `ok`
3. Listen on port `8080`
4. Run in the background and verify with `wget`
