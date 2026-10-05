# MOO2-SGC 0.4.0-alpha.1: separate GitHub uploads

## 1. Source repository

Copy the **contents** of `repository/` into the root of your existing launcher
repository. The root must contain `README.md`, `manager/`, `packaging/`, `docs/`,
`manifests/`, `requirements/`, `tests/`, and `.github/`. Do not put everything inside
an extra `repository/` directory on GitHub, and do not upload the source ZIP as the
only repository file.

Preserve hidden `.github/`, `.gitignore`, and `.gitattributes` files. The source
snapshot includes prior test evidence and requirement handoffs. Review those
before making a private repository public. No new project license was selected.
The existing manual release workflow still requires the owner's signing and
hosting configuration; this packaging step does not configure Actions or secrets.

## 2. Release downloads, separate from source

The flat files in `release-assets/` are for **Release assets**, not the source tree.
Create a draft tagged `v0.4.0-alpha.1`, with title
`MOO2-SGC 0.4.0-alpha.1 — offline installation preview`. Paste `RELEASE-NOTES.md` into
the description, mark it as a prerelease, attach the separate assets, and save a
draft while you review. Do not mark this as a stable/latest public release.

A draft/private release is not an unauthenticated public download endpoint. The
current client has no authenticated GitHub-release download flow. Manual download
for this local test is separate from testing the future online bootstrap.

GitHub creates source archives from a repository tag automatically. Those source
archives are **not** substitutes for the signed `launcher-*.zip` assets.

## 3. Minimum Windows test files

Keep these five files in the SAME local folder:

```text
MOO2-SGC-Setup.exe
launcher-0.4.0-alpha.1-windows-amd64.zip
manifest.json
manifest.sig
START-SETUP-WINDOWS.cmd
```

Run `START-SETUP-WINDOWS.cmd`. The script installs into
`%LOCALAPPDATA%\MOO2-SGC-Upload-Test`, leaving the earlier standard installation
alone. Repeat installation uses the same isolated test location.

Optional integrity check in PowerShell, after downloading all assets:

```powershell
Get-FileHash .\MOO2-SGC-Setup.exe -Algorithm SHA256
```

Compare the result with `SHA256SUMS.txt` obtained from this handoff. Checksums are
not a substitute for the setup's embedded signature trust. Keep signed JSON and
ZIPs byte-for-byte intact: do not unzip/repack the launcher package, format the
JSON, rename the package, or mix metadata from a separately generated test kit.

## 4. Linux and Mac

Use the matching setup executable, launcher ZIP, and helper with the same two
manifest files. Terminal invocation avoids requiring the helper itself to retain
an executable bit after download:

```sh
sh ./START-SETUP-LINUX.sh
# Intel Mac:
sh ./START-SETUP-MAC-INTEL.command
# Apple Silicon Mac:
sh ./START-SETUP-MAC-APPLE-SILICON.command
```

The scripts check the native platform/architecture. On an Apple Silicon Mac,
use a native arm64 Terminal session and the Apple Silicon helper rather than
running the Intel helper under translation. They do not disable Gatekeeper or
otherwise bypass operating-system security.

## 5. Test the launcher, then the game

When the launcher opens, import the exact previously supplied `base.zip` and
`patch-1.50.26.zip` from the private testing bundle, or the original matching
archives. These are local files; **do not upload them to the GitHub repository or
Release assets**. Select your existing DOSBox runtime, prepare a community
profile, verify it, then test title screen/new game/save/reload before networking.
PRSL and Chat stay unavailable in this build.

## 6. What is missing for the intended small ONLINE installer

The supplied alpha has **no GitHub metadata endpoints and no package mirrors**.
These are valid separate files for a manual-download/offline installation test;
they are not an online-ready release. The exact created repository URL was not
provided in this conversation and has not been assumed to be `MOO2-SGC/launcher`.

For a real bootstrap test, configure the exact repository/release URLs and an
owner-controlled signing identity, rebuild the setup/launcher with that public
trust, and sign metadata containing real package URLs. Keep the private key
outside the repository and outside every uploaded artifact. Editing the current
manifest or `trust-public.json` alone cannot do this; the current disposable
signing key was not retained in the release.

Use `docs/DEPLOYMENT.md` for that next build. The future bootstrap will verify and
download the launcher from GitHub; R2 is optional. Neocities is not involved in
the download or test. Pin an explicit alpha-release endpoint rather than assuming
GitHub's latest-stable shortcut will discover this prerelease.

## Verification

the separate packaging verification report records this pass's checks. Prior test
reports remain historical reports; this pass does not claim to rerun the entire
game/mod-manager regression suite or any actual game session.

## Documentation used for GitHub operations

Checked during this packaging pass:
- https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository
- https://docs.github.com/en/repositories/working-with-files/managing-files/adding-a-file-to-a-repository
- https://docs.github.com/en/repositories/releasing-projects-on-github/linking-to-releases
