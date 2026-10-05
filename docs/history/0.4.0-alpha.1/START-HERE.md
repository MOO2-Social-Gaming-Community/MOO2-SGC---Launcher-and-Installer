# MOO2-SGC 0.4.0-alpha.1 — start here

This is the small-bootstrap/modular-launcher testing release. **It is not yet connected to a production update feed.** The offline kit deliberately exercises cryptographic verification, installation and launcher handoff with its own signed release bytes, without pretending that the proposed GitHub/R2 endpoints are live.

## On your Windows computer

Extract the complete **WINDOWS_TEST_KIT** into a new writable folder. Keep the previous 0.3 private bundle; its `payloads` directory supplies your local game and patch archives.

Run **START-SETUP.cmd**, not the bare setup executable. It passes this kit's `release` folder explicitly. The bootstrap verifies metadata, the launcher archive, and installed files; runs a version health check; and opens the local browser dashboard. Its terminal stays open while the launcher is running.

In **Packages & updates**, import your original `Master of Orion 2.zip` or the prior private bundle's `payloads/base.zip`. Then choose the patch importer and import `MOO2-1.50.26.zip` or the prior bundle's `payloads/patch-1.50.26.zip`. These are verified local copies; nothing is uploaded and the originals are unchanged. This alpha recognizes those exact archives, not arbitrary installations or every Steam/GOG edition.

Under **Runtime**, select your existing DOSBox executable or explicitly use the pinned runtime download. Choose **Community — standard**, **Prepare / repair environment**, then **Verify files**. Expect `ok: true`. Only then try a new single-player game, save, exit and reload. PRSL and the new Chat extension remain independent, disabled features.

The default application root is `%APPDATA%\MOO2-SGC` on Windows, `$XDG_CONFIG_HOME/MOO2-SGC` (normally `~/.config/MOO2-SGC`) on Linux, and `~/Library/Application Support/MOO2-SGC` on macOS. Game/user state lives in its `userdata` subdirectory, outside launcher versions. This installation does not automatically merge or remove your older portable data.

No Python, Node, Go compiler, Steam login, cloud login or administrator installation is required to use the compiled manager. A browser and a compatible DOSBox runtime are still needed. OS permission/security warnings remain possible because Windows/macOS executable signing is not provided.

## Later starts, repair and staged updates

Use **OPEN-LAUNCHER** for normal later starts. It verifies the installed launcher and does not require current online metadata. **REPAIR-LAUNCHER** rebuilds the app using this kit. **VERIFY-LAUNCHER** checks it. **ROLLBACK-LAUNCHER** selects the retained verified previous version when one exists.

To exercise the update path, enter this kit's absolute `release` folder path under **Launcher distribution**. Check signed release, then Stage launcher update. Exit the game and click **Exit launcher**; run **APPLY-STAGED**. It will verify/apply the staged release and reopen the manager. Applying the same release tests staging and activation; it is not discovery of a newer version. The manager cannot replace itself while it owns the installation lock.

The kits use a disposable, explicitly labeled **development** public key. Do not mix setup binaries and manifests from different development builds; do not publish these as official community releases. The signing private key is not included and was discarded. Persistent maintainer-controlled production keys and endpoint deployment must precede a public update channel. New installations/updates respect the manifest expiry noted in PLATFORM.txt; already installed verified software may continue offline. Do not change your clock or disable validation to bypass expiry.

## Linux and Mac

Use the matching architecture's START-SETUP.sh or START-SETUP.command. Preserve executable permissions after extraction (for example, `chmod +x MOO2-SGC-Setup *.sh` on Linux). The Mac executables are not signed/notarized and have not been run on macOS here. This package does not disable platform security or claim every older Mac is supported.

## What has and has not been tested

The final compiled Linux setup and launcher passed signed install, repair, rollback, real loopback handoff, local archive import and staged-update tests. Existing and new Go tests, manager integration and legacy migration also passed; details are in docs/TEST-REPORT.md.

No DOSBox game was executed; Windows and Mac programs were compiled but not executed; live GitHub/R2 deployment and PRSL gameplay have not been tested. This is a distribution/environment-manager alpha, not a completed gameplay mod.

No commercial game data, DOSBox, private keys, cloud credentials or font files are included. If a step fails, preserve its exact output and use the dashboard's Export diagnostics when available. Do not force the application through security warnings by disabling antivirus or running it as administrator.
