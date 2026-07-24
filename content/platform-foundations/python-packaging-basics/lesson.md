# Package a Python module

Services become libraries. A tiny `pyproject.toml` plus tests is the packaging baseline platform teams expect.

## Tasks

1. Create `payments_fees/__init__.py` with `fee_cents(amount) = amount * 2 // 100`
2. Write `pyproject.toml` name `payments-fees`, version `0.1.0`
3. Add `tests/test_fees.py` asserting `fee_cents(1000) == 20`
4. Run tests and save output to `/workspace/test-output.txt`
