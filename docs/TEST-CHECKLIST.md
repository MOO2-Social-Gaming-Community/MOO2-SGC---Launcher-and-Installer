# Windows 11 acceptance — 0.4.1

## Publish and download

Extract the repository ZIP contents directly into the cloned root, preserve `.git`, commit all supplied files and push the default branch. In GitHub's Actions tab, wait for **Publish prepared release** to succeed. Confirm a published, non-draft `v0.4.1` release exists and has `MOO2-SGC-Setup.exe`, launcher ZIPs, `manifest.json` and `manifest.sig`. The public installer requires a public repository and completed asset publication, not merely a source commit.

Download setup to a fresh folder with **no** launcher ZIP, manifest or private game archive beside it. Run it normally. Do not disable protection or elevate to administrator to force a blocked executable.

## First installation

Expected: download/verification progress, installed launcher local browser screen, version 0.4.1, and a Start-menu shortcut when Windows permits it. The default data root is `%APPDATA%\MOO2-SGC\stable\userdata` and application files remain outside that data directory.

Use **Community — standard**. Under **Prepare your owned game**, paste the path to the exact archive previously supplied (`Master of Orion 2.zip`, or the older private bundle's `payloads/base.zip`). A folder matching the fingerprinted DOS 1.31 baseline is also accepted; other Steam/GOG editions are not automatically guaranteed. Click **Prepare game for play**. Required patch and runtime downloads are explicit consequences of this action.

Expected: verified source import; pinned 1.50.26 patch and DOSBox available; generated configuration; successful workspace verification; no automatic game start. Optional PRSL and new Chat remain unavailable.

## Actual game acceptance

Click **Launch MOO2**. Confirm title screen, sound, input and screen size. Start a **new** single-player game using the chosen ruleset. Take several turns; save normally; exit; reopen the installed launcher from the Start menu and reload the same profile/save.

Record actual observations, not only that the DOSBox process started. Do not use the installer diagnostics screen as evidence that gameplay works.

After single-player works, test a second machine with matching rules. Host/Join configures the IPX tunnel; still use MOO2's Multiplayer / Network UI. Automatic lobby discovery, a public relay and automatic router setup are not implemented.

## Recovery and updates

Verify the prepared files. Rebuild the same profile and confirm its save survives. Switching to Original DOS creates a separate unpatched environment; keep the patched profile intact. Do not deliberately corrupt a valuable personal installation.

Use **Check signed release → Stage launcher update → Apply staged update and restart**. A same-version stage tests handoff, not a future-version migration. Confirm user data remains. Do not update while a game is running.

## Report failures

Record exact step, exact message, operating-system build, selected source/ruleset/runtime and whether a title screen actually appeared. Retain `%APPDATA%\MOO2-SGC\stable\logs\bootstrap.log`, use **Export diagnostics**, and retain the relevant game-process log. Do not post commercial game files, signing material, or an active local-dashboard URL/token.
