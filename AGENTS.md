---
relatedConfigurations: ['.fitnessrc.json']
---

# xcode-tooling

Shared Xcode Cloud scripts and fastlane lanes, released by tag. Every Swift app in the org pins one release. See [README.md](README.md).

## Rules

Keep plans in GitHub issues and comments. Never commit plan documents, screenshots, or review artifacts.

- Work from a plan. Non-trivial work follows a `Plan`-labeled [GitHub Issue](https://github.com/may-journal/xcode-tooling/issues?q=label%3APlan).
- This repo is public, so never name an app, a person, a secret or a private path.
- The `flagWords` in [cspell.json](cspell.json) fail any file that names an app.
- A change reaches every app, so keep it compatible or mark it breaking. Release Please sets the version from the commit type.
- Scripts run with `sh` from the app's root, on a fresh Xcode Cloud machine. Use POSIX shell and only the tools macOS ships.
- No lane uploads a build. Xcode Cloud is each app's only uploader.
- Pair each change with a `changelog/` entry, and a requirement change with its test. Run `go test -count=1 ./requirements` before a push.
