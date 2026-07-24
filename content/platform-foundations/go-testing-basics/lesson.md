# Prove behavior with Go tests

Before you wire HTTP handlers to databases, lock the math behind fees with `go test`.

## Tasks

1. Implement `FeeCents(amount int) int` as `amount * 2 / 100` in `/workspace/fee.go`
2. Add `/workspace/fee_test.go` asserting `FeeCents(1000) == 20`
3. Run `go test ./...` and save output to `/workspace/test-output.txt`

## Example

```go
func FeeCents(amount int) int {
  return amount * 2 / 100
}
```
