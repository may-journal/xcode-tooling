# 0007 Capture Preflight

## Why

An app's owner whose capture cannot start reads what is missing. The run names the tool, the file or the setting, and stops before it takes the lock or makes a simulator.

## Measurement

capture runs that stop naming what is missing

-

capture runs that cannot start

## Requirements

- 0007.1
  - Given Owner has a Mac with no `jq`
    - When a capture runs `screenshots.sh`
      - Then it exits 69 naming `jq` and makes no simulator
- 0007.2
  - Given Owner has an app with no screenshots config
    - When a capture runs `screenshots.sh`
      - Then it exits 66 naming the config path and makes no simulator
- 0007.3
  - Given Owner has an app whose screenshots config is not JSON
    - When a capture runs `screenshots.sh`
      - Then it exits 78 saying so and makes no simulator
- 0007.4
  - Given Owner has an app whose screenshots config names no scheme, then no sizes
    - When a capture runs `screenshots.sh`
      - Then each run exits 78 naming that setting and makes no simulator
