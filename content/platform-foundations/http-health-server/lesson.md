# Serve a Go health endpoint

Load balancers and Kubernetes probes need a cheap liveness signal. A `/health` handler is the smallest useful HTTP service.

## Tasks

1. Write `/workspace/main.go` using `net/http`
2. Handle `/health` with response body `ok`
3. Listen on `:8080`
4. Run the server in the background and verify with `wget`

## Skeleton

```go
package main

import (
  "fmt"
  "net/http"
)

func main() {
  http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "ok")
  })
  http.ListenAndServe(":8080", nil)
}
```
