# 0008 Capture Timers

## Why

A Mac that runs a capture is left clean when the capture ends. Each bounded step stops its own timer, so no process outlives the run.

## Measurement

capture runs that leave no timer behind

-

capture runs that start

## Requirements

1. Leaves no timer behind
   1. GIVEN Owner has a capture whose clone finishes well inside its bound
      1. WHEN the capture ends
         1. THEN no timer of that capture still runs
