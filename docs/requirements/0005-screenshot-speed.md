# 0005 Screenshot Speed

## Why

An app's owner gets the same images sooner. A run spends its time on the app, not on setup, idle slots or redrawing what stayed the same.

## Measurement

capture runs that boot each device type once

-

capture runs that start

## Requirements

1. Boots a device once
   1. GIVEN Owner has an app with two sizes on one device
      1. WHEN a capture runs `screenshots.sh`
         1. THEN that device boots fresh once
         2. THEN each size runs on a clone
2. Fails a device that never boots
   1. GIVEN Owner has a device that never finishes its first boot
      1. WHEN a capture runs `screenshots.sh`
         1. THEN the run fails with no simulator left behind
3. Clones after a cut write
   1. GIVEN Owner has a first boot whose shutdown cut a file write short
      1. WHEN a capture clones that device
         1. THEN the clone starts
         2. THEN the size runs
4. Starts the third size early
   1. GIVEN Owner has three sizes running two at once
      1. WHEN the quick size finishes
         1. THEN the third size starts while the slow one runs
5. Composes the same bytes
   1. GIVEN Owner has images already composed from the same captures
      1. WHEN the compose tool runs again
         1. THEN each image stays byte for byte
6. Names a missing capture
   1. GIVEN Owner has a size missing one capture
      1. WHEN the compose tool runs
         1. THEN it fails
         2. THEN it names that capture
7. Fails a device that can't be cloned
   1. GIVEN Owner has a device that can't be cloned
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it fails after three tries with no simulator left behind
8. Bounds a clone that never finishes
   1. GIVEN Owner has a device whose clone never finishes
      1. WHEN a capture runs `screenshots.sh`
         1. THEN each try gives up after its bound
         2. THEN the run fails
