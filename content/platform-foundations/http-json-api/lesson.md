# Build a tiny JSON API in Go

Platform services speak JSON. This lab adds a `/payments` read endpoint while keeping `/health` for probes.

## Tasks

1. Keep `/health` → `ok` on `:8080`
2. Add `/payments` returning `[{"id":1,"amount":25}]`
3. Set `Content-Type: application/json`
4. Run the server in the background

## Example handler

```go
http.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.Write([]byte(`[{"id":1,"amount":25}]`))
})
```
