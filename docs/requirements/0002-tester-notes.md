# 0002 Tester Notes

## Why

Testers read what changed in each TestFlight build. After an archive, the script writes the merge's changelog entry as the build's What to Test notes.

## Measurement

archived app builds whose notes hold the merge's changelog entry, or none when the app has no changelog

-

archived app builds that run this release

Source: `ci_scripts/ci_post_xcodebuild.sh:20`

## Requirements

- 0002.1
  - Given Owner has an app with changelog entries, then a head commit adding an older-named one, then no changelog
    - When Xcode Cloud runs `ci_post_xcodebuild.sh` after a build, then after each archive
      - Then the build writes no notes; the notes hold the newest entry, then the head's, then none at all
