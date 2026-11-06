# 0003 Screenshots

## Why

An app's owner gets its App Store screenshots from one tool, whenever the app changes and before each release. The app's requirement doc names each frame.

## Measurement

capture runs that leave no simulator behind

-

capture runs that start

## Requirements

- 0003.1
  - Given Owner has an app and a booted simulator
    - When a capture runs `screenshots.sh`
      - Then it exits 75 and makes no simulator
- 0003.2
  - Given Owner has an app whose requirement doc names frames with headlines
    - When a capture runs `screenshots.sh`
      - Then each frame reaches the compose tool by the doc's name and headline
- 0003.3
  - Given Owner has an app whose test keeps failing
    - When a capture runs `screenshots.sh`
      - Then it fails after three tries and deletes each simulator it made
- 0003.4
  - Given Owner has a merge to main that changes no watched path
    - When the agent polls `screenshots-watch.sh`
      - Then it captures nothing
- 0003.5
  - Given Owner has a merge to main that changes a watched path
    - When the agent polls `screenshots-watch.sh` twice
      - Then it captures that commit once
- 0003.6
  - Given Owner has an app change and a booted simulator
    - When the agent polls, then polls again once the simulator is gone
      - Then the first poll captures nothing and the second captures the change
- 0003.7
  - Given Owner prepares a release while a simulator is booted
    - When the release runs `screenshots-release.sh`
      - Then it fails, so nothing uploads
- 0003.8
  - Given Owner prepares a release from a clean release commit
    - When the release runs `screenshots-release.sh`
      - Then the images come from that commit
