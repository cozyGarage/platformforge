# GitHub Pages for PlatformForge

PlatformForge's lab runtime needs Docker and `platformforge serve`, so the full product cannot host on GitHub Pages.

What we ship on Pages instead:

- brand landing page
- quick-start clone instructions
- generated DevOps Engineer Path curriculum from `content/paths/devops-engineer.yaml` + lab manifests

## Local build

```bash
make site
# open site-dist/index.html or serve site-dist with any static server
```

## Deploy

Workflow: `.github/workflows/deploy-pages.yml`

- triggers on push to `main` and `workflow_dispatch`
- builds with `scripts/build-site.py`
- force-publishes an orphan `gh-pages` branch (CDN-friendly)
- uploads `site-dist/` and deploys with `actions/deploy-pages` when Pages is enabled

### Live now (no Settings click required)

https://cdn.jsdelivr.net/gh/cozyGarage/platformforge@gh-pages/index.html

### Official GitHub Pages host

Enable once:

1. https://github.com/cozyGarage/platformforge/settings/pages
2. **Build and deployment → Source: GitHub Actions**

Then re-run **Deploy GitHub Pages**. Expected URL: `https://cozygarage.github.io/platformforge/`
Pages enabled; redeploy 2026-07-25T09:37Z
