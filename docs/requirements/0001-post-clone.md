---
relatedConfigurations: ['../../.fitnessrc.json']
---

# 0001 Post Clone

## Why

Every cloud run of an app generates the same Xcode project, so a build never depends on the machine it lands on. The script installs one XcodeGen release, checked by its SHA-256, and runs the app's own step first.

## Measurement

app cloud runs whose project comes from XcodeGen 2.46.0, or that fail naming the checksum

-

app cloud runs that clone this release

## Requirements

- 0001.1
  - Given Owner has an app with `project.yml` and a download failing the checksum, one with a committed project, then XcodeGen 2.46.0 installed
    - When Xcode Cloud runs `ci_post_clone.sh` from the app's root
      - Then the app's step runs, then it exits 1 naming the checksum; the second skips; the installed copy generates in the app's root
