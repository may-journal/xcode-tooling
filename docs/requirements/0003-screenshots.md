# 0003 Screenshots

## Why

An app's owner gets its App Store screenshots from one tool. The app's requirement doc names each frame, and a capture never leaves a simulator behind.

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
      - Then it fails after three tries and deletes every simulator it made
