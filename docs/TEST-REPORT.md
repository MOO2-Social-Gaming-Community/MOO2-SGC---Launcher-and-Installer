# Test report — MOO2-SGC 0.4.1

## Executed in this development cycle

| Category | Result | What ran |
|---|---:|---|
| Go tests with race detector | **79 top-level; 152 including subtests, all passed** | Manager/API/configuration, import checks, package signing, download failures, installed-file indexes and transactions. The totals overlap; do not add them. |
| Publishing safeguards | **7 passed** | Mocked `gh` command runner: draft creation, readback, repeat no-op, partial upload recovery and refusal to overwrite mismatched/published assets. |
| Native bootstrap/launcher smoke | **11 passed** | Actual final Linux executables: signed offline install, health, repeat setup, installed helper, real local HTTP API and orderly shutdown. |
| Online bootstrap/update fixture | **15 passed** | Actual final production-configured Linux executables through a local TLS CONNECT proxy; interrupted transfer, resume, approved-host redirect, tampered metadata/package refusal, actual check/stage/apply and launcher restart. |
| Owned folder / one-action preparation | **23 passed** | Actual final Linux launcher and the user's local owned game/patch archives. Imports 408 fingerprinted files, excludes prior saves, refuses changed data, prepares community rules, generates DOSBox configuration and preserves a synthetic save across rebuild. |
| Existing manager integration rerun | **39 passed** | Actual final Linux launcher, file repair/deactivation/reactivation, original vs patched environments, source preservation, diagnostics, local API access controls; Chromium interface exercised with offline DOM fixtures. |
| GitHub Desktop-style round trip | **10 passed** | Disposable local Git repository with `core.autocrlf=true`, all signed source/release bytes preserved, files correctly tracked, game/key sentinels ignored and release signatures verified after checkout. |
| Native builds | **Four targets compiled** | Windows x64, Linux x64, Intel Mac, Apple Silicon. Only Linux executed in this environment. |

`go vet`, JavaScript syntax, Python compilation, signed manifest/package verification and source-to-artifact correspondence also passed. These are checks, not additional gameplay tests.

## Evidence and reproduction

Current evidence is under `evidence/0.4.1/`. `release/0.4.1/BUILD-RESULTS.json` records compiler and artifact sizes; its `executed_here: false` fields describe the build step only. Later Linux execution is separately recorded by `linux-native-smoke.json`, `online-fixture.json`, `owned-game-acceptance.json` and `manager-integration.json`.

From the repository root:

```
cd manager
# Then return to the repository root for the Python commands below.
go test -race -count=1 ./...
go vet ./...
```

```
python -m unittest discover -s tests -p "test_publish*.py" -v
python packaging/publish_prebuilt.py
python tests/native_smoke.py --output local-native-smoke.json
python tests/repository_roundtrip.py --output local-git-roundtrip.json
```

The HTTPS fixture additionally needs Python's `cryptography` package and is Linux-only:

```
python tests/online_bootstrap_fixture.py --output local-https-fixture.json
```

Owned-game tests require locally supplied archives and an extracted native launcher:

```
python tests/owned_game_acceptance.py --base <OWNED-ZIP> --patch <PATCH-ZIP> --launcher <NATIVE-LAUNCHER> --output local-owned-acceptance.json
```

Never commit the supplied commercial fixture archives. The scripts do not fetch them or upload their contents.

## Bugs discovered and corrected during this cycle

Folder imports initially passed file validation but were rejected by the older workspace verifier, which recognized only the original archive hash. A typed folder-source identity now verifies actual installed files against the compiled 408-file baseline; changing both a file and its local receipt does not satisfy that check. Historical original-archive workspaces remain accepted.

The signed update helper initially lacked an explicit POSIX executable mode after package extraction; helper execution is now installed and tested. Repository checkout initially changed `go.mod` line endings under `core.autocrlf=true`; explicit source attributes now preserve all signed build-input hashes. Prepared environments now write their launch configuration before activation, normalize the archived host drive to emulated `C:\`, and set DOSBox sound IRQ 5 to match the fingerprinted source configuration.

These findings are the reason full-file, actual-binary and Git round-trip tests were run in addition to unit tests.

## What the network test does and does not prove

The final executable contains the user's exact GitHub repository URLs and non-development public signing key. For testing only, its child-process environment points HTTPS through a loopback CONNECT proxy and trusts a temporary local CA. Real HTTPS requests, redirects, range resumes, manifest/package signatures and real compiled application installation occur. No insecure mode or test certificate is embedded in the delivered application.

The fixture is **not a request to live GitHub**. Two interrupted transfers exercise failure without activation; a later request resumes and verifies the package. The update test stages the **same 0.4.1 version**, invokes the signed helper, waits for the old lock to be released, restarts a new actual launcher session and preserves a user-data marker. It is not proof of migration to an unreleased future version.

The `gh` publication tests are command-level simulations. No release, account, repository setting or cloud object was changed. Repository visibility, default-branch policy and Actions permissions could not be verified remotely here. The workflow still has to run successfully after the owner's push.

## Explicitly not tested

No Windows/macOS executable was run here. The repository configures native smoke jobs for Windows/macOS/Linux, but their future results are not counted. No Windows Start-menu link, SmartScreen dialog, macOS notarization, or native GUI acceptance is claimed.

No DOSBox binary was available in this environment; official DOSBox and patch network downloads were not exercised against their live hosts. The owned-file preparation test deliberately selects a labeled process test double instead of downloading an emulator. The process-launch regression test is also a test double. Thus no title screen, save loading, tactical combat, LAN game, Internet match or PRSL turn interception was executed. Save-preservation tests use a synthetic sentinel, not a validated MOO2 campaign.

The screenshot is a Linux browser render with offline API fixtures. Actual server behavior was tested separately over localhost. It is not a Windows screenshot or a claim of browser-to-server end-to-end navigation.

PRSL and the proposed new Chat mod remain disabled and cannot be enabled through this manager. Cloudflare R2 support is optional but no mirror is configured or tested live. The code has not received an independent security audit. The included compiler is the available Go 1.23.2; subsequent broader production builds should use a maintained patched toolchain. Windows Authenticode and macOS notarization are absent.

## Artifact handoff

Prepared files use version 0.4.1 without a prerelease suffix. Current metadata expires on 2027-01-03 UTC; future installs/updates need renewed signed metadata or a newer signed release after expiry. Installed verified applications can still launch offline. The private signer is deliberately outside every public artifact. The owner's first required acceptance test remains a clean Windows download, installation, owned-source preparation, new game, save, exit and reload.

Historic 0.3/0.4 files under `manager/evidence` or `docs/history` are not new 0.4.1 evidence. This report, together with `evidence/0.4.1`, is the current test record.
