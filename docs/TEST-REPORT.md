# MOO2-SGC 0.4.6 — development and regression test report

Completed: 2026-10-06T06:36:42.594874+00:00. Public release revision: 46. This is a launcher/update/kernel repair release, not a claim of new in-game PRSL or matchmaking functionality.

## Verdict and exact scope

The signed 0.4.6 release passes the internal update, existing-install migration, real-owned-file reconstruction, kernel, configuration, and source/package integrity tests below. The final Linux executables actually ran. Windows x64 and both macOS architectures compiled but were not executed here. **No DOSBox emulator or MOO2 game ran in this cycle.** No real network game, live GitHub download or Dopefish availability is certified.

## Reproduced failures

1. **Wrong launcher version:** in a controlled local HTTPS fixture serving genuine previously signed release bytes, the unmodified 0.4.5 bootstrapper accepted 0.4.2 from GitHub's `/latest` address. Its first-valid-signature policy did not require the selected launcher to be at least the setup version. The new executable skips that stale release, requests its pinned 0.4.6 endpoint, installs the right package, verifies the installed executable and reports 0.4.6. It refuses to hand off to 0.4.2 when only old metadata is available.
2. **Missing network driver:** the unmodified 0.4.2 manager imported the user's real Steam source but produced a Community environment without RKERNEL.COM. The old verifier and old launch gate accepted it. The new manager detects the same old environment before starting any game process, imports the owned archived driver into a private verified cache, and repairs the profile while preserving synthetic saved-game bytes.

These results reproduce a cause that matches the user report. They do not prove whether the user's live GitHub publication was incomplete, latest selection was stale, a cached response was received, or an old local shortcut was used. That distinction needs the actual bootstrap log. The new identity banner and last-launch records make it visible.

## Final test results

| Suite | Passed | Scope |
|---|---:|---|
| Go tests with race detector | 111 top-level; 215 including subtests | Run with `-race -count=1` against final source; these are not 326 independent tests. |
| Publishing safeguards | 10 | Includes stale latest repair, byte-identical re-publication and refusal to demote a newer release; mocked `gh`, not live GitHub. |
| Stale release / actual signed upgrade regression | 15 | Real 0.4.5 setup, real 0.4.2 and 0.4.6 signed packages, actual Linux processes; local HTTPS proxy fixture. |
| Old game environment / kernel repair regression | 36 | Actual old/new Linux managers and owned Steam, rkernel and 1.50.26 archives. |
| Full portable real-file regression | 50 | Canonical manual b23, Steam comparison, CD 1.2 → 1.31 → b23 → 1.50.26, reconstruction, rollback and file preservation. |
| Explicit local-ZIP regression | 9 | Actual supplied canonical ZIP; no legacy exact-ZIP path; source unchanged. |
| Portable/bootstrap existing-root integration | 18 | Actual signed setup, offline installation, launch/exit, locking, same-key upgrade and retained user data. |
| Compiled Linux bootstrap smoke | 11 | Actual installed launcher process, signatures, API and shutdown; no game. |
| HTTPS interrupted/resumed application update | 15 | Actual executables, local HTTPS fixture, staged update/restart, retained user data. |
| Browser interface | 24 | Real local API behind an explicitly substituted browser transport, not native loopback navigation; version/path banner, import, service choices, disabled SGC, layout. |
| GitHub Desktop-style Git checkout | 10 | Disposable Git repository with `core.autocrlf=true`; source/package bytes and signatures retained. |
| `go vet` / JavaScript syntax | PASS | Static checks, separate from the counts above. |

All listed tests completed. An initial combined tool call timed out during the longer portable suite; a subsequent full run completed all 50 checks. A first Git checkout test found six rewritten CMD scripts had LF line endings contrary to their CRLF attributes. Those were corrected, all packages rebuilt and signed, and the checkout test passed. Browser review also caught a local profile retaining an invalid empty service choice; local profiles now retain Direct/LAN as the valid inactive default. Final artifacts were rebuilt after that correction and retested.

## Kernel assertions

Required file: `RKERNEL.COM` (31,095 bytes), not an EXE.

Canonical SHA-256:
`18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97`

Accepted historical input:
`0cf378f98f00e6308805cf5d0678e7478cf055c5ff95a827d1645a77eb5f5013`

Only this known input receives the exact three-byte correction. The output was compared byte-for-byte with the kernel in the user's Steam archive. Unknown, ambiguous, oversized, wrongly named or symlinked inputs are refused. No generic kernel binary is bundled in the public distribution.

For every 1.40b23/1.50.26 launch, including local play, the manager checks that canonical file in the **actual selected game's directory**. Merely having it in a ZIP, the baseline folder or a different profile is insufficient. Generated network AUTOEXEC selects the mounted C drive and its root before executing the game. The original direct-program standalone harness invocation remains intact.

Driver import changes only a private verified cache. Reconstruction applies it to a new game generation, verifies the result, preserves same-engine save bytes, and retains the preceding generation. Tests deliberately removed or corrupted the managed driver and confirmed preflight refused to start the process; repair restored it. The original supplied archives were unchanged. Save preservation means bytes, not a live semantic game-load test.

## Application identity and recovery

Setup enforces a local numeric three-part version floor, validates the signed package against that floor, verifies the actual installed file set, and runs an executable version health check before handoff. Explicit rollback is separate from automatic stale-release fallback.

The page exposes actual launcher version/executable, application root and game root. The kernel status names its exact target. New logs under userdata/logs include `last-network-preflight.json` and `last-launch.json`; bootstrap.log remains under the application root's logs directory. Do not share an active browser authentication token.

The signing identity is unchanged. Prepared signed metadata expires 2027-01-04T06:34:38Z. Neither update verification nor this report is an Authenticode/notarization claim. The public bundle contains no commercial game files, DOSBox runtime, private signing key, or fonts.

## Network and unimplemented features

Generated Direct Host/Join and Dopefish Host/Join commands were checked. Both Dopefish roles connect as clients to the shared IPX service; game creation/joining remains in MOO2's own menu. Tests do not connect to that external service and do not assert it is up. Process tests use an explicitly labelled shell test double instead of DOSBox.

PRSL and the new Chat extension remain disabled. MOO2-SGC Online is a disabled future option, not a deployed matchmaking backend. No network/game protocol or executable hook was added in this release.

## Reproduction and evidence

See `evidence/QA-SUMMARY.json` and the individual JSON/log reports. The new focused scripts are `tests/stale_release_acceptance.py` and `tests/upgrade_kernel_acceptance.py`; run their `--help` for private-fixture inputs. Never commit game fixtures or signing keys when repeating them. Existing portable and ZIP scripts cover the canonical baseline and historical lineage.

Final tested Linux launcher SHA-256: `84915ed7f3d263bca81cfb2b41931e12493339c8f9ec25d7bf4fd890e1cfd302`.

Final manifest SHA-256: `1425971e53a0f0e0d54ff88422df0d6f86649f6da019178adc1bfeb7dad9218d`.

The Windows acceptance checklist remains mandatory: correct 0.4.6 identity → correct root → refreshed owned source → repaired/verified selected profile → title and save/reload → direct network game → Dopefish. A launcher process starting, file hashing success, or generated IPX command is not evidence of actual multiplayer success.
