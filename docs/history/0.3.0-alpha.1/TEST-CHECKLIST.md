# First hands-on acceptance cycle

Record OS, CPU architecture, manager version, DOSBox version, profile, engine, ruleset and the exact error/log path for any failed step. Test the community game first; there is no live PRSL to test yet.

## Windows first

| Step | Expected result | Your result |
|---|---|---|
| Extract the full bundle | Executable and payloads remain side by side | |
| Run START-WINDOWS.cmd | Console remains open; local dashboard opens | |
| Select existing DOSBox | Runtime status shows the selected path | |
| Prepare Community — standard | Operation succeeds; native game remains unmodified | |
| Verify files | `ok: true`; zero immutable failures | |
| Launch | DOSBox and MOO2 title screen appear | |
| New game | Play several turns without crash | |
| Save / exit / relaunch / load | Your campaign survives | |
| Fullscreen option | Next launch respects the selection | |
| Duplicate profile | New profile has independent selection/state | |
| Prepare different Core | New game uses that ruleset | |
| Rebuild original profile | Its own save remains | |
| Activate retained generation | Selected generation launches with its own saves | |
| Export diagnostics | JSON path is shown; no saves/token are included | |
| Exit launcher | Local server closes; manager.lock disappears | |

## Optional network acceptance

Use separate machines and separate writable environments. Compare full verified fingerprints, not only friendly version labels. Start with the same emulator build where possible.

Host starts tunnel; client joins host IPv4/port; both use MOO2's Multiplayer / Network menu. Start a new two-player game. Advance at least ten turns, encounter relevant dialogs, save, exit both clients, reconnect and load. Record whether native waiting/chat, turn resolution and synchronization behave normally. Do not describe PRSL as active: it is not.

## Then Mac and Ubuntu

Repeat the same sequence on the 2015 Intel Mac (check its installed OS against the runtime requirement) and Ubuntu x64. Test Windows host → Mac client, Windows host → Ubuntu client, and a reversed host arrangement. Successful single-player launch is not equivalent to multiplayer acceptance.

## Runtime download test

Test the existing-runtime path first to isolate failures. Then explicitly download the verified runtime on one machine and confirm the version and game launch again. A checksum/download failure must leave the old environment usable. Network errors should not be “fixed” by disabling protections.

## What to send back

The exact failed step, error text/screenshot, `data/logs/diagnostics-*.json`, and the matching `game-*.log` are sufficient to begin diagnosis. Review local paths and host addresses before sharing. Do not send `open-launcher.txt`; it contains the private local UI access token.
