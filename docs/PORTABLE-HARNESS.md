# Portable harness integration contract — 0.4.5

## Authority and scope

Inputs are the user's `MOO2-SGC-Portable-Runtime-Handoff-2026-10-05(1).zip`, `MOO2-SGC-Portable-Harness-0.3.0-HANDOFF(1).zip`, and `Master of Orion 2 - v1_40b23(3).zip`. This contract preserves their terminology and tested settings. The handoff documents user-observed Windows gameplay; the supplied 0.3.0 hardening and this new manager integration are not independently certified by a Linux file test. User reports broader harness testing; repeat the final Windows acceptance against the new wrapper.

The existing Go installer/updater remains the application architecture. Portable Harness 0.3.0-HANDOFF and Launcher 0.4.5 have independent versions. This integration translates its normalization and launch contract into native Go; it does not run the old destructive PowerShell refresh implementation.

## Source and normalized target

Canonical private ZIP SHA-256:
`50a8985601bc8f73af0486536c3413978bdf260d5d3363b5175ca809d22a2cbb`

Manual source contains two distinct executables:

| File / role | SHA-256 |
|---|---|
| Source ORION2.EXE, official 1.31 | `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` |
| Source Orion2v140.exe / normalized ORION2.EXE, 1.40b23 | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| Original RKERNEL.COM | `0cf378f98f00e6308805cf5d0678e7478cf055c5ff95a827d1645a77eb5f5013` |
| LAN-fixed RKERNEL.COM | `18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97` |

A newly staged baseline preserves official 1.31 as ORION131.EXE, retains the verified Orion2v140.exe, and copies b23 to ORION2.EXE. Kernel edits are permitted only on the entire matching old file: 0x765C 01→00, 0x765D 74→75, 0x7697 01→00. Output must match the fixed hash. Already-fixed input is unchanged; unknown input is rejected. No original archive is patched, extracted in place or uploaded.

The manual archive has 799 files. The importer copies only 408 explicitly recognized root files into a private content-addressed snapshot; archived saves, installers, wrappers and other unrelated contents remain only in the original archive. New prepared baseline contains additional normalization/configuration files. Existing same-engine harness saves are retained separately from archive import.

Seven CD/manual vs Steam LBX data files differ even though the b23 engine matches. The compatibility fingerprint includes immutable game assets, not just engine and configuration strings. Exact fingerprint equivalence is conservative evidence of file equivalence, not proof of network compatibility; differing fingerprints must not be silently declared compatible.

## Runtime and invocation

DOSBox Staging 0.83.0 Windows executable SHA-256:
`67a907289bad3e25a9be7dd4dc38d882c1c7e74949f8daa103a52476d99f24f5`

`manager/assets/harness-runtime.json` pins all **660** supplied runtime files, including DLLs and resources, by name, length and hash. They are not included in public release payloads. Portable Windows preparation/launch verifies `ROOT/runtime/windows` and will not silently substitute global DOSBox. Download fallback uses the already-pinned official archive and verifies the same extracted file set. A present but damaged local runtime is refused rather than overwritten in place.

Standalone launch is an argument vector, not shell text:

```text
runtime/windows/dosbox.exe
--noprimaryconf
--conf <absolute ROOT/config/moo2.conf>
[--fullscreen, only when selected]
<absolute ROOT/game/ORION2.EXE>
```

Working directory is the game directory. No stale primary user config, autoexec mount, CPU-cycle override or `/skipintro` is added for standalone play. Extra game-local emulator configs are refused. Network mode alone adds its separate IPX autoexec; that path still needs multiplayer acceptance.

Default profile preserves the supplied harness behavior (800×600 window, `fullscreen=off`, forced-borderless fullscreen when requested, `pause_when_inactive=off`, `mute_when_inactive=off`, OpenGL output, auto aspect, sharp shader, SB16 base220 IRQ5 DMA1 HDMA5) and adds the 0.4.5 presentation fix: `viewport = fit` plus `integer_scaling = off`. Those two rendering directives explicitly scale the 4:3 MOO2 image to the available 4:3 window area instead of allowing avoidable internal padding. Fullscreen still uses `aspect = auto`, so widescreen fullscreen preserves the game aspect ratio rather than stretching it. The supplied config began with a stray single-backslash line; integration removes that non-directive.

DIG.INI uses SBLASTER.DIG and DMA_16_BIT=-1; MDI.INI uses SBPRO2.MDI; ORIONCD.INI contains exactly `.\` without a newline. Do not reintroduce the previous launcher's SB16.DIG/absolute CD defaults. These files match the handoff bytes. They remain editable after the initial normalized preparation; changed user audio settings are not silently certified as the original configuration.

## Layout, transactions and recovery

The explicit schema-1 `moo2-sgc-portable.json` marker opts into the portable layout. Only the named `baseline` profile at engine 1.40b23 may use `ROOT/game`. Other profiles, even b23 duplicates, use isolated generation directories under `userdata/environments`.

Preparation stages and verifies a complete candidate first. It then retains the full previous game under `backups/baseline/<id>/game`, renames the candidate to `game`, and commits receipts in `state`. A write-ahead journal covers the directory/receipt steps. It is not a claim that multiple directory/file replacements are one atomic filesystem transaction.

A pending journal blocks launch/preparation until explicit recovery. Recovery restores the prior tree, retains a failed candidate, or completes cleanup only after verifying the committed new tree. Invalid/traversal journals are refused. Raw unregistered harness backups remain available as files but are not offered as certified manifest-backed rollback choices.

Same-engine repairs preserve saves/race/settings; recognized schema-3 editable audio also survives. Initial adoption or upgrade from old launcher defaults intentionally reinstates the handoff audio rather than copying incompatible old default INIs. The complete previous directory remains backed up. No save migration between engine versions is automatic. A new community profile starts separately, with no baseline saves copied implicitly. Rollback does not merge newer saves into an older snapshot.

No runtime, CD data or application is uninstalled by deleting the source. Application updates alter only signed launcher generations; game repair is a separate explicit action. Closing only the browser tab does not stop its local server. The installation and user-data locks remain until the process exits.

## Backward compatibility and development boundaries

Old schema-2 workspaces and stored source receipts remain readable. Old snapshots that omitted RKERNEL must be reimported before new normalization. Old AppData installations stay put when updated. No registry edits, elevation, PowerShell execution-policy bypass, global DOSBox edits, Steam/GOG file writes, or file-system links are needed.

Cross-platform application builds remain available, but the exact runtime contract validated by the supplied user handoff is Windows. Linux/macOS runtime behavior needs separate game acceptance. PRSL/Chat are disabled. No relay, matchmaking server, self-updating game-engine policy or new gameplay feature is implied.
