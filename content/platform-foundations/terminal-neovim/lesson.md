# Edit faster with Neovim basics

You do not need a full IDE on a bastion host. A tiny `init.lua` plus confident edits get incident configs fixed quickly.

## Tasks

1. Write `/workspace/.config/nvim/init.lua` with line numbers and `ignorecase`
2. Fix `/workspace/etc/systemd/system/api.service`:
   - `Description=Platform API`
   - `ExecStart=/usr/bin/true`
3. Copy the ExecStart line into `/workspace/edit-proof.txt`

## Minimal init.lua

```lua
vim.opt.number = true
vim.opt.ignorecase = true
```
