# MOO2-SGC 0.4.6

This maintenance release addresses two linked failures: a new installer accepting an older signed launcher through GitHub's latest endpoint, and a legacy managed game lacking RKERNEL.COM.

Setup now enforces a local launcher-version minimum, retries the version-pinned metadata endpoint, rechecks the actual executable before handoff, and prints version/root identity. Offline installation is subject to the same floor. Explicit rollback remains an explicit, separate operation. Release publication verifies the latest pointer without replacing existing signed bytes or demoting newer releases.

The manager now validates the kernel in the exact game directory before every 1.40b23/1.50.26 launch, exposes that status, and records paths and hashes in launch/preflight logs. It supports a verified owned kernel cache to repair old incomplete source snapshots through staged reconstruction. Normal complete baseline imports supply the file without a separate download. Commercial files remain outside public packages.

Keep the baseline at C:\Games\MOO2-SGC. Install the new integration kit, confirm 0.4.6, explicitly prepare the Community profile from the complete owned baseline ZIP, and verify before opening Network Game. Same-engine saves and prior generations are retained. Existing AppData installs are not silently relocated.

Direct/LAN and Dopefish selections remain available; PRSL, new Chat and MOO2-SGC matchmaking are not enabled. Windows/macOS execution, real gameplay and external-service availability require user acceptance. See docs/TEST-REPORT.md and evidence/ for the test scope and actual results.
