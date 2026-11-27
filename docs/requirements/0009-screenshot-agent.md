# 0009 Screenshot Agent

## Why

An app's owner installs, updates or removes the screenshot agent with one command, on any Mac. The job carries the config's label, so each app has its own.

## Measurement

agent installs that leave one loaded job

-

agent installs that run

## Requirements

1. Loads the polling job
   1. GIVEN Owner has an app's main checkout
      1. WHEN Owner runs `screenshots-agent.sh`
         1. THEN launchd loads a job by the config's label that polls every five minutes from that checkout
2. Makes the LaunchAgents folder
   1. GIVEN Owner has a Mac account with no `LaunchAgents` folder
      1. WHEN Owner runs `screenshots-agent.sh`
         1. THEN it makes the folder
         2. THEN it loads the job
3. Reloads one job
   1. GIVEN Owner has the agent installed
      1. WHEN Owner runs `screenshots-agent.sh` again
         1. THEN it unloads the old job first
         2. THEN it leaves one loaded job
4. Removes the agent
   1. GIVEN Owner has the agent installed
      1. WHEN Owner runs `screenshots-agent.sh` with `SCREENSHOTS_REMOVE` set
         1. THEN it unloads the job
         2. THEN it deletes the job file
         3. THEN it deletes its copy of the watch script
