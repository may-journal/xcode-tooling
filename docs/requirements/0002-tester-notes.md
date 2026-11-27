# 0002 Tester Notes

## Why

Testers read what changed in each TestFlight build. After an archive, the script writes the merge's changelog entry as the build's What to Test notes.

## Measurement

archived app builds whose notes hold the merge's changelog entry, or none when the app has no changelog

-

archived app builds that run this release

## Requirements

1. Writes the tester notes
   1. GIVEN Owner has an app with changelog entries
   2. GIVEN a head commit adding an older-named one
   3. GIVEN no changelog
   4. GIVEN Xcode Cloud runs `ci_post_xcodebuild.sh`
      1. WHEN a build ends
         1. THEN the build writes no notes
      2. WHEN each archive ends
         1. THEN the notes hold the newest entry
         2. THEN the notes hold the head's
         3. THEN the notes hold none at all
