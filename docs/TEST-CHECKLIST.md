# Windows acceptance — 0.4.5 portable integration

## Before changing the working reference

Keep a separate known-good copy of your current harness. Close old launchers and DOSBox normally. Keep the private game ZIP and signing backup outside the GitHub clone. Do not run the supplied historical PowerShell refresh at the same time as the new manager.

## Portable baseline

Extract the integration overlay into C:\Games\MOO2-SGC. Confirm there is one root, a full runtime\windows distribution, the owned baseline ZIP, and START-MOO2-SGC.cmd. Run that script normally; no administrator mode or security disabling. First use installs the authenticated offline launcher; next use verifies and opens it without reinstalling.

Confirm 0.4.5, portable root, baseline profile and windowed default. Leave source blank. Prepare game for play must recognize the manual b23 source, create/normalize game, retain an existing game backup and report verification success. It must not install 1.50.26 for this profile or auto-start DOSBox.

Confirm actual game: Ver 1.40b23, video/music/effects, mouse and keyboard. In the default 800×600 window, the game image should now fill the available 4:3 drawable area without the avoidable black padding seen in 0.4.3. Focus another application and check the intended pause/mute behavior. Start new game, play several turns, save, exit normally, close/reopen launcher and reload. Test fullscreen separately; do not treat it as proven by default-window success.

## Explicit local ZIP regression

Before relying on Steam detection, paste the full path to the recognized local `Master of Orion 2 - v1_40b23.zip` into **Prepare your owned game** and run **Prepare game for play**. 0.4.5 must identify the archive by its contents, prepare the selected profile, and leave the ZIP byte-for-byte unchanged. This specifically retests the 0.4.3 bug where the source field incorrectly treated a selected owned-game ZIP as the old exact `base.zip` payload format.

## Preservation and failure boundaries

Verify source ZIP hash before/after (expected in docs/PORTABLE-HARNESS.md). Check Steam/GOG is untouched. Prepare/repair the same baseline and reload its save. Confirm a backup directory is retained. Do not deliberately corrupt the only valuable game or save. Exit launcher before using SGC command helpers. Each helper must return meaningful failure instead of bypassing a running-process lock.

A journal error and a stale process lock are different. Recover interrupted baseline only after all game/launcher processes are closed. Unknown or tampered source/runtime files must fail without silently falling back to another emulator.

## Optional community version

Choose Community — standard. Prepare 1.50.26 and verify its distinct directory under userdata\environments. Recheck that game\ORION2.EXE and baseline save are unchanged. Launch current community as a new game. Its saves are separate; do not assume cross-version save compatibility. PRSL and new Chat must remain unavailable independently.

## Update and later networking

After publishing v0.4.5, check/stage/apply signed release. Same-version reinstall tests handoff, not future-version migration. Confirm root marker, source, runtime, profiles, game and saves remain. Existing AppData installs are not automatically moved.

After single-player/save/repair success, repair the 1.50.26 profile once so the exact LAN-fixed `RKERNEL.COM` is present. `Verify files` must pass before Network Game. Test **Direct / LAN** with two computers using matching compatibility fingerprints. Then test **moo2.thedopefish.com**: both Create and Join should launch with `IPXNET CONNECT moo2.thedopefish.com 213`; inside MOO2 one player creates a named Network Game and the other joins it. Record whether the external service is reachable; MOO2-SGC does not control its availability. MOO2-SGC Online remains disabled and PRSL remains a separate later test.

Record exact failure, OS build, chosen profile/source/runtime, whether title screen appeared, and relevant bootstrap/game log. Redact active loopback tokens and personal file paths before public issues. Do not post commercial game bytes or private signing keys.
