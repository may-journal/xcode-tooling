# 0004 Screenshot Runs

## Why

An app's owner can run sizes side by side, and one bad size can't pass or leak a simulator or clone.

## Measurement

capture runs whose sizes overlap and leave no simulator behind

-

capture runs with more than one size

## Requirements

1. Runs sizes together
   1. GIVEN Owner has an app with two sizes set to run two at once
      1. WHEN the sizes start
         1. THEN both run at the same time
2. Fails a size that runs out of tries
   1. GIVEN Owner has an app with one size that keeps failing and one that passes
      1. WHEN the failing size runs out of tries
         1. THEN the run fails with no simulator left behind
3. Deletes a dead capture's simulator
   1. GIVEN Owner has a booted simulator left by a dead capture
      1. WHEN the next capture starts
         1. THEN it deletes that simulator and goes on
4. Ignores a quote with no frame
   1. GIVEN Owner has an acceptance that quotes words but names no frame
      1. WHEN the capture reads the doc
         1. THEN that acceptance is no frame
5. Counts a tooling bump
   1. GIVEN Owner has a merge that bumps the tooling version
      1. WHEN the agent polls
         1. THEN that counts as an app change
6. Refuses an untracked file
   1. GIVEN Owner has an untracked source file beside a release commit
      1. WHEN the release starts its capture
         1. THEN it refuses naming the commit it wants
7. Moves past a failed publish
   1. GIVEN Owner has a capture whose publish fails
      1. WHEN the agent polls that commit again
         1. THEN it skips the commit
      2. WHEN the agent polls a later app change
         1. THEN it captures the later change
8. Rejects a bad job count
   1. GIVEN Owner sets jobs to zero or a word
      1. WHEN a capture runs `screenshots.sh`
         1. THEN it stops before it makes a simulator
