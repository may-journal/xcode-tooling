# 0004 Screenshot Runs

## Why

An app's owner waits for one size, not all of them in turn, and one bad size can't pass or leak a simulator.

## Measurement

capture runs whose sizes overlap and leave no simulator behind

-

capture runs with more than one size

## Requirements

- 0004.1
  - Given Owner has an app with two sizes
    - When the sizes start
      - Then both run at the same time
- 0004.2
  - Given Owner has an app with one size that keeps failing and one that passes
    - When the failing size runs out of tries
      - Then the run fails with no simulator left behind
- 0004.3
  - Given Owner has a booted simulator left by a dead capture
    - When the next capture starts
      - Then it deletes that simulator and goes on
- 0004.4
  - Given Owner has an acceptance that quotes words but names no frame
    - When the capture reads the doc
      - Then that acceptance is no frame
- 0004.5
  - Given Owner has a merge that bumps the tooling version
    - When the agent polls
      - Then that counts as an app change
- 0004.6
  - Given Owner has an untracked source file beside a release commit
    - When the release starts its capture
      - Then it refuses, naming the commit it wants
- 0004.7
  - Given Owner has a capture whose publish fails
    - When the agent polls that commit again, then polls a later app change
      - Then it skips the commit, then captures the later change
