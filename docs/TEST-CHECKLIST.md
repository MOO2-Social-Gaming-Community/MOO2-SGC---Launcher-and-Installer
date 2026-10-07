# Windows acceptance checklist — 0.4.7

1. Close the actual old launcher and game. Extract the portable kit into C:\Games\MOO2-SGC; retain the working runtime and owned source. Run START-MOO2-SGC.cmd. Confirm setup, selected launcher and UI identity all show 0.4.7; a previous-version line is historical, not the target.
2. Select Community — standard / 1.50.26. Set Create (or Join) and Dopefish, then Save profile. Confirm the network plan shows ORION150.EXE and CONNECT moo2.thedopefish.com 213, not STARTSERVER. Do not mix Standard with the separate Multiplayer ruleset on the second machine.
3. Exit the launcher, reopen and confirm Community and the saved role/service are restored. Switch to baseline and confirm its Core menu says not applicable, then return to Community. No game should be upgraded/downgraded simply by switching the selection.
4. Run Check Dopefish connection (no game), confirm the explicit request, and read the actual DOSBox CONNECT/STATUS output. Capture it. Type EXIT. Check the launcher unlocks afterward. No successful process startup is counted as network success.
5. Prepare/verify the Community profile; confirm the exact RKERNEL.COM is verified in that selected directory. Launch, confirm Version 1.50.26 on the title screen, enter Network and create/join the same named session from matching clients. Record the first failure; do not manually modify generated configs to hide it.
6. Retest single player, save, quit and reload. Verify Steam and the separate baseline still launch their original versions. Rebuild/repair only the intended profile, and retain backups.

Evidence: userdata/logs/last-network-check.json, last-network-preflight.json, last-launch.json, Export diagnostics, and a DOSBox screen capture. The stdout log may omit DOS shell status messages. Redact the launcher's private local token/URL; do not upload game files or signing keys. Do not disable protection or elevate privileges to force a failed diagnostic.

After GitHub publication, test the standalone installer separately; the portable kit's offline success is not evidence of a live GitHub download. Native Windows, actual DOSBox/MOO2 gameplay and live external UDP must be observed on the test machine.
