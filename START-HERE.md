# MOO2-SGC Launcher and Installer — 0.4.1

Master of Orion II Social Gaming Community's modular bootstrapper, environment manager, updater, configurator and community-mod launcher.

**Current scope:** owned DOS game import, community patch 1.50.26, selectable bundled rulesets/add-ons, DOSBox setup, verified environments, application updates and recovery. **PRSL and the proposed Chat extension remain separate and unavailable.** This version number is not a claim of game or platform certification.

## Maintainer: unzip, commit and push with GitHub Desktop

Extract this archive's **contents directly into your cloned repository root**. Preserve the existing `.git` directory. There is no extra enclosing repository folder in the ZIP.

The destination is:

`MOO2-Social-Gaming-Community/MOO2-SGC---Launcher-and-Installer`

Commit all supplied files, including `.github/`, `release/0.4.1/`, `.gitattributes` and `.gitignore`, then push to the repository's **default branch**. The included **Publish prepared release** workflow verifies the signed files, runs tests, creates a draft `v0.4.1` Release, attaches the individual downloads, downloads them back for verification, and publishes the Release. It refuses to replace different bytes in an already published version.

**The push intentionally triggers publication of the prepared software Release.** The public installer cannot access draft/private assets without credentials; none are embedded. The repository must be public for anonymous installation, and repository/organization settings must allow GitHub Actions and its requested `contents: write` permission.

No signing secret, personal access token, local Go installation or manual asset upload is required for publishing the **already signed 0.4.1** package. Future rebuilt versions require a maintainer signing key; see `docs/SIGNING.md`.

Check the **Actions** tab. When **Publish prepared release** succeeds, download `MOO2-SGC-Setup.exe` from the `v0.4.1` Release and run it from an otherwise empty folder. That tests the online bootstrap path, not an adjacent offline package.

## Player: install and prepare

1. Run `MOO2-SGC-Setup.exe` normally; administrator privileges are not required. It verifies the signed release, downloads and installs the platform launcher, and opens its local browser interface.
2. Choose the desired profile, engine and community ruleset/add-ons. For the first test, use **Community — standard**.
3. Under **Prepare your owned game**, select a recognized installed DOS folder or paste your earlier private bundle's `payloads/base.zip` path. Click **Prepare game for play**. It imports local game files, obtains the checksum-pinned community patch and DOSBox when missing, builds and verifies a separate environment.
4. Click **Launch MOO2**. First confirm a new single-player game can save and reload; then test multiplayer.

Import never uploads your owned game or alters the original installation. Folder matching currently checks the 408 file fingerprints from the supplied DOS 1.31 baseline; it does **not** claim support for every Steam/GOG/localized/modded edition. The previously supplied exact base ZIP remains supported. Unrecognized files stop import with a specific error.

The original commercial game is **not** included in this repository or downloaded from GitHub. The community patch and DOSBox come directly from their official upstream hosts, not from an unlicensed repackage.

## Updating and recovery

In **Launcher distribution**, choose **Check signed release**, then **Stage launcher update**, then **Apply staged update and restart**. The signed installed helper waits for the running launcher to exit and release its lock. Game sessions block update/rebuild operations. Application updates preserve profiles and saves.

Game patch updates are separately pinned. This version supports only original DOS 1.31 and community 1.50.26; observing a newer upstream version does not authorize an unsupported engine or PRSL adapter.

Windows setup creates a Start-menu **MOO2-SGC** shortcut when possible. It opens the verified installed launcher offline. Re-running the downloaded Setup checks the release feed. The default application root is `%APPDATA%\MOO2-SGC\stable`; this separates the new signing identity from earlier disposable development kits. Earlier installations are neither deleted nor automatically migrated.

Advanced setup commands: `--command verify`, `repair`, `rollback`, `launch-installed`, `apply-staged`. `--offline PATH` remains available for deliberately testing a matching signed release directory; it is **not** selected automatically.

## Important test boundaries

The development environment executes Linux programs. Windows/macOS binaries are cross-compiled; their native smoke tests are configured to run on GitHub after your push. No result from a not-yet-run GitHub workflow is claimed here. Read `docs/TEST-REPORT.md` for locally executed evidence.

The installer/launcher is unsigned by Windows Authenticode and the Mac files are not notarized. Package signatures are separate and do not suppress OS warnings. Do not disable antivirus or run as administrator to force a test through. The available local compiler is recorded in `release/0.4.1/BUILD-RESULTS.json`; use an up-to-date supported Go toolchain for subsequent owner-controlled production builds.

No live DOSBox/MOO2 multiplayer or PRSL gameplay certification is claimed. A version of **0.4.1** intentionally has no alpha/beta suffix while these limitations remain explicit.

## Layout

- `manager/`: launcher, bootstrapper, shared distribution code and tests.
- `packaging/`: signed build, input verification and publishing tools.
- `release/0.4.1/`: exact public artifacts for GitHub Desktop push-to-release.
- `manifests/trust.json`: public keys and the exact GitHub endpoints; no private key.
- `tests/`: native, HTTPS-fixture and publishing acceptance checks.
- `evidence/0.4.1/`: current test output. Older evidence is historical.
- `requirements/`: accepted architecture and modular-mod requirements.

GitHub Releases is primary. Optional Cloudflare R2 transport support remains, but no R2 endpoint/account has been configured in this build. Neocities is not a binary or updater dependency.
