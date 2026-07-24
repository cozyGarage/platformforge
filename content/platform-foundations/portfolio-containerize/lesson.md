# Portfolio — containerize the API

Recruiters and hiring managers look for a Dockerfile that is pinned, non-root, and boringly correct.

## Starting state

`main.go` and `go.mod` for the payments API are already in `/workspace`.

## Tasks

1. Author `/workspace/Dockerfile` using `golang:1.22-alpine` (no `latest`)
2. Run as non-root (`USER app` or `USER 1000`)
3. Expose `8080` and start the API via `CMD`/`ENTRYPOINT`
4. Add `/workspace/.dockerignore` with `.git`
