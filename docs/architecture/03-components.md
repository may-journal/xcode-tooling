---
relatedConfigurations: ['../../.fitnessrc.json']
---

# Components

```mermaid
C4Component
    title Cloud scripts and shared lanes

    Person(owner, "1 Owner")
    System_Ext(asc, "8 App Store Connect API", "REST")
    System_Ext(xcodegen, "7 XcodeGen 2.46.0", "Release zip")

    Container_Boundary(scripts, "2 Cloud scripts (ci_scripts)") {
        Component(clone, "3 Post clone", "sh", "ci_scripts/ci_post_clone.sh")
        Component(notes, "4 Tester notes", "sh", "ci_scripts/ci_post_xcodebuild.sh")
    }

    Container_Boundary(lanes, "5 Shared lanes (fastlane)") {
        Component(fastfile, "6 Lanes", "fastlane", "fastlane/Fastfile")
    }

    Rel(clone, xcodegen, "A")
    Rel(owner, fastfile, "B")
    Rel(fastfile, asc, "C")
    Rel(fastfile, clone, "D")

    UpdateElementStyle(owner, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(clone, $bgColor="#b45309", $borderColor="#92400e", $fontColor="#ffffff")
    UpdateElementStyle(notes, $bgColor="#b45309", $borderColor="#92400e", $fontColor="#ffffff")
    UpdateElementStyle(fastfile, $bgColor="#b45309", $borderColor="#92400e", $fontColor="#ffffff")
    UpdateElementStyle(xcodegen, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
    UpdateElementStyle(asc, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
```

```mermaid
C4Component
    title Legend

    Person(human, "Person", "Owner who runs a lane")
    Container_Boundary(area, "Container boundary") {
        Component(part, "Component", "Tech", "One file of this repo")
    }
    System_Ext(remote, "External", "A download or API it calls")

    UpdateElementStyle(human, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(part, $bgColor="#b45309", $borderColor="#92400e", $fontColor="#ffffff")
    UpdateElementStyle(remote, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
```

Numbers on nodes and letters on arrows match the callout table.

| #   | Description | Why |
| --- | --- | --- |
| 1   | Same actor as the other views. | Runs a lane on a Mac. |
| 2   | The cloud scripts, from `ci_scripts`. | Xcode Cloud runs them inside each app's build. |
| 3   | `ci_scripts/ci_post_clone.sh` runs the app's `app_post_clone.sh`, then generates the project when `project.yml` exists. | One XcodeGen version for every run, per [0001](../requirements/0001-post-clone.md). |
| 4   | `ci_scripts/ci_post_xcodebuild.sh` writes the merge's changelog entry after an archive. | Testers read what changed, per [0002](../requirements/0002-tester-notes.md). |
| 5   | The shared lanes, from `fastlane`. | Each app's fastlane imports them. |
| 6   | `fastlane/Fastfile` holds the shared lanes and their App Store Connect helpers. | The app's lane file imports it at the pinned tag. |
| 7   | The XcodeGen release zip on GitHub. | Checked by its SHA-256 before it runs. |
| 8   | The App Store Connect REST API. | Builds, groups and cloud runs live there. |
| A   | Downloads XcodeGen unless that version is installed. | A fresh cloud machine has none. |
| B   | Runs a lane through the app's lane file. | The app's Appfile and Gymfile name it. |
| C   | Reads and starts runs with the team's key. | No lane uploads. |
| D   | `archive` runs the app's post-clone hook, which reaches this script. | Local and cloud projects match. |
