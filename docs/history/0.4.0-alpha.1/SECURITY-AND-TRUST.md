# Security and trust contract — distribution schema 1

## Exact signature format

Use the provided `releasectl`, rather than guessing JSON serialization rules.

- Manifest detached signature: Ed25519 over ASCII `MOO2-SGC-MANIFEST-V1`, a NUL byte, and the **exact manifest.json bytes**, including whitespace and its final newline.
- Package signature: Ed25519 over ASCII `MOO2-SGC-PACKAGE-V1`, a NUL byte, and Go `encoding/json.Marshal` of the `PackageMessage` record in source. This record contains ID, Kind, Version, Platform, Filename, Format, Entrypoint, SHA256, Size and Files in that order. Files is a sorted-key JSON map containing `sha256` and `size` records. This is a domain-separated signature over the package content digest and execution/file-layout identity, not an Ed25519 signature over the full ZIP byte stream.
- The complete package ZIP is SHA-256 hashed. Exact size is also required. The signed manifest additionally covers mirror URLs, priorities, dependencies, supported engine hashes and redistribution designation.
- Detached signature JSON is `{ "key_id": "...", "signature": "<base64>" }`. Public trust keys are hex-encoded 32-byte Ed25519 public keys embedded at build time.

The installer rechecks extracted files against the authenticated file index, not merely a locally editable receipt. Receipt/hash tampering tests are included. Downloaded JSON cannot introduce its own trusted root.

## Implemented checks

HTTPS only; exact allowed hostnames; redirect validation; bounded downloads and expansion; platform/version/package identity checks; strict JSON with duplicate-key and nesting limits; manifest schema/feed/channel checks; maximum 93-day metadata lifetime; expiry; persistent revision high-water mark; same-revision/different-content refusal; package size/hash/signature checks; traversal, symlink, reserved-name and case-collision refusal; per-user staging; application/game-data separation; immutable version directories; health check; explicit rollback; exclusion lock shared by Setup and the running new launcher.

Go TLS test servers use test-only certificates in test clients. Production certificate verification is never disabled. Development trust may explicitly permit a nonstandard HTTPS port on a loopback address for tests; production trust does not. There is no HTTP exception, no unsigned update mode, and no web UI for replacing signing keys.

Expected HTTP failures, interrupted transfers, invalid ranges, content mismatches, and corrupted primary mirrors are tested. Query-bearing transport URLs are excluded from new distribution diagnostic messages. Older upstream/runtime download logging remains a separate legacy path; inspect exported diagnostics before posting publicly.

## Important limitations

This is NOT a TUF implementation or an independent security audit. It lacks threshold roles, delegated signing, automated root rotation/revocation, external transparency, secure monotonic hardware state, and a finished key-loss recovery procedure. A compromised trusted release-signing key can authorize malicious software. The signed metadata lifetime limits replay after expiration but cannot discover a newer release that an attacker conceals while older metadata remains valid.

The trusted-state file is ordinary per-user local storage. Another process running as the same user can delete state or replace the launcher itself; these protections do not defend against an already compromised OS account. Expired metadata cannot authorize a new install. Previously authenticated installed software can still be launched offline after expiration; that is an explicit availability policy.

Files are synchronized before atomic replacement, and failed process-level staging leaves the old pointer unchanged. This is not a certification against every filesystem, disk-full, sudden-power-loss, or Windows filesystem behavior. The exclusive lock is intentionally not silently cleared after a crash. Check that no manager/game/setup process is alive before manually removing a stale lock.

The `--version` health check verifies the signed expected version and process startup, not GUI usability or gameplay. Cross-platform compilation is not native OS acceptance.

## Development versus production

The delivered offline kit uses an explicitly marked **development-only signing identity**. A disposable private key was generated for signing and removed from its temporary directory after building; it is not in any deliverable. The public key is embedded in Setup and launcher binaries. These test kits do not establish an official MOO2-SGC publisher identity and are not a production trust root to reuse.

Production requires owner-controlled keys, an approved public trust configuration, account/role separation, recoverable offline key backups, reviewed release approval, and an update/revocation plan. This source does not silently generate production keys or select an existing Data Pioneer account.

Windows Authenticode and macOS signing/notarization are separate from update-package signatures. Neither has been performed here. Do not disable endpoint protection or run as administrator to bypass an unexpected warning.

## References consulted

- https://pkg.go.dev/crypto/ed25519
- https://theupdateframework.io/security/
- https://docs.github.com/en/actions/reference/security/secure-use

The design borrows threat categories from TUF guidance without claiming TUF conformance.
