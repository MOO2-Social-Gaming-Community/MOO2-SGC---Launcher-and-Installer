# 0.4.6 acceptance checklist

## Application boundary

- Exit all games and old launchers. Launch only the new setup or the portable root START-MOO2-SGC.cmd.
- Confirm setup 0.4.6, launcher 0.4.6 and the intended root. No fallback to 0.4.2 is acceptable.
- Online mode: confirm published v0.4.6 assets and a successful publishing Action. Offline kit: no publication prerequisite.
- Do not modify signed assets, clear trust state, disable protection or delete live locks.

## Game boundary

- Choose Community — standard; explicitly select the complete owned manual b23 source ZIP and Prepare game for play.
- Verify files. Inspect the new network-kernel line and generated game path. RKERNEL.COM must be beside ORION150.EXE, not just in the source archive or baseline folder.
- Source ZIP and original Steam installation remain unchanged. Existing baseline remains 1.40b23.
- Confirm current profile 1.50.26 title, input, audio, new game, several turns, save, exit, relaunch and load.
- Confirm a same-engine profile repair preserves the save; do not deliberately corrupt valuable real data.

## Network boundary

- First test Direct/LAN Create and Join from two machines using the same profile and data.
- Both choose Multiplayer → Network in-game; create one distinctive named game and join it.
- Both clients must progress several turns, save and resume together before certification.
- Then test Dopefish: both clients connect to the third-party service, while creation/joining is done in MOO2.
- External connection failure does not establish a kernel failure. Capture exact DOSBox and game messages.

## Report

Export diagnostics; retain bootstrap.log, last-network-preflight.json, last-launch.json, game log and ORION2.LOG where present. Record exact version, app root, game root, service, role, visible error and whether a real game was entered. Redact dashboard tokens and unrelated private paths before posting publicly.
