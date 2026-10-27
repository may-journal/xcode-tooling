# System Context

```mermaid
C4Context
    title xcode-tooling

    Person(owner, "1 Owner")
    Person(agent, "2 AI agent")
    System(tooling, "3 xcode-tooling", "Cloud scripts and shared lanes, released by tag")
    System_Ext(apps, "4 Swift apps", "Each pins one tag and runs a shim")
    System_Ext(cloud, "5 Xcode Cloud", "Clones an app, runs its scripts, archives and uploads")
    System_Ext(release, "6 Release Please", "Tags each release from conventional commits")

    Rel(owner, tooling, "A")
    Rel(agent, tooling, "B")
    Rel(release, tooling, "C")
    Rel(apps, tooling, "D")
    Rel(cloud, apps, "E")

    UpdateElementStyle(owner, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(agent, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(tooling, $bgColor="#1d4ed8", $borderColor="#1e3a8a", $fontColor="#ffffff")
    UpdateElementStyle(apps, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
    UpdateElementStyle(cloud, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
    UpdateElementStyle(release, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
```

```mermaid
C4Context
    title Legend

    Person(who, "Person", "Owner or agent who changes the tooling")
    System(sys, "System", "The tooling this repo releases")
    System_Ext(ext, "External", "An app or service that uses it")

    UpdateElementStyle(who, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(sys, $bgColor="#1d4ed8", $borderColor="#1e3a8a", $fontColor="#ffffff")
    UpdateElementStyle(ext, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
```

Numbers on nodes and letters on arrows match the callout table.

| #   | Description | Why |
| --- | --- | --- |
| 1   | Approves pull requests and the release they make. | Every change ships through a reviewed pull request. |
| 2   | Opens those pull requests. | Agents cannot approve their own work. |
| 3   | This repo: `ci_scripts/` and `fastlane/Fastfile`. | One source for every app's cloud scripts and lanes. |
| 4   | Each Swift app in the org, with its shim and version file. | An app names the exact release it runs. |
| 5   | Apple's build service, each app's only uploader. | It runs only scripts the app commits, so the shim fetches the rest. |
| 6   | The release workflow on main. | Versions come from commit types, with nothing new to run. |
| A   | Merges a change, then the release pull request. | The merge is the approval. |
| B   | Same path as the owner, ending at the owner's approval. | One flow for people and agents. |
| C   | Tags `vX.Y.Z` and moves `vX`. | A tag is what an app pins. |
| D   | The shim clones the pinned tag; the lane file imports the lanes from it. | A Mac and the cloud run the same release. |
| E   | Runs the app's `ci_post_clone.sh` and `ci_post_xcodebuild.sh`. | Both hand their name to the app's shim. |
