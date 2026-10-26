---
relatedConfigurations: ['../../.fitnessrc.json']
---

# Containers

```mermaid
C4Container
    title xcode-tooling

    Person(owner, "1 Owner")

    System(release, "2 One tagged release", "Cloud scripts and shared lanes, read in place by each app")

    Container_Ext(shim, "3 App shim", "POSIX sh", "Clones the pinned tag on Xcode Cloud or a Mac")
    Container_Ext(laneFile, "4 App lane file", "fastlane", "Imports the shared lanes at the same tag")
    System_Ext(asc, "5 App Store Connect API", "REST")

    Rel(shim, release, "A")
    Rel(laneFile, release, "B")
    Rel(release, asc, "C")
    Rel(owner, laneFile, "D")
```

```mermaid
C4Container
    title Legend

    Person(actor, "Person", "Owner who runs a lane")
    System(unit, "System", "What this repo releases")
    Container_Ext(outside, "External", "App file or API that calls in or is called")
```

Numbers on nodes and arrows match the callout table. Nothing here builds or deploys on its own. Each app's shim and lane file run the release in place, so this view draws no container here.

| #   | Role | Why |
| --- | --- | --- |
| 1   | Same actor as the system context. | Runs lanes on a Mac. |
| 2   | The cloud scripts and shared lanes at one tag. | Each app runs exactly the release it names. |
| 3   | The shim each app commits, run by its two hook scripts. | Apple runs only committed scripts. |
| 4   | The lane file each app owns, with its own lanes below the import. | fastlane's own way to share lanes. |
| 5   | Builds, tester groups and Xcode Cloud runs. | The team's API key signs each call. |
| A   | Clones the tag the app's version file names, and runs the script of the same name. | The app pins one release. |
| B   | `import_from_git` at the same tag. | A Mac runs what the cloud runs. |
| C   | Reads builds and groups, and starts a run on main. | Nothing here uploads a build. |
| D   | Runs `fastlane archive` or a status lane. | A Mac proves a build without uploading it. |
