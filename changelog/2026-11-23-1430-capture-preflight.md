### 2026.11.23.1430

> An app owner whose capture cannot start reads which tool, file or setting is missing.

- Fix: `screenshots.sh` exits 69 naming `jq` when it is not installed.
- Fix: it exits 66 naming the config path when the file is missing.
- Fix: it exits 78 when the config is not JSON, or names no project, scheme, test, requirement, output or size.
- Fix: each of these stops before the lock and before any simulator, and a bad `jobs` value still exits 64.
