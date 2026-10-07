# MOO2-SGC 0.4.7 — start here

## Update the repository

Extract the repository ZIP contents into the existing GitHub Desktop clone; preserve `.git`. Commit and push the default branch. Wait for **Publish prepared release** to publish **v0.4.7**. Files in release/0.4.7 are already signed with the existing identity. Do not edit/repack them. Keep game archives and signing keys OUT of the repository.

## Update your portable installation

Close the game. Use **Exit launcher**, not merely closing its browser tab. Back up the working harness.

Extract the portable integration ZIP into **C:\Games\MOO2-SGC**, not an extra nested folder. Keep the existing complete `runtime\windows` directory, owned base ZIP, game files and userdata. Run **START-MOO2-SGC.cmd** there. The bundled signed offline package installs/upgrades the application without depending on GitHub publication.

The standalone online Setup executable should be run only after the v0.4.7 Release is published. Replacing source files in GitHub is not a published software Release. The setup line about the **previously installed launcher** describes the old version, not the one being started; check **Launching/ready** and the UI's running identity.

## First 0.4.7 network test

Choose **Community — standard** explicitly to run **1.50.26**. The separate baseline is still 1.40b23 and remains the first-run default. This release remembers your chosen profile for subsequent reopen/restarts.

Choose **Create network game** (or Join on the second machine) and **moo2.thedopefish.com — public IPX**, then **Save profile**. The network-plan box should show **ORION150.EXE** and **IPXNET CONNECT moo2.thedopefish.com 213**. Both players use that same connection; create/join the named game inside MOO2.

Use **Check Dopefish connection (no game)** first. Read CONNECT/STATUS in its DOSBox window. Type **EXIT** to close it. Then prepare/verify the Community environment, confirm Network kernel verified, and launch. Diagnostic success does not certify multiplayer; retain the first failure's exact message.

The standalone **CHECK-DOPEFISH.cmd** runs the same diagnostic after the launcher is closed. It now uses signed setup/launcher locks and runtime checks, not a bare unverified runtime call.

## Preserve the reference

1.40b23 baseline, optional 1.50.26 and owned Steam/GOG installations remain distinct. PRSL, new Chat and SGC Online remain disabled. No public package contains commercial game data or the DOSBox runtime. Existing private signing-key backup is unchanged. Windows Authenticode/macOS notarization are not provided.

See `docs/NETWORK-AND-PROFILES-0.4.7.md` and `docs/TEST-REPORT.md` (or `sgc-docs/` in the portable kit). No Windows/game/live-Dopefish acceptance is claimed by this build.
