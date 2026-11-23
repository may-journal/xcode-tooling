# 0008 Capture Timers

## Why

A Mac that runs a capture is left clean when the capture ends. Each bounded step stops its own timer, so no process outlives the run.

## Measurement

capture runs that leave no timer behind

-

capture runs that start

## Requirements

- 0008.1
  - Given Owner has a capture whose clone finishes well inside its bound
    - When the capture ends
      - Then no timer of that capture still runs
