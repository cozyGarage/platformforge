# Close basic PCI-DSS gaps

PCI-DSS expects cardholder data to stay protected in transit and at rest — and preferably never stored. This lab hardens a reckless checkout config.

## Starting state

`config/checkout.env` serves HTTP, logs PANs, and stores them in `data/cards.txt`.

## Tasks

1. Switch `listen_url` to `https://...`
2. Set `log_pan=false` and `store_pan=false`
3. Delete `/workspace/data/cards.txt`
4. Write `/workspace/evidence/pci-checklist.md` noting encryption in transit and no PAN storage
