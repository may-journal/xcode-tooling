# 0004 Screenshot Sizes

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
