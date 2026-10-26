---
relatedConfigurations: ['../../.fitnessrc.json']
---

# Containers

```mermaid
C4Container
    title xcode-tooling

    Person(owner, "1 Owner")
    System_Ext(fastlaneRun, "2 An app's fastlane run", "Its lane file imports the shared lanes at the pinned tag")
    System_Ext(cloudRun, "3 An app's Xcode Cloud run", "Its shim clones the pinned tag and runs the cloud scripts")
    System_Ext(asc, "4 App Store Connect API", "REST")
    System_Ext(xcodegen, "5 XcodeGen 2.46.0", "Release zip")

    Rel(owner, fastlaneRun, "A")
    Rel(fastlaneRun, asc, "B")
    Rel(fastlaneRun, xcodegen, "C")
    Rel(cloudRun, xcodegen, "D")

    UpdateElementStyle(owner, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(fastlaneRun, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
    UpdateElementStyle(cloudRun, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
    UpdateElementStyle(asc, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
    UpdateElementStyle(xcodegen, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
```

```mermaid
C4Container
    title Legend

    Person(actor, "Person", "Owner who runs a lane")
    System_Boundary(unit, "System boundary") {
        Container(part, "Container", "Tech", "Runnable or deployable unit")
    }
    System_Ext(outside, "External", "App file or API that calls in or is called")

    UpdateElementStyle(actor, $bgColor="#334155", $borderColor="#1e293b", $fontColor="#ffffff")
    UpdateElementStyle(part, $bgColor="#0f766e", $borderColor="#115e59", $fontColor="#ffffff")
    UpdateElementStyle(outside, $bgColor="#64748b", $borderColor="#475569", $fontColor="#ffffff")
```

Numbers on nodes and letters on arrows match the callout table. This repo ships no container: its scripts and lanes run inside each app's runs, from the tag the app pins. [03-components.md](03-components.md) shows their files.

| #   | Description | Why |
| --- | --- | --- |
| 1   | Same actor as the system context. | Runs lanes on a Mac. |
| 2   | `fastlane` in an app, on a Mac, with the shared lanes imported. | A Mac runs what the cloud runs. |
| 3   | Xcode Cloud building an app, whose shim fetches the cloud scripts. | Each app runs exactly the release it names. |
| 4   | Builds, tester groups and Xcode Cloud runs. | The team's API key signs each call. |
| 5   | The XcodeGen release the post-clone script downloads. | One generator version for every app. |
| A   | Runs `fastlane archive` or a status lane. | A Mac proves a build without uploading it. |
| B   | The shared lanes read builds and groups, and start a run on main. | Nothing here uploads a build. |
| C   | `archive` runs the post-clone script, which fetches XcodeGen. | Local and cloud projects match. |
| D   | The post-clone script fetches XcodeGen unless that version is installed. | A fresh cloud machine has none. |
