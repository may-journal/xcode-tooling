---
relatedConfigurations: ['.fitnessrc.json', 'release-please-config.json']
---

# xcode-tooling

Shared Xcode Cloud scripts and fastlane lanes for the may-journal Swift apps, released by tag. Each app pins one release, and a bump pull request moves the pin when a new tag lands.

## What it holds

| Path | Runs | Does |
| --- | --- | --- |
| `ci_scripts/ci_post_clone.sh` | After Xcode Cloud clones the app | Runs the app's `ci_scripts/app_post_clone.sh`, then generates the project with XcodeGen 2.46.0, checked by its SHA-256 |
| `ci_scripts/ci_post_xcodebuild.sh` | After each xcodebuild action | After an archive, writes the merge's changelog entry to `TestFlight/WhatToTest.en-US.txt` |
| `fastlane/Fastfile` | On a Mac, through the app's lane file | The `status`, `cloud_status`, `cloud_build` and `archive` lanes. None uploads a build. |

Both scripts run from the app's root.

## How an app uses a release

Xcode Cloud runs only scripts an app commits, so each app holds a small shim:

| App file | Holds |
| --- | --- |
| `ci_scripts/xcode-tooling.version` | One line: the tag the app runs, such as `v1.0.0` |
| `ci_scripts/xcode-tooling.sh` | The shim. It clones that tag with `git clone --depth 1 --branch` and runs the script named like the link it was called by. A missing tag fails the build, naming it. |
| `ci_scripts/ci_post_clone.sh`, `ci_scripts/ci_post_xcodebuild.sh` | Links to the shim |
| `fastlane/Fastfile` | `import_from_git` of this repo's `fastlane/Fastfile` at the same tag, then the app's own lanes |

Org tooling seeds these files in each app, and Renovate bumps the version file when a release lands. This repo is public because an Xcode Cloud script has no token to clone a private repo.

## Releases

Release Please keeps a release pull request open from the conventional commits on main. Merging it tags `vX.Y.Z` and moves `vX`. A `fix` makes a patch, a `feat` a minor release, and a breaking change a major one.

## Development

Run `go test -count=1 ./requirements`. Each test runs a script with `sh` from a temporary app's root, as the shim does. The [requirements](docs/requirements) say what each script promises, and the [architecture docs](docs/architecture) show how the parts fit.

Use the shared hooks with `git config --local core.hooksPath .fitness-hooks`; [fitness setup](.github/fitness-setup.md) installs them.

## License

[MIT](LICENSE)
