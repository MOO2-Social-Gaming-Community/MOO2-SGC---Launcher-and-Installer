# MOO2-SGC 0.4.3

Portable Windows integration using the owner's working 1.40b23 harness.

- Default fresh Windows root is C:\Games\MOO2-SGC; default profile is 1.40b23.
- Recognizes the manual CD-derived b23 archive by exact file fingerprints, including the alternate patch-output executable.
- Adds normalized canonical ORION2.EXE, retained ORION131.EXE and exact-hash RKERNEL LAN correction.
- Uses the supplied DOSBox Staging 0.83.0 direct-program command, 800×600 window, audio files and inactive-window behavior.
- Verifies the complete 660-file local runtime; public SGC downloads contain no emulator, fonts or game data.
- Stages a new baseline, retains complete old game directories, preserves same-engine saves and offers explicit interrupted-transaction recovery.
- Keeps optional 1.50.26/community mods in separate game environments; PRSL/Chat remain disabled.
- Maintains existing signing identity and old-root update compatibility. Offline integration kit can be tested before GitHub publication.

The owner has tested the underlying harness on Windows. This release's new normalization, manager and process integration was tested on Linux using real owned-file fixtures and a labeled process test double. Native Windows gameplay, multiplayer, macOS/Linux game execution and live GitHub download are not certified here. See docs/TEST-REPORT.md and START-HERE.md.
