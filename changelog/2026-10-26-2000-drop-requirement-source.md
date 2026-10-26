---
relatedConfigurations: ['../release-please-config.json']
---

### 2026.10.26.2000

- Docs: requirement docs no longer end their Measurement with a `Source:` line, which went stale whenever code moved
- Docs: the test named for each acceptance is what leads from a requirement to its code
- Docs: `fitness requirements --fix` deletes any `Source:` line a doc still carries
