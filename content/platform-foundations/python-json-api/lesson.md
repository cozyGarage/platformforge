# Build a Python JSON API

Platform services need structured responses. Add `/payments` while keeping `/health` green.

## Tasks

1. Keep `/health` → `ok` on `:8080`
2. Add `/payments` returning `[{"id":1,"amount":25}]`
3. Set `Content-Type: application/json`
4. Run the server in the background
