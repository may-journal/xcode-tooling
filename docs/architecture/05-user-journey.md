---
relatedConfigurations: ['../../.fitnessrc.json']
---

# User Journey

## Legend

```mermaid
flowchart TB
    subgraph types["Node types"]
        persona((1 Persona)):::persona
        screen[2 Screen]:::screen
        action(3 Action):::action
        system[(4 System)]:::system
    end

    classDef persona fill:#eef2ff,stroke:#6366f1,color:#312e81,stroke-width:2px
    classDef screen fill:#f8fafc,stroke:#64748b,color:#1e293b,stroke-width:2px
    classDef action fill:#fff7ed,stroke:#f97316,color:#9a3412,stroke-width:2px
    classDef system fill:#ecfdf5,stroke:#10b981,color:#064e3b,stroke-width:2px

```

Numbers match the callout table. A "screen" is a pull request page or TestFlight, since the tooling has no GUI.

| #   | Description | Why |
| --- | --- | --- |
| 1   | Persona | The owner or an agent; both use pull requests. |
| 2   | Screen | What the person reads. |
| 3   | Action | A merge the person performs. |
| 4   | System | A step that runs by itself. |

## Journey

```mermaid
flowchart TB
    owner((1 Owner or agent)):::persona
    clone[(2 Post clone generates the project<br/>Req 0001)]:::system
    mismatch[(3 A download failing its checksum stops the run<br/>Req 0001)]:::system
    notes[4 TestFlight notes hold the merge's entry<br/>Req 0002]:::screen

    owner --> clone
    clone --> mismatch
    clone --> notes

    classDef persona fill:#eef2ff,stroke:#6366f1,color:#312e81,stroke-width:2px
    classDef screen fill:#f8fafc,stroke:#64748b,color:#1e293b,stroke-width:2px
    classDef action fill:#fff7ed,stroke:#f97316,color:#9a3412,stroke-width:2px
    classDef system fill:#ecfdf5,stroke:#10b981,color:#064e3b,stroke-width:2px
```

Numbers match the callout table. Components: [03-components.md](03-components.md).

| #   | Description | Type | Why |
| --- | --- | --- | --- |
| 1   | Owner or agent. | Persona | Same path for both. |
| 2   | The app's shim runs this release's post-clone script from the app's root. | System | One XcodeGen version, per [0001](../requirements/0001-post-clone.md). |
| 3   | A download that fails its SHA-256 exits 1, naming the checksum. | System | No unverified binary runs, per 0001.1. |
| 4   | After the archive, the What to Test notes. | Screen | Written per [0002](../requirements/0002-tester-notes.md). |
