### 2026.11.23.1450

> A Mac that runs a capture is left with no stray timer process when the capture ends.

- Fix: a boot, clone or test that finishes early stops its timer's `sleep`, not only the timer's shell.
- Fix: a step that runs past its bound is still stopped, as before.
- Fix: a timer told to stop before its `sleep` has started still leaves none behind.
- Test: a new acceptance proves no timer outlives a capture.
