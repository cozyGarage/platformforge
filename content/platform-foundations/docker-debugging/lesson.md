# Repair an unsafe image definition

The supplied Dockerfile uses a moving base tag, copies the full build context, runs as root, and starts a placeholder process. Those choices make releases unpredictable and increase the impact of a compromise.

Rewrite the file as a small, pinned specification. The lab intentionally validates the Dockerfile rather than exposing the host Docker socket to your learner shell.

## Production principle

Container images should be deterministic, minimal, and non-root by default. Build contexts should include only what the application needs.
