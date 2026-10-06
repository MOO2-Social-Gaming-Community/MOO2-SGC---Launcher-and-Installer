# Start here — MOO2-SGC 0.4.5

## Two different destinations

**Repository ZIP → your GitHub Desktop clone.** Extract its contents into the clone, preserving `.git`; commit and push. Wait for Publish prepared release and a published `v0.4.5`. Do not extract the game/harness into the clone.

**Portable integration ZIP → `C:\Games\MOO2-SGC`.** This is an overlay containing only SGC code and signed launcher packages. Preserve a separate backup of the working handoff before applying it. The archive contains no enclosing directory; do not create `C:\Games\MOO2-SGC\MOO2-SGC` accidentally.

## Required local inputs

Keep your `Master of Orion 2 - v1_40b23.zip` beside the new scripts. The uploaded `(3)` filename is also recognized when the archive bytes match; no repacking is needed. Keep the entire tested DOSBox distribution at `runtime\windows`, not just `dosbox.exe`.

Your existing `game\` is not deleted on startup. On explicit preparation it is backed up in full, and same-version saves are retained. Existing handoff preparation/launch scripts are not overwritten, but do not run them simultaneously with the manager. The new SGC path owns preparation and verification once adopted.

## First test

1. Run **START-MOO2-SGC.cmd**. First launch uses the included signed offline application package. Confirm version **0.4.5** and portable root `C:\Games\MOO2-SGC` in the dashboard.
2. Keep **Portable baseline — 1.40b23**, leave the source field blank, and click **Prepare game for play**. This reuses the local verified runtime, recognizes the root archive, stages a clean copy, normalizes it and preserves the previous game directory. It does not start the game or apply 1.50.26.
3. Click **Launch MOO2**. Confirm Ver 1.40b23, 800×600 window, audio/input, new game, several turns, save, exit, restart and load that save. The original archive must remain unchanged.
4. Once baseline acceptance passes, choose **Community — standard** and prepare 1.50.26 separately. Its game and saves live under `userdata\environments\community`, not over the baseline.

The local browser tab is the launcher UI. Keep its console process open. Closing only the browser tab does not release the application lock; use **Exit launcher**. The command helpers below require the dashboard/game to be closed first.

## Optional command helpers

- `SGC-Prepare-Baseline.cmd`: prepare without a browser; accepts a source ZIP/folder dragged onto it, or blank discovers the canonical archive.
- `SGC-Play-Baseline.cmd` and `SGC-Play-Fullscreen.cmd`: launch the installed, verified baseline with the current signed launcher.
- `SGC-Verify-Baseline.cmd`: file verification only, not game execution.
- `SGC-Recover-Baseline.cmd`: recover an interrupted baseline transaction. Failed candidate directories are retained, not erased.

Do not delete a lock merely because a command says another manager is open. Close the actual processes first. A stale lock after a crash is distinct from a pending game-directory transaction; recovery does not bypass process locks.

## Updates and old installations

Check signed release → Stage launcher update → Apply staged update and restart. These operations update the application, not the game engine/profile or runtime. Existing verified applications can launch offline. First-time offline installation still requires nonexpired authenticated metadata and the correct system clock.

The same signing identity is retained. Keep your private key backup outside both the working root and repository. The Windows EXEs remain unsigned by Authenticode; do not disable security software or elevate to force a blocked test.

An existing 0.4.2 AppData installation is not relocated by its updater. To use the new portable layout, use the integration kit in the new root. AppData saves/settings remain untouched and are not silently merged into this baseline. Older imported snapshots omitted RKERNEL.COM; a rebuild from those snapshots may request reimport of the original source. Existing schema-2 environments can still be verified.

## Failure evidence

Record the first exact error and whether MOO2's title screen actually appeared. Export diagnostics from the dashboard. Bootstrap log: `C:\Games\MOO2-SGC\logs\bootstrap.log`. Game process logs and diagnostic ZIPs: `C:\Games\MOO2-SGC\userdata\logs` (the UI returns exact paths).

Do not post game data, saves, private signing material, or the active local-dashboard URL/token. Do not patch generated files to bypass verification. Keep the previous known-good harness available as your recovery reference.
