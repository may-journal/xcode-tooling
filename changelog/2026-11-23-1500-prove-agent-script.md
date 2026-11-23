### 2026.11.23.1500

> An app owner installs, updates or removes the screenshot agent with one command, on any Mac.

- Fix: `screenshots-agent.sh` makes `~/Library/LaunchAgents` when a Mac account has none.
- Test: requirement 0009 proves an install loads one job by the config's label.
- Test: it proves a second run replaces the job, and a removal deletes its files.
- Docs: the user journey gains the step that installs the agent.
