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
- uploads `site-dist/` via `actions/upload-pages-artifact`
- deploys with `actions/deploy-pages`

Enable once in the GitHub UI if needed:

1. Repo **Settings → Pages**
2. **Build and deployment → Source: GitHub Actions**

Expected URL: `https://cozygarage.github.io/platformforge/`
