# 0006 Moved Tag

## Why

An app's owner sees fresh screenshots after a tooling release, with no change to the app. The agent keeps the commit its tag named at the last capture, and a moved tag counts as an app change.

## Measurement

tooling releases captured once by each watching agent

-

tooling releases that move a tag an agent follows

## Requirements

- 0006.1
  - Given Owner has a tooling release that moves the app's tag while main stays
    - When the agent polls `screenshots-watch.sh` twice
      - Then it captures once
- 0006.2
  - Given Owner has an agent with no record of the tag
    - When the agent polls
      - Then it captures nothing and records the commit the tag names
- 0006.3
  - Given Owner has a moved tag and a tooling repo the agent can't reach
    - When the agent polls
      - Then it captures nothing and keeps its record
- 0006.4
  - Given Owner has a commit whose capture failed
    - When the agent polls, then polls again after a release moves the tag
      - Then it skips the commit, then tries it again
