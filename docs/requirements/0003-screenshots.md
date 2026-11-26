# 0003 Screenshots

## Why

An app's owner gets its App Store screenshots from one tool, whenever the app changes and before each release. The app's requirement doc names each frame, and each try runs on its own clone.

## Measurement

capture runs that leave no simulator behind

-

capture runs that start

## Requirements

1. Refuses a booted simulator
   1. GIVEN Owner has an app and a booted simulator
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it exits 75 and makes no simulator
2. Names each frame
   1. GIVEN Owner has an app whose requirement doc names frames with headlines
      1. WHEN a capture runs `screenshots.sh`
         1. THEN each frame reaches the compose tool by the doc's name and headline
3. Gives up after three tries
   1. GIVEN Owner has an app whose test keeps failing
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it fails after three tries and deletes each simulator it made
4. Skips an unwatched merge
   1. GIVEN Owner has a merge to main that changes no watched path
      1. WHEN the agent polls `screenshots-watch.sh`
         1. THEN it captures nothing
5. Captures a watched merge once
   1. GIVEN Owner has a merge to main that changes a watched path
      1. WHEN the agent polls `screenshots-watch.sh` twice
         1. THEN it captures that commit once
6. Waits for a free simulator
   1. GIVEN Owner has an app change and a booted simulator
      1. WHEN the agent polls
         1. THEN the first poll captures nothing
      2. WHEN the agent polls again once the simulator is gone
         1. THEN the second captures the change
7. Stops a release with a simulator
   1. GIVEN Owner prepares a release while a simulator is booted
      1. WHEN the release runs `screenshots-release.sh`
         1. THEN it fails so nothing uploads
8. Captures the release commit
   1. GIVEN Owner prepares a release from a clean release commit
      1. WHEN the release runs `screenshots-release.sh`
         1. THEN the images come from that commit
