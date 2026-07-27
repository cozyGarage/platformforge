# Build an Ansible inventory

Platform deploys start with knowing *which* hosts get which roles. A static inventory keeps bastion and payments nodes addressable as groups.

## Ticket hosts

- `bastion` → `10.0.0.10`
- `payments-01` → `10.0.1.11`
- `payments-02` → `10.0.1.12`

## Tasks

1. Write `/workspace/ansible/inventory.ini`
2. Create `[bastion]`, `[payments]`, and `[deploy:children]`
3. Record `INVENTORY_OK`

Tip code: `MISSING_CHILDREN`.
