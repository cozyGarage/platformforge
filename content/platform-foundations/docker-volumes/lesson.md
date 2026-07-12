# Persist data with volumes

Ephemeral containers lose data on restart. Named volumes are the Compose pattern Boot.dev covers in **Learn Docker — Storage**.

## Task

Edit `/workspace/compose.yaml` to:

1. Declare a named volume `webdata` at the top level
2. Mount `webdata` at `/usr/share/nginx/html` for the `api` service

## Example snippet

```yaml
services:
  api:
    volumes:
      - webdata:/usr/share/nginx/html
volumes:
  webdata:
```
