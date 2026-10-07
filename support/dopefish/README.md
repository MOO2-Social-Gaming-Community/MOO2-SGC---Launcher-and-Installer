# Network checks and remembered profiles — 0.4.7

## What this release is for

The 0.4.6 console correctly upgraded a previously installed 0.4.2 launcher. Its profile selector nevertheless always reopened on baseline 1.40b23, and the disabled Core menu misleadingly displayed 1.50 standard. Dopefish configuration existed, but a configured command did not prove a working IPX connection.

0.4.7 packages the earlier Dopefish check into the manager, adds profile persistence and accurate version/network summaries, and preserves all existing engine, kernel and signed-update safeguards. This is not a new game patch or a claim that the remote Comm Failure is resolved.

## Game versions and persistence

Fresh installations still default to the independent **Portable baseline — 1.40b23**. Select **Community — standard** to use **1.50.26 / ORION150.EXE**. 1.50 is not applied to the baseline in place.

Selecting a saved profile writes only `userdata/ui-selection.json`. The preference survives application updates, browser reopen, and changing localhost ports because it is not kept in browser storage. It does not prepare a game, change a ruleset, or copy saves. Missing/invalid selection preferences fall back to baseline. An upgrade from 0.4.6 has no reliable saved selection to infer; choose Community once.

**Save profile** retains edited role/service/ruleset settings. Preparing a profile also saves it. Launch uses the visible choices and remembers the selected existing profile but does not overwrite saved profile options; the CLI fullscreen override remains launch-only.

A baseline or other pre-1.50 engine displays **Not applicable — original [version] rules**. The Core selector contains no 1.50 choice in that state. Changing back to Community enables its Core/add-ons.

## The network plan

The visible plan uses the same backend command generator as the generated game configuration. It shows selected game executable, endpoint and exact IPXNET command; it is configuration evidence, not reachability evidence.

For **Dopefish**, Create and Join both use:

```
IPXNET CONNECT moo2.thedopefish.com 213
```

The player creates or joins the named game inside MOO2. Direct/LAN stays separate: the creator starts a local IPX server, joiners connect to it. Local/single-player config does not create an IPX tunnel.

For a controlled test, both machines should use **Community — standard / 1.50.26 / 1.50 standard**, then select Create or Join plus Dopefish, **Save profile**, prepare/verify and launch. Do not substitute Community — multiplayer unless both players deliberately choose its different ruleset.

## Connection check in the launcher

Use **Check Dopefish connection (no game)** under Play together. Confirm the explicit network request. The existing DOSBox runtime opens in a separate window and runs CONNECT followed by `IPXNET STATUS`. Read that screen and capture it when unclear. Type **EXIT** to close the diagnostic before starting MOO2.

No game is mounted or launched. The manager does not treat a process exit code as successful connectivity. While that diagnostic is running, other launch/prepare/update/quit actions are blocked through the normal process gate. If a runtime is missing or altered, restore/install it through the manager rather than bypassing verification.

The portable **CHECK-DOPEFISH.cmd** provides the same diagnostic through the signed CLI, with runtime verification and application locks. Exit the launcher first to use that helper; otherwise use the in-launcher button. It can use the signed offline 0.4.7 package and does not need a fresh GitHub download when that package/current launcher is available.

## Files and evidence

- `userdata/network-checks/dopefish-*/dopefish.conf`: dedicated diagnostic config. Generated afresh; does not overwrite the game's config.
- `userdata/logs/last-network-check.json`: runtime, arguments, endpoint, local directory and log. `connection_verified` remains false: automatic network-result classification is not implemented.
- `userdata/logs/dopefish-*.log`: host-side DOSBox process output. DOS shell messages may only be visible on the DOSBox screen.
- `userdata/logs/last-launch.json`: selected game version, game executable/directory, generated config path, network plan and RKERNEL preflight.
- **Export diagnostics** includes the last independent connection-check record without commercial game files or private signing material.

`support/dopefish/dopefish-diagnostic.conf` is the text reference shipped in the kit; the launcher uses its identical embedded resource. Editing that reference does not modify a signed launcher or generated game config. Readonly generation is intentional.

## Limits

A successful diagnostic proves only that its own tunnel connected at that moment; the actual game opens its own session afterward. Both machines must still agree on game version/rules and join the same in-game session. An inability to connect can be remote-service or network-path related; this release does not assert the Dopefish server is currently reachable.

PRSL, new Chat and SGC Online are unavailable, independent future components. The known canonical RKERNEL.COM remains required beside the selected engine. Source archives, Steam/GOG folders, baseline runtime and same-engine saves are not replaced by this diagnostic.

## Primary references reviewed for this packaging cycle

- DOSBox IPX CONNECT and STATUS: https://www.dosbox.com/wiki/Connectivity
- DOSBox Staging IPX settings: https://www.dosbox-staging.org/0.83/manual/networking/ipx/

These references support command semantics, not uptime of a third-party server. No live Dopefish session was verified during the build.
