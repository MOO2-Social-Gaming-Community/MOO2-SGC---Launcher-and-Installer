# Test report — MOO2-SGC 0.4.0-alpha.1

## Summary

| Category | Result | Scope |
|---|---:|---|
| Go tests with race detector | **72 top-level passed; 141 including subtests** | All packages, strict manifests, signatures, transports, install transactions, manager/API behavior. Do not add top-level and subtest totals together. |
| Manager integration | **39 checks passed** | Actual Linux manager HTTP API, offline Chromium UI rendering, real game-file reconstruction, save/config preservation and refusal paths. |
| Signed bootstrap integration | **30 checks passed** | Final compiled Linux setup and launcher; signed offline install, health, handoff, repair, rollback, update staging/apply and source preservation. |
| Legacy migration integration | **7 checks passed** | Actual old 0.3 and new 0.4 Linux binaries sharing schema-1 profile data; reconstruction preserves synthetic save data and previous generations. |
| Additional validation | Passed | go vet, JavaScript/Python syntax, workflow YAML parsing, generated manifest JSON Schema validation, release publication dry-run and refusal of development trust. |
| Native builds | Four targets compiled | Windows x64, Linux x64, Intel Mac, Apple Silicon Mac. Only Linux executed. |
| DOSBox / MOO2 gameplay | **Not run** | No title-screen, turn-resolution, multiplayer or PRSL success claim. |
| GitHub/R2 deployment | **Not performed** | No public feed or actual provider outage exercised. |

## Evidence

Current machine-readable evidence is under `manager/evidence`: `go-tests.jsonl`, `go-coverage.txt`, `integration-tests.json`, `bootstrap-integration.json`, `migration-integration.json`, `publish-dry-run.json`, `BUILD-RESULTS.json` and `RELEASE-SUMMARY.json`. The built-artifact record reports compilation only; subsequent subprocess tests establish actual Linux execution. Prior 0.3 evidence is segregated under `docs/history/0.3.0-alpha.1` and is not counted as new tests.

The existing manager integration used the updated source before the final downloader log-redaction refinement; final unit and signed bootstrap/migration tests used the final source/binaries. This did not change the UI/profile code. The report does not falsely claim every matrix combination was rerun on every operating system.

## What the network and transaction tests exercise

The Go network tests use actual localhost TLS servers with test-only trusted certificates. Production code retains normal certificate validation, HTTPS, approved host restrictions, bounded metadata/download sizes, and no plaintext override. Cases include primary HTTP failure; metadata/payload mirror fallback; interrupted transfer; valid and invalid Content-Range; a server ignoring Range; a corrupt prefix; corrupt primary bytes; repeated attempts; wrong size/hash/signature; stale and equivocal metadata; and diagnostic URL-token redaction.

The installer authenticates an installed-file index, not only an outer ZIP. Tests attempt to modify both a launcher file and its local receipt, reject a failed health check without changing the active pointer, preserve unrelated user data, recover from partial work, and prohibit unsigned package execution. Package publication tests reject commercial game assets, executables and nested/private payload patterns in launcher archives.

The final signed executable integration checks actual compiled launcher health, successful setup-to-manager HTTP handoff, a blocked concurrent updater, local base/patch imports, and a **931-file** verified community workspace. Applying a staged launcher package preserves every user-data file byte-for-byte. No file hash comparison substitutes for game execution.

## Coverage and qualifications

Go statement coverage: manager 36.8%; distribution library 71.3%; release command 21.9%. Bootstrap and build configuration show 0% in the Go unit coverage report because they are exercised by separate real-process Python integration tests; that integration is not merged into the unit coverage metric. Coverage percentages are not a security audit or correctness proof.

The launcher UI screenshots are offline fixture renders. The game-process lifecycle uses a deliberately labeled process test double. Migration and repair save tests use a synthetic sentinel, not a loaded MOO2 campaign. Windows/macOS binaries are unsigned/unnotarized, cross-compiled and unexecuted. The 0.4 work does not repeat or elevate the older isolated PRSL instruction probe into live-game evidence.

## Before public operation

Owner-controlled production keys; real endpoint/hostname configuration; Windows/macOS acceptance; native executable signing/notarization strategy; source and upstream redistribution licensing; real provider download/mirror exercises; reviewed maintained Go toolchain and pinned workflow dependencies; and actual MOO2 acceptance remain open. The development publication gate is intentionally fail-closed.

## Final archive acceptance

After creating the downloadable kits, 32 additional packaging checks passed: ZIP integrity and per-file hashes for all four targets; a single matching launcher payload per platform; no game/private-key/font file payloads; shell-script parsing; actual installation and verification from the final extracted Linux kit; and exact preservation of the supplied architectural handoff. Evidence: `manager/evidence/packaging-verification.json`. These checks do not imply Windows/macOS execution.
