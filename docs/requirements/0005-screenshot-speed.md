# 0005 Screenshot Speed

## Why

An app's owner gets the same images sooner. A run spends its time on the app, not on setup, idle slots or redrawing what stayed the same.

## Measurement

capture runs that boot each device type once

-

capture runs that start

## Requirements

- 0005.1
  - Given Owner has an app with two sizes on one device
    - When a capture runs `screenshots.sh`
      - Then that device boots fresh once and each size runs on a clone
- 0005.2
  - Given Owner has a device that never finishes its first boot
    - When a capture runs `screenshots.sh`
      - Then the run fails with no simulator left behind
- 0005.3
  - Given Owner has a first boot whose shutdown cut a file write short
    - When a capture clones that device
      - Then the clone starts and the size runs
- 0005.4
  - Given Owner has three sizes running two at once
    - When the quick size finishes
      - Then the third size starts while the slow one runs
- 0005.5
  - Given Owner has images already composed from the same captures
    - When the compose tool runs again
      - Then each image stays byte for byte
- 0005.6
  - Given Owner has a size missing one capture
    - When the compose tool runs
      - Then it fails and names that capture
- 0005.7
  - Given Owner has a device that can't be cloned
    - When a capture runs `screenshots.sh`
      - Then it fails after three tries with no simulator left behind
