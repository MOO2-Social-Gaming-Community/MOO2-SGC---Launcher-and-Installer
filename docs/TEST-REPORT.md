# Test report — MOO2-SGC 0.4.2

## Result and scope

The supported Steam import, CD installation lineage, independent environments and signed application update passed the executed tests described below. **No DOSBox/MOO2 game execution, live two-player match, or Windows/macOS acceptance test was performed here.** The local execution platform was Linux x64. Other platform artifacts were cross-compiled.

| Test family | Result | Evidence |
|---|---:|---|
| Go tests, race detector enabled, real owned fixtures supplied | 89 top-level tests / 191 including subtests passed | `evidence/0.4.2/go-tests.jsonl` |
| Go static vet | Passed | `evidence/0.4.2/go-vet.log` |
| Compiled manager with actual Steam/CD/patch files, plus real 0.4.1 migration | 48 checks passed | `evidence/0.4.2/lineage-acceptance.json` |
| Actual signed 0.4.1 bootstrap installing signed 0.4.2 | 12 checks passed | `evidence/0.4.2/signed-upgrade.json` |
| Final signed Linux bootstrap/launcher startup | 11 checks passed | `evidence/0.4.2/native-smoke.json` |
| Browser DOM/interaction with explicit offline transport fixture | 14 checks passed | `evidence/0.4.2/ui-acceptance.json` |
| GitHub Desktop-style autocrlf Git round trip | 10 checks passed | `evidence/0.4.2/git-desktop-roundtrip.json` |
| Release-publisher safeguards | 7 tests passed | `evidence/0.4.2/publisher-tests.log` |

These are different layers of evidence, not 295 independent proofs of gameplay. Existing historical reports under earlier version directories were not relabeled as new results.

## Original failure and regression

The 0.4.1 Windows report showed that the launcher loaded and a DOSBox runtime was present, but the known Steam `ANWINFIN.LBX` differed from the single old DOS 1.31 asset fingerprint. The new importer first identifies the actual DOS executable, then validates a complete supported asset set. The supplied Steam `Orion2.exe` is 1.40b23 even though its README and Windows executable refer to 1.31.

The new compiled manager imported the actual Steam folder and ZIP, preserved its known ANWINFIN variant, excluded source saves/store wrappers, built baseline and current environments, and left every source file and source archive unchanged. A deliberately modified Steam asset was rejected without replacing the previous imported-source pointer.

## CD reconstruction

| Target from supplied CD 1.2 | Verified installed files | Observed lineage |
|---|---:|---|
| 1.2 | 400 | 1.2 |
| 1.31 | 408 | 1.2 → 1.31 |
| 1.40b23 | 408 | 1.2 → 1.31 → 1.40b23 |
| 1.50.26 | 536 | 1.2 → 1.31 → 1.40b23 → 1.50.26 |

Each target's executable matched its compiled expected SHA-256. The reconstructed baseline is byte-for-byte identical to the clean Steam DOS executable. This is file-transform/output verification, not execution of the historical patcher or the game. The Windows patcher's XP/system-integration code is never run by the manager.

The Steam branch did not download or require the official 1.31 prerequisite. Attempting to reconstruct original CD 1.2 from Steam's later executable was refused.

## Recovery, updates and migration

Tests modified a managed engine and confirmed launch refusal. Reconstruction restored the exact engine while preserving a synthetic save and user audio settings. The previous generation remained present. Source archives were hashed before and after testing and remained identical.

The actual original 0.4.1 manager built a schema-1 environment; 0.4.2 verified it, rebuilt it with the baseline lineage, preserved the synthetic save and retained the old generation. This is forward migration only: old 0.4.1 cannot interpret new source/profile features and schema-2 receipts after an application downgrade.

An unmodified 0.4.1 bootstrap authenticated and installed the signed 0.4.2 package using the original key identity. The trusted revision advanced from 41 to 42; user-data sentinel and previous application generation remained. Reusing old metadata was rejected. No private key was needed or present on the test client.

## Network, browser and process limitations

Official 1.31 package import was tested with the actual archived update and a repackaged copy served through a local TLS test server. Each required installed member was checked against compiled hashes. The live public patch endpoint and live GitHub Release were not fetched in this acceptance run. Publication happens only when the owner pushes and the GitHub workflow succeeds.

Browser navigation to localhost was blocked by the execution environment's browser policy. The UI test therefore used an explicitly labeled offline DOM/transport fixture, with responses supplied by the running manager's real HTTP API through the test harness. Controls, resolutions, version choices and narrow layout passed; this is not an end-to-end native browser-network acceptance test. No browser policy was disabled. The screenshot is of the current interface under that fixture.

The launch lifecycle test used an executable named `dosbox` that prints **TEST DOUBLE ONLY — NO DOSBOX OR GAME EXECUTION** and exits. Its log was checked explicitly. It proves process/config/log behavior only, not title-screen reachability, sound, input, a loaded save, or multiplayer synchronization.

## Reproduce

From `manager/`, run `go test -race -count=1 ./...` and `go vet ./...`. Optional owned-file tests require `MOO2_TEST_ASSETS` to point to the private directory containing the supplied source archives; they skip when the variable is absent and must never require game files in public CI.

Use `tests/lineage_acceptance.py` with `--assets`, `--launcher`, `--old-launcher` and `--output` to repeat the real-file integration. Use `tests/native_smoke.py` and `tests/repository_roundtrip.py` for the final prepared artifacts. `tests/ui_acceptance.py --offline-dom` repeats the explicitly mediated browser test; omit that flag only where direct local browser navigation is supported.

The owner's next required evidence is the Windows checklist: update to 0.4.2 → recognize Steam 1.40b23 → prepare current 1.50.26 → title screen → new game → turns → save → exit → reload.
