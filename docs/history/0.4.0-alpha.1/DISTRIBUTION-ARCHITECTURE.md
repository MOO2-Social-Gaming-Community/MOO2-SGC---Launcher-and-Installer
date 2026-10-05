# Distribution architecture — 0.4.0-alpha.1

## Authority and scope

`requirements/DISTRIBUTION-HANDOFF.md` is the user's supplied architectural authority. This implementation follows that document's provider hierarchy, bootstrap/launcher boundary, optional PRSL policy, commercial-data boundary, and no-repository-sprawl direction. This report distinguishes working local software from infrastructure that has not been deployed.

| Decision | Integration |
|---|---|
| Neocities Free is the public website only | No bootstrap, metadata, package, or runtime dependency on Neocities. Website code is not modified in this cycle. |
| GitHub is canonical | GitHub-labeled primary endpoints/mirrors in trust and signed package metadata; release tooling defaults to `MOO2-SGC/launcher`. No organization or repository was created here. |
| Dedicated MOO2-SGC Cloudflare account; R2 secondary | Optional HTTPS mirror plus best-effort S3 replication tooling. No other project account, credential, billing relationship, or bucket is assumed. |
| SourceForge deferred | Not an accepted provider in schema 1; no integration added. |
| Launcher ecosystem in one repository | Bootstrap, environment manager, downloader, verifier, updater, importer, configuration resolver, and release tooling share one source tree. |
| PRSL separate repository/component | No PRSL runtime code is included in the bootstrapper. The existing GUI keeps PRSL and Chat disabled independently. Earlier PRSL research remains a separate handoff. |
| No public commercial base-game distribution | All new kits omit game data. Local archive import verifies the known baseline without uploading it. |

## Runtime boundaries

```
setup executable
    -> embedded public trust configuration
    -> signed manifest (GitHub primary / optional R2 fallback)
    -> separately signed launcher package + signed installed-file index
    -> content-addressed cache
    -> staged immutable launcher generation
    -> --version health check
    -> atomic current-generation pointer
    -> full launcher with stable --app-root and --data paths
          -> local user-owned game source
          -> existing community patch/runtime paths
          -> profiles, rulesets, optional supported add-ons
          -> game preparation / verification / launch
```

The offline acceptance kit uses the same signature, file verification, staging, health check and activation path. Only the source of bytes changes from HTTPS to an explicitly supplied local folder. It does NOT claim that production URLs already exist.

## Source map

| Location | Responsibility |
|---|---|
| `manager/cmd/bootstrapper` | Minimal setup, install/repair/verify/rollback, launcher handoff, apply-staged operation. |
| `manager/shared/distribution/manifest.go` | Schema, strict JSON, signatures, feed/channel scoping, expiry and rollback protections. |
| `manager/shared/distribution/download.go` | HTTPS transport, mirror ordering, range resume, exact-byte verification, bounded retries. |
| `manager/shared/distribution/install.go` | Immutable generations, signed receipts/file verification, activation and rollback. |
| `manager/shared/distribution/fs.go` | Safe paths, bounded ZIP extraction, atomic writes, process exclusion. |
| `manager/shared/buildconfig` | Build-time embedded public keys/endpoints; unconfigured by default. |
| `manager/distribution.go` | Local game/patch archive importer and launcher update check/stage integration. |
| Existing manager files | Game workspace generation, patch overlay, mod selection, local UI, runtime and game launch. |
| `packaging` | Cross-platform builds, signatures, modular archives, explicit publication and mirror automation. |

## Stable paths and compatibility

Bootstrap installations separate application versions from user data:

```
MOO2-SGC/
    components/launcher/current.json
    components/launcher/versions/g-<hash>-<nonce>/
        receipt.json
        files/<launcher + notices>
    cache/packages/<sha256>.zip
    distribution/trusted-state.json
    distribution/pending/
    userdata/                    # existing manager data schema 1
        profiles/
        active/
        environments/
        cache/                   # owned game/patch imports
        logs/
    logs/bootstrap.log
```

The old portable `data/` layout is still accepted with `--data`. There is no forced migration or deletion. Installing a new bootstrap-managed copy defaults to a separate directory; it does not silently discover or merge an old portable installation. Game saves and application rollback have different boundaries. Rolling back a launcher does not roll back game worlds.

## Updates and recovery

1. Authenticate metadata using an embedded key. Validate feed/channel, schema, lifetime, revision and package descriptors.
2. Persist the newest authenticated revision even if the subsequent download fails. Lower revisions cannot quietly reappear via another mirror.
3. Select the native launcher package, fetch verified bytes into a content-addressed cache, and keep incomplete transfers as `.part` files.
4. Stage a new generation; never overwrite a running launcher executable.
5. Verify the extracted files against the SIGNED per-file index and run `--version` with a timeout.
6. Atomically update `current.json`; retain the previous generation.
7. Launch the selected immutable executable with a separate stable data directory.

The running manager can check/stage updates. Applying a staged update is explicitly restart-assisted: exit the game and manager, then run Setup with `--command apply-staged`. There is no background polling, silent self-restart, or active-game replacement.

## Intentionally incomplete

Production endpoints, persistent owner-controlled signing keys, actual GitHub/R2 deployment, full TUF role/key rotation, native platform acceptance, general Steam/GOG folder recognition, arbitrary downloadable mod-package installation, runtime repack licensing, and a live PRSL adapter are NOT completed by this release.

`requires` and `engine_hashes` are signed metadata fields for future component resolution; the current bootstrap installs only a launcher. The manager continues using its existing supported community-mod resolver. A schema field is not evidence that all future component types are already installable.
