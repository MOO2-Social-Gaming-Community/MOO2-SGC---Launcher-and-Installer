# MOO2-SGC Launcher and Installer — 0.4.2

A small signed bootstrapper and modular environment manager for owned Master of Orion II installations. Windows, Linux and macOS binaries share the same installer and game-preparation logic. **Real game execution still requires user acceptance testing.**

## This release fixes the Steam source rejection

0.4.1 recognized one legacy DOS 1.31 snapshot. It could find the Steam folder but rejected `ANWINFIN.LBX` because legitimate Steam assets differ from that snapshot. 0.4.2 recognizes the supplied English Steam **DOS 1.40b23** engine and its verified asset set, as well as the supplied English CD **1.2** files. It reads the DOS executable, not the Windows executable, README or ZIP filename, to select the source recipe.

- Steam source: **1.40b23 → 1.50.26**. No downgrade and no unnecessary 1.31 prerequisite download.
- Fresh CD source: **1.2 → official DOS 1.31 → 1.40b23 → 1.50.26**.
- Legacy 1.31 source: **1.31 → 1.40b23 → 1.50.26**.

**1.40b23 is the effective SGC baseline. 1.50.26 remains the default current community profile.** Baseline, original CD and legacy official profiles are separately selectable. Future fan versions require an exact supported package, not blind updating.

All patches affect a separate managed copy. Original Steam/GOG/CD files and saves are never modified or uploaded. Unknown editions are rejected with a specific diagnostic, not silently accepted. GOG copies must match a recognized content set; no universal GOG/localization compatibility is claimed.

## Publish using GitHub Desktop

Extract the repository ZIP contents directly into your existing clone, preserving `.git`. Commit and push the default branch. Wait for **Publish prepared release** to publish **v0.4.2**. The prepared public files are in `release/0.4.2/`; do not edit them or the signed manifest.

The repository is `MOO2-Social-Gaming-Community/MOO2-SGC---Launcher-and-Installer`. GitHub Releases is primary; optional R2 is not provisioned. This release uses the **same signing key as 0.4.1**. No signing secret needs to be added for publishing already-signed artifacts.

## Update and first test

Close any running game. In 0.4.1 use **Check signed release → Stage launcher update → Apply staged update and restart**, or run the new standalone setup after v0.4.2 is published. Confirm **0.4.2** appears in the launcher.

Select **Community — standard**, paste your actual Steam game directory under **Prepare your owned game**, and press **Prepare game for play**. Your already installed DOSBox is reused. Verify the reported source is **Steam English DOS 1.40b23**, then press **Launch MOO2**. The first acceptance target is title screen → new single-player game → several turns → save → exit → reload.

See [START-HERE](START-HERE.md), [Windows checklist](docs/TEST-CHECKLIST.md), [test report](docs/TEST-REPORT.md), and [source/lineage design](docs/SOURCE-AND-LINEAGE.md).

## Scope

Source recognition, patch transforms, independent workspaces, community mod selection, repair, signed application updates and launch configuration are implemented. PRSL and the new Chat extension are still independently disabled. No matchmaking/relay, automatic lobby discovery, or live PRSL hook is enabled. Do not infer gameplay certification from file checks or process startup.

No commercial game archives, full game executables, saved games, private keys, or DOSBox runtime are included in the public repository. See [third-party notices](THIRD-PARTY-NOTICES.md). Keep the owner's existing signing backup outside the repository.
