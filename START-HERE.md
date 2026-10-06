# MOO2-SGC 0.4.6 — start and repair

## Keep the working root

Use `C:\Games\MOO2-SGC`. Keep a separate backup of the proven harness. The integration kit contains only project files, not commercial game data, DOSBox, or private signing keys. Leave Steam/GOG and the source ZIP unchanged.

## Recommended recovery from the 0.4.2 handoff problem

1. Quit MOO2 and use **Exit launcher** in every open SGC launcher. Closing a browser tab alone does not stop its local server. Do not delete locks while a process is running.
2. Extract the **0.4.6 PORTABLE_INTEGRATION ZIP contents** into `C:\Games\MOO2-SGC`, overwriting only its matching project integration files. Do not extract it inside the Git repository, a versioned component folder, or a second nested MOO2-SGC folder.
3. Run `START-MOO2-SGC.cmd` in that exact root, not an old Downloads executable or old Start-menu shortcut. It installs the signed offline 0.4.6 package, bypassing a stale GitHub latest pointer, then launches the verified installation.
4. Confirm the console and page show **0.4.6**. The page displays the actual executable and application root. New setup will refuse to launch 0.4.2; it no longer treats an older signed release as an acceptable fallback.
5. Select **Community — standard** (1.50.26). Under Prepare your owned game, supply `C:\Games\MOO2-SGC\Master of Orion 2 - v1_40b23.zip`. Click **Prepare game for play**. This explicitly refreshes any incomplete old import and reconstructs a separate current environment. Same-engine saves remain in the repaired profile and previous generation; baseline game/ stays independent.
6. Confirm **Network kernel verified** and **Verify files** succeeds. The kernel path must be in the same managed game folder as ORION150.EXE.
7. Test **Create network game → Direct / LAN** first. Launch and choose Multiplayer → Network in MOO2. Then test the third-party Dopefish service with both players using the same engine/ruleset/assets.

The baseline profile remains 1.40b23. Selecting Community is not an in-place upgrade of baseline game/ or Steam.

## Standalone online setup / GitHub Desktop

Extract the repository ZIP contents into your cloned repository, preserving `.git`; commit and push. Wait for **Publish prepared release** to succeed and confirm `v0.4.6` is a published release. Run the standalone **MOO2-SGC-Setup-0.4.6.exe** afterward. It rejects older metadata and tries its pinned release endpoint. If the release is not published, it stops with a clear error rather than running 0.4.2. The offline kit above works before publication.

Updating through an old AppData installation keeps that installation and data in AppData; it does not silently move saves to the portable root. Use the portable kit for the canonical root. The new page displays the distinction.

## Standalone archived network kernel

The required filename is **RKERNEL.COM**, not RKERNEL.EXE. Normal preparation recovers it from the complete owned game source. A legacy incomplete snapshot can also be repaired by selecting **Packages & updates → Network kernel — owned RKERNEL.COM or rkernel.zip**, entering your owned archive/file path, and importing it. Then **Prepare / repair selected profile**.

Only the two pinned historic hashes are accepted. The older 31,095-byte driver receives the exact three-byte known transformation; the result must equal the canonical SHA-256 `18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97`. Unknown inputs are refused. The active game is not changed by import alone. Do not rename random .EXE files or download an unverified replacement.

## Diagnostics / limits

Under the active application root: `logs\bootstrap.log`. Under its user-data folder: `logs\last-network-preflight.json`, `logs\last-launch.json`, and `logs\game-*.log`. Use **Export diagnostics** too. Logs now identify setup/launcher versions, paths, selected engine, network service, kernel hash and working directory. Do not share private signing material or an active dashboard token.

Package signatures remain separate from Windows Authenticode; these executables are not Authenticode-signed or macOS-notarized. Do not disable system protection to force a test. Linux file/update regressions are not evidence of real Windows/DOSBox multiplayer; the final in-game acceptance remains on the user's machines. Dopefish availability is not controlled or certified by SGC.
