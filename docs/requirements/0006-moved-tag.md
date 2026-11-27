# 0006 Moved Tag

## Why

An app's owner sees fresh screenshots after a tooling release, with no change to the app. The agent keeps the commit its tag named at the last capture, and a moved tag counts as an app change.

## Measurement

tooling releases captured once by each watching agent

-

tooling releases that move a tag an agent follows

## Requirements

1. Captures a moved tag once
   1. GIVEN Owner has a tooling release that moves the app's tag while main stays
      1. WHEN the agent polls `screenshots-watch.sh` twice
         1. THEN it captures once
2. Records a first tag
   1. GIVEN Owner has an agent with no record of the tag
      1. WHEN the agent polls
         1. THEN it captures nothing
         2. THEN it records the commit the tag names
3. Keeps its record when the tooling is out of reach
   1. GIVEN Owner has a moved tag
   2. GIVEN a tooling repo the agent can't reach
      1. WHEN the agent polls
         1. THEN it captures nothing
         2. THEN it keeps its record
4. Retries a failed commit after a release
   1. GIVEN Owner has a commit whose capture failed
      1. WHEN the agent polls
         1. THEN it skips the commit
      2. WHEN the agent polls again after a release moves the tag
         1. THEN it tries the commit again
