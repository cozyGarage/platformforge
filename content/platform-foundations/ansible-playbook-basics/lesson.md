# Write a hardening playbook

Inventory names the fleet. A playbook declares the desired package state and when services should bounce.

## Tasks

1. Target `hosts: deploy` with `become: true`
2. Install `nginx` with `ansible.builtin.apt` / `state: present`
3. `notify: Restart nginx` and define that handler with `state: restarted`
4. Record `PLAYBOOK_OK`

Tip code: `MISSING_BECOME`.
