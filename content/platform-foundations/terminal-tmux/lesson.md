# Survive with tmux sessions

SSH drops. Tmux keeps your incident response windows alive. Platform engineers standardize prefix keys and named windows so muscle memory works on every bastion.

## Tasks

1. Write `/workspace/.tmux.conf` with `prefix C-a` and `mouse on`
2. Create detached session `platform` with windows `editor` and `logs`
3. Record `tmux ls` output to `/workspace/session.txt`

## Useful commands

```sh
tmux new-session -d -s platform -n editor
tmux new-window -t platform -n logs
tmux ls
```
