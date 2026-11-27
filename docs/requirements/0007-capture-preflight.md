# 0007 Capture Preflight

## Why

An app's owner whose capture cannot start reads what is missing. The run names the tool, the file or the setting, and stops before it takes the lock or makes a simulator.

## Measurement

capture runs that stop naming what is missing

-

capture runs that cannot start

## Requirements

1. Names a missing jq
   1. GIVEN Owner has a Mac with no `jq`
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it exits 69 naming `jq`
         2. THEN it makes no simulator
2. Names a missing config
   1. GIVEN Owner has an app with no screenshots config
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it exits 66 naming the config path
         2. THEN it makes no simulator
3. Rejects a config that is not JSON
   1. GIVEN Owner has an app whose screenshots config is not JSON
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it exits 78 saying so
         2. THEN it makes no simulator
4. Names a missing setting
   1. GIVEN Owner has an app whose screenshots config names no scheme
   2. GIVEN no sizes
      1. WHEN a capture runs `screenshots.sh`
         1. THEN each run exits 78 naming that setting
         2. THEN each run makes no simulator
