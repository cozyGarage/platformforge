# Portfolio — payments API service

Portfolio projects need a real artifact, not just lab checkmarks. This is the application core you will containerize and deploy next.

## Tasks

1. Implement `/workspace/main.go` with:
   - `GET /health` → `ok`
   - `GET /payments` → `[{"id":1,"amount":25}]` + `Content-Type: application/json`
2. Write `/workspace/README.md` with **Overview**, **Run locally**, and **Endpoints**
3. Run the server on `:8080` in the background
