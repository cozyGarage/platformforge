# Mask PII before export

FinTech platforms cannot hand vendors raw PANs, SSNs, or personal emails. Masking (or tokenization) is a compliance control you apply before data leaves the trust boundary.

## Starting state

`/workspace/export/customers.raw.csv` contains names, emails, full PANs, and SSNs.

## Tasks

1. Write `/workspace/export/customers.masked.csv`
2. Keep `id` and `name`
3. Replace every email with `[REDACTED]`
4. Mask PAN to last 4 digits (e.g. `***********1111`)
5. Replace SSN with `***-**-****`

## Example row

```csv
1,Ada Lovelace,[REDACTED],***********1111,***-**-****
```
