# MOO2-SGC 0.4.1

A modular installer and launcher for owned Master of Orion II DOS installations.

Download **MOO2-SGC-Setup.exe** on Windows. It retrieves and verifies the platform launcher from this Release; the launcher imports your owned game, downloads supported upstream components, prepares selectable community mod configurations, verifies the environment, and launches DOSBox.

This release adds GitHub Desktop push-to-release packaging, online bootstrap configuration, signed staged-update restart, known-folder import, and one-action preparation. Original DOS 1.31 and community 1.50.26 are supported package targets. PRSL and the proposed Chat extension remain independent, unavailable features; no experimental engine hook is installed.

No commercial game data, DOSBox binary, signing private key, GitHub token or cloud credential is included. Upstream patch/runtime downloads are explicit and checksum-verified.

**Test limitations:** no live MOO2/DOSBox gameplay or multiplayer certification is claimed. The local build environment executes Linux only; native Windows/macOS smoke tests are configured separately in GitHub Actions. Executables lack Windows Authenticode and macOS notarization. Version 0.4.1 intentionally has no prerelease suffix; see the test report for actual evidence rather than inferring certification from its name.

Use a new single-player game/save/reload as the first user acceptance test. Keep the source installation unchanged and report failures with the launcher's exported diagnostics.
