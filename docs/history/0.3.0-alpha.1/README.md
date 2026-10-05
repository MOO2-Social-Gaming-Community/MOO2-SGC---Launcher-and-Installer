# MOO2 Mod Manager — source and evidence

**0.3.0-alpha.1, private testing release.**

This source package contains no original MOO2 executables, LBX assets, commercial game ZIP, fan-patch executable, DOSBox runtime, or signing keys. It contains the new standalone Go environment manager and web UI, generated upstream descriptor metadata, tests, documentation, evidence, and the earlier PRSL Lab 0.2.0 as a separate preserved research component.

## Build and unit tests

Go 1.23 or later; no external Go modules. Build environment used Go 1.23.2 on Linux x64. Choose an appropriately maintained toolchain and audit before a public release.

```sh
cd manager
go test -race -cover -v ./...
go vet ./...
go build -trimpath -o moo2-manager .
```

For a Windows build from another host:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o MOO2-Mod-Manager.exe .
```

`build.sh` builds Windows x64, Linux x64, Intel Mac and Apple Silicon Mac artifacts. Cross-compilation is not runtime acceptance.

To run with the user's separate private payload bundle:

```sh
./moo2-manager --root '/path/to/extracted/private/bundle' --data '/path/to/writable/test-data'
```

CLI operations support `--command resolve`, `--command prepare`, and `--command verify`, with `--profile community|multiplayer|original`. Initial startup creates three default profiles in the chosen data directory. No PRSL-capable launch mode exists.

## Owned-fixture integration tests

Run only against a disposable data directory. This test intentionally corrupts an executable in a retained test generation, writes a fake save sentinel, and uses a explicitly named process TEST DOUBLE. It never runs MOO2. The private source archives are not modified.

The test uses Python, Playwright and a Chromium executable for offline DOM rendering. It tests actual HTTP APIs separately. Browser fixture coverage is not end-to-end browser-network coverage.

```sh
export MOO2_PRIVATE_ROOT='/path/to/extracted/private/bundle'
export MOO2_MANAGER_EXE="$PWD/moo2-manager"
export MOO2_TEST_DATA="$(mktemp -d)"
export MOO2_CHROMIUM='/usr/bin/chromium'
python integration_test.py
```

Evidence is written beside the script under `evidence`. The user's input archive paths are not required in the container's original form; the private bundle provides `payloads/base.zip` and `payloads/patch-1.50.26.zip`. They must have the exact pinned source hashes.

The recorded test run used the development paths shown in its logs. These are provenance, not recommended user installation paths.

## Regenerate the upstream descriptor catalog

`build_catalog.py` reads (but does not execute) the supplied upstream ZIP. Set `MOO2_PATCH_ZIP` to the exact 1.50.26 archive. It extracts descriptors and enabled dependencies into `assets/catalog.json`. Generation is not a runtime plugin importer. Review the output before compiling a new package recipe.

## PRSL is separate

`prsl-research/lab-0.2.0` preserves the prior coordinator, exact-build mapper, laboratory x86 gate, and reports. Its tests were rerun; the live-game adapter is still absent. The new manager does not import or execute that research package.

Read `docs/TEST-REPORT.md` for exact evidence boundaries, `docs/ARCHITECTURE.md` for implementation decisions, and `requirements/REQUIREMENTS-v0.1.0.md` for accepted product direction. This release is not a declaration that the broader requirements are complete.
