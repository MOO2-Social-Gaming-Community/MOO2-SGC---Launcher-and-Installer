# Windows 11 acceptance — 0.4.2

## Update

Confirm GitHub published v0.4.2. Close the game. In 0.4.1, check/stage/apply the signed launcher update, or run the standalone 0.4.2 setup normally. Confirm the resulting launcher says 0.4.2, existing profiles remain, and the existing DOSBox runtime is detected. Keep the private signing backup outside the repo; no new signing setup is needed.

## Steam source (first regression test)

Select Community — standard and the actual Steam folder previously rejected. Prepare game for play. Expected source version: **1.40b23**; expected lineage: **1.40b23 → 1.50.26**. The importer must not reject the known Steam ANWINFIN variant as legacy 1.31 data. Verify the managed files and confirm the Steam source has not been edited. No game should auto-start.

Press Launch MOO2. Observe title screen, version, mouse, keyboard, sound and display. Start a new single-player game, take several turns, save, exit and reload that save. Record an actual title/game screen, not merely a DOSBox process identifier.

## Baseline

Choose the separate SGC baseline — 1.40b23 profile with Steam as source. Prepare and launch. Confirm no 1.50 Core/mods are enabled. Verify the 1.50 profile remains independently available.

## CD lineage

Use a separate profile and the original CD 1.2 archive/folder. Choose current 1.50.26 and prepare. Expected progression: 1.2 → 1.31 → 1.40b23 → 1.50.26. Record any official prerequisite download error; do not substitute unverified files. Local official 1.31 ZIP import is the supported offline alternative.

Test Original CD 1.2 only with its source provided. A request to downgrade a Steam 1.40b23 source to original CD bytes should be refused, not silently fabricated.

## Preservation and multiplayer

Rebuild the same-engine profile and verify its save survives. Existing other profiles and original Steam files must remain untouched. Do not deliberately corrupt your real source installation.

Only after single-player succeeds, test Host/Join with two computers using matching engine/rulesets. Use the in-game Multiplayer/Network interface. LAN discovery, online matchmaking, relay and live PRSL remain out of scope for this test.

## Report failures

Capture the first failing action, exact message, selected source version/target/profile and runtime. Export diagnostics and retain the game-process log and `%APPDATA%\MOO2-SGC\stable\logs\bootstrap.log`. Remove local dashboard tokens/private paths as appropriate before publishing an issue. Never include source game archives or signing keys.
