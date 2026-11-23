# 0009 Screenshot Agent

## Why

An app's owner installs, updates or removes the screenshot agent with one command, on any Mac. The job carries the config's label, so each app has its own.

## Measurement

agent installs that leave one loaded job

-

agent installs that run

## Requirements

- 0009.1
  - Given Owner has an app's main checkout
    - When Owner runs `screenshots-agent.sh`
      - Then launchd loads a job by the config's label that polls every five minutes from that checkout
- 0009.2
  - Given Owner has a Mac account with no `LaunchAgents` folder
    - When Owner runs `screenshots-agent.sh`
      - Then it makes the folder and loads the job
- 0009.3
  - Given Owner has the agent installed
    - When Owner runs `screenshots-agent.sh` again
      - Then it unloads the old job first and leaves one loaded job
- 0009.4
  - Given Owner has the agent installed
    - When Owner runs `screenshots-agent.sh` with `SCREENSHOTS_REMOVE` set
      - Then it unloads the job and deletes the job file and its copy of the watch script
