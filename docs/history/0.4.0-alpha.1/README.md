# MOO2-SGC Launcher — 0.4.0-alpha.1

**Modular distribution development release.** The small bootstrap installer and full launcher are now separate programs with a shared, signed distribution module. This source tree follows `requirements/DISTRIBUTION-HANDOFF.md`. It is the intended content of the single `MOO2-SGC/launcher` repository, not evidence that the remote repository already exists.

## What works

A compiled setup verifies a signed manifest and launcher package, stages an immutable generation, verifies its installed files, executes a bounded launcher `--version` health check, atomically activates it, and opens the full launcher. Installation, repair, verified rollback and next-start application of staged updates are implemented. User data lives outside application generations.

The full launcher retains original 1.31/community 1.50.26 selection, the existing supported community-mod resolver and local browser UI, generation repair, profiles, save preservation and DOSBox launch controls. It adds exact-archive local game/patch import and explicit signed launcher update check/stage controls. PRSL and Chat remain independently unavailable; no experimental game hooks run.

GitHub is the primary distribution provider; Cloudflare R2 is optional failover. Neocities is only the human-facing website. SourceForge is deferred. No cloud accounts, remote repositories, buckets, production keys or live feed were configured by this work.

## Start with the matching platform kit

The separate offline acceptance kits include only a setup executable, a signed launcher ZIP, public trust metadata and documentation. Run their `START-SETUP` script, which passes the local `release/` directory explicitly. Do not mix packages from different development builds. These kits use a disposable development key, not a production community identity.

There are no game files or DOSBox runtime in these kits. Import the exact user-owned base archive and 1.50.26 patch from the prior private bundle or original uploads. Generic Steam/GOG folder recognition remains future work.

## Source map

`manager/` contains the full Go application and embedded UI. `manager/cmd/bootstrapper` and `manager/cmd/releasectl` are the setup and release-signing commands. `manager/shared/distribution` contains trust, metadata, download and transactional-install code. `packaging/` holds reproducible build and explicitly gated publication scripts. `tests/` contains real executable acceptance tests. `manifests/` documents schema 1. Prior PRSL research remains a separate handoff, not a dependency of this repository.

## Build and test

Local build toolchain used: Go 1.23.2 on Linux x64; no external Go modules. The source requires Go 1.23+. Choose and pin an appropriately reviewed, maintained toolchain for public builds rather than treating the local laboratory toolchain as a production recommendation.

```sh
cd manager
go test -race -count=1 ./...
go vet ./...
go build -trimpath -o moo2-sgc-launcher .
cd ..
python3 packaging/build_release.py --development --revision 1 --out /absolute/path/to/output
```

The last command compiles Windows x64, Linux x64, Intel Mac and Apple Silicon Mac setup/launcher pairs; creates separately signed launcher ZIPs; and destroys its temporary private key. It embeds only public trust. A plain `go build` produces an unconfigured distribution client and must not be advertised as a working online setup.

A development release cannot be published using `packaging/publish_release.py --execute`. Production deployment needs owner-controlled keys and genuine configured endpoints; see `docs/DEPLOYMENT.md`. Signing keys never belong in this source tree, a release ZIP, an issue, or a support log.

## Reproduce executable acceptance tests

Use disposable output/data directories. No test below executes MOO2. Real game reconstruction uses the exact original archives and synthetic save sentinels.

```sh
python3 tests/bootstrap_integration.py --release /absolute/path/to/output \
  --evidence manager/evidence \
  --base '/path/to/Master of Orion 2.zip' --patch '/path/to/MOO2-1.50.26.zip'

# Existing manager UI/API/profile tests, with private payloads/base.zip and patch-1.50.26.zip:
cd manager
MOO2_PRIVATE_ROOT='/path/to/private-fixture' \
MOO2_TEST_DATA='/path/to/disposable/data' \
MOO2_MANAGER_EXE='/path/to/compiled/linux/launcher' \
python3 integration_test.py
```

The UI test requires Python, Playwright and Chromium. It renders offline response fixtures; the actual loopback API is tested separately. It is not a live browser-to-backend end-to-end test. Network failure tests use real local TLS test servers, not an actual GitHub outage or live R2.

Read `docs/TEST-REPORT.md` for measured results and limitations; `docs/SECURITY-AND-TRUST.md` for the threat model; `docs/DISTRIBUTION-ARCHITECTURE.md` for implementation boundaries. Historical 0.3 documents under `docs/history` are retained for provenance only and are superseded by current documents.

## GitHub upload preparation

See [GITHUB-UPLOAD.md](GITHUB-UPLOAD.md) for source-versus-release placement and
the offline-only status of the preserved 0.4.0-alpha.1 binaries.
