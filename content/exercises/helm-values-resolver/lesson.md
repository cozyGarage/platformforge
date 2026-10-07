# Resolve Helm values like Helm

`helm install -f prod.yaml --set image.tag=v2` does not "pick one" source of values. Helm layers them, and the order is the whole contract: chart defaults, then each `-f` file left to right, then each `--set`. A `--set` on the command line beats a file committed to Git, which is why a forgotten flag can silently override a reviewed change.

This lab rebuilds that behaviour in about thirty lines of Python so the precedence rule becomes something you have implemented, not something you remember. The theory is in the reading *Helm Chart Structure and Release Management*.

## Tasks

1. `deep_merge`: maps merge key by key, lists are replaced (Helm does not concatenate lists), inputs are never modified.
2. `set_flag`: `image.tag=v2` becomes a nested map; `3` and `true` are parsed as an integer and a boolean.
3. `resolve`: apply defaults, each `-f` file, then each `--set`, in that order.

Run the whole suite any time with `cd /workspace && python -m unittest discover -s tests`.
