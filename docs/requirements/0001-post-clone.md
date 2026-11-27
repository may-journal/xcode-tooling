# 0001 Post Clone

## Why

Every cloud run of an app generates the same Xcode project, so a build never depends on the machine it lands on. The script installs one XcodeGen release, checked by its SHA-256, and runs the app's own step first.

## Measurement

app cloud runs whose project comes from XcodeGen 2.46.0, or that fail naming the checksum

-

app cloud runs that clone this release

## Requirements

1. Generates the project
   1. GIVEN Owner has an app with `project.yml`
   2. GIVEN a download failing the checksum
   3. GIVEN one with a committed project
   4. GIVEN XcodeGen 2.46.0 installed
      1. WHEN Xcode Cloud runs `ci_post_clone.sh` from the app's root
         1. THEN the app's step runs
         2. THEN it exits 1 naming the checksum
         3. THEN the second skips
         4. THEN the installed copy generates in the app's root
