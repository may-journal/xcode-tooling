# xcode-tooling

Shared Xcode Cloud scripts and fastlane lanes for the may-journal Swift apps, released by tag. Each app pins one release, and a bump pull request moves the pin when a new tag lands.

## What it holds

| Path | Runs | Does |
| --- | --- | --- |
| `ci_scripts/ci_post_clone.sh` | After Xcode Cloud clones the app | Runs the app's `ci_scripts/app_post_clone.sh`, then generates the project with XcodeGen 2.46.0, checked by its SHA-256 |
| `ci_scripts/ci_post_xcodebuild.sh` | After each xcodebuild action | After an archive, writes the merge's changelog entry to `TestFlight/WhatToTest.en-US.txt` |
| `ci_scripts/screenshots.sh` | On a Mac, by `screenshots` lane | Captures each store frame at each size, and deletes every simulator it made |
| `ci_scripts/screenshots-compose.swift` | By `screenshots.sh` | Sets each headline above its screen at the exact store size |
| `ci_scripts/screenshots-agent.sh` | On a Mac, from the app's main checkout | Installs a launchd agent that polls `main`; `SCREENSHOTS_REMOVE=1` takes it away |
| `ci_scripts/screenshots-watch.sh` | By that agent | Captures once per commit that changes the app, and refreshes one pull request when images change |
| `ci_scripts/screenshots-release.sh` | By `capture_for_release` | Captures from the release commit, and fails rather than let a release carry stale images |
| `fastlane/Fastfile` | On a Mac, through the app's lane file | The `status`, `cloud_status`, `cloud_build`, `archive`, `screenshots` and `capture_for_release` lanes. None uploads a build. |

Both scripts run from the app's root.

## How an app uses a release

Xcode Cloud runs only scripts an app commits, so each app holds a small shim:

| App file | Holds |
| --- | --- |
| `ci_scripts/xcode-tooling.version` | One line: the tag the app runs, such as `v1.0.0` |
| `ci_scripts/xcode-tooling.sh` | The shim. It clones that tag with `git clone --depth 1 --branch` and runs the hook script it is given. A missing tag fails the build, naming it. |
| `ci_scripts/ci_post_clone.sh`, `ci_scripts/ci_post_xcodebuild.sh` | Two-line hooks that run the shim with their own name |
| `fastlane/Fastfile` | `import_from_git` of this repo's `fastlane/Fastfile` at the same tag, then the app's own lanes |

Org tooling seeds these files in each app, and Renovate bumps the version file when a release lands. This repo is public because an Xcode Cloud script has no token to clone a private repo.

## App Store screenshots

An app runs `ci_scripts/xcode-tooling.sh screenshots.sh` or `fastlane ios screenshots`. Its own requirement doc holds the scenes: one acceptance per frame, whose When reads `When I look at the <name>` and whose Then quotes the headline. A test per acceptance reaches the frame, leaves a `.wants` file named for the acceptance number in `SCREENSHOT_SHUTTER`, and waits for `.taken`. `SCREENSHOT_WINDOW`, such as `669x951`, asks for a window smaller than the simulator.

`fastlane/screenshots.json` names the rest:

| Key | Holds |
| --- | --- |
| `label` | The launchd label, also the name of the lock and the logs |
| `project`, `scheme`, `bundle`, `test` | What to build and which test suite to run |
| `grant` | Privacy services to allow, so no system prompt covers a frame |
| `requirement`, `output`, `background` | The scenes doc, the image folder, and the color behind each headline |
| `jobs` | How many sizes run at once, two by default |
| `watch` | Paths that count as an app change, beside the requirement doc |
| `repo`, `branch`, `runsIssue`, `agent`, `trailer` | Where a watched capture publishes and notes each result |
| `sizes` | Each `folder`, simulator `device`, final `width` and `height`, and optional `window`, `capture`, `screen` and `datedStatusBar` |

A capture exits 75 while any simulator is booted, since one agent drives a simulator at a time. Sizes pick the newest runtime that supports their device. A size can shoot a window on a larger simulator: `capture` is that simulator's pixels, and `screen` is the corner the window fills. Two runs always differ by a few pixels, so a change under 0.02% of an image keeps the old file.

## How the sizes run

It builds once, then runs the sizes at once, each on its own simulator with its own shutter and raw folder. `jobs` in the config, or `SCREENSHOTS_JOBS`, sets how many run together, two by default. On an 18-core Mac, three or four at once ran no faster than one at a time, and test timing broke. A failed size fails the run after three tries of its own, and every simulator is deleted however the run ends.

## When a capture runs

The agent polls `origin/main` every five minutes and exits at once unless a commit changed a watched path. It records the last captured commit, so each change is captured once and a skipped run retries on the next poll. A failed commit isn't retried, and the next change picks it up. A release lane calls `capture_for_release` first; a busy simulator or an unclean tree stops the release.

## Releases

Release Please keeps a release pull request open from the conventional commits on main. Merging it tags `vX.Y.Z` and moves `vX`. A `fix` makes a patch, a `feat` a minor release, and a breaking change a major one.

## Development

Run `go test -count=1 ./requirements`. Each test runs a script with `sh` from a temporary app's root, as the shim does. The [requirements](docs/requirements) say what each script promises, and the [architecture docs](docs/architecture) show how the parts fit.

Use the shared hooks with `git config --local core.hooksPath .fitness-hooks`; [fitness setup](.github/fitness-setup.md) installs them.

## License

[MIT](LICENSE)
