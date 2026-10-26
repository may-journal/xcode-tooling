---
relatedConfigurations: ['../../.fitnessrc.json']
---

# Components

```mermaid
C4Component
    title Cloud scripts and shared lanes

    Person(owner, "1 Owner")

    System_Boundary(scripts, "Cloud scripts (ci_scripts)") {
        Component(clone, "2 Post clone", "sh", "ci_scripts/ci_post_clone.sh")
        Component(notes, "3 Tester notes", "sh", "ci_scripts/ci_post_xcodebuild.sh")
    }

    System_Boundary(lanes, "Shared lanes (fastlane)") {
        Component(fastfile, "4 Lanes", "fastlane", "fastlane/Fastfile")
    }

    System_Ext(xcodegen, "5 XcodeGen 2.46.0", "Release zip")
    System_Ext(asc, "6 App Store Connect API", "REST")

    Rel(clone, xcodegen, "7")
    Rel(owner, fastfile, "8")
    Rel(fastfile, asc, "9")
    Rel(fastfile, clone, "10")
```

```mermaid
C4Component
    title Legend

    Person(human, "Person", "Owner who runs a lane")
    System_Boundary(area, "Folder boundary") {
        Component(part, "Component", "Tech", "One file of this repo")
    }
    System_Ext(remote, "External", "A download or API it calls")
```

Numbers on nodes and arrows match the callout table.

| #   | Part | Why |
| --- | --- | --- |
| 1   | Same actor as the other views. | Runs a lane on a Mac. |
| 2   | `ci_scripts/ci_post_clone.sh` runs the app's `app_post_clone.sh`, then generates the project when `project.yml` exists. | One XcodeGen version for every run, per [0001](../requirements/0001-post-clone.md). |
| 3   | `ci_scripts/ci_post_xcodebuild.sh` writes the merge's changelog entry after an archive. | Testers read what changed, per [0002](../requirements/0002-tester-notes.md). |
| 4   | `fastlane/Fastfile` holds the shared lanes and their App Store Connect helpers. | The app's lane file imports it at the pinned tag. |
| 5   | The XcodeGen release zip on GitHub. | Checked by its SHA-256 before it runs. |
| 6   | The App Store Connect REST API. | Builds, groups and cloud runs live there. |
| 7   | Downloads XcodeGen unless that version is installed. | A fresh cloud machine has none. |
| 8   | Runs a lane through the app's lane file. | The app's Appfile and Gymfile name it. |
| 9   | Reads and starts runs with the team's key. | No lane uploads. |
| 10  | `archive` runs the app's post-clone hook, which reaches this script. | Local and cloud projects match. |
