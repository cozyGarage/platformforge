# Keep secrets out of git

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) security topics (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

Tokens in `config.env` become incident reports. Quarantine, ignore, rotate.

## Tasks

1. Move `API_TOKEN` into `/workspace/secrets.env`
2. Keep `APP_ENV=dev` in `app/config.env` (no token left there)
3. Add `secrets.env` to `.gitignore`
4. Write `/workspace/docs/ROTATE.md` about rotating `API_TOKEN`

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps
- Modifications: secret quarantine drill with deterministic validators
