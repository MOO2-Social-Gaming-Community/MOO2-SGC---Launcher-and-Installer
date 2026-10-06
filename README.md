# MOO2-SGC Launcher and Installer — 0.4.6

**Portable baseline first. Optional community upgrades second.**

This release integrates the supplied **Portable Harness 0.3.0-HANDOFF** into the existing signed installer, updater and modular environment manager. The owner has demonstrated the underlying 1.40b23 DOSBox path on Windows 11. The new manager integration still needs Windows acceptance; file checks and Linux process fixtures are not a gameplay certification.

## 0.4.6 fixes

Setup will not install or launch 0.4.2 as the fallback for a new release. It rejects stale latest metadata, tries the version-specific endpoint, rechecks the actual executable, and displays exact version/root identity. The release publisher verifies the latest pointer.

The manager verifies RKERNEL.COM in the exact executable directory before every baseline/current launch. Complete source preparation supplies it automatically; an explicit owned RKERNEL.COM/rkernel.zip importer also supports staged repair of old incomplete snapshots. Source files stay untouched, and same-engine saved bytes survive repair. See [recovery instructions](START-HERE.md) and [investigation](docs/UPGRADE-AND-KERNEL-0.4.6.md).

## Canonical Windows working directory

```text
C:\Games\MOO2-SGC\
    Master of Orion 2 - v1_40b23.zip   # private owner-supplied source
    runtime\windows\dosbox.exe      # existing verified harness distribution
    game\ORION2.EXE                  # normalized writable 1.40b23 baseline
    config\moo2.conf                 # tested window/audio contract
    state\                          # baseline receipts and transaction journal
    backups\baseline\               # complete retained baseline snapshots
    components\launcher\            # signed application generations
    userdata\                       # profiles, cache, optional game environments
```

Do not put this private working directory inside your GitHub clone. The public repository and release contain no commercial game data, emulator, fonts or private signing keys. The portable integration ZIP contains **project code only** and reuses the runtime from the owner's supplied harness.

The manually patched source is recognized by file hashes, not its name. It retains official 1.31 in `ORION2.EXE` and b23 in `Orion2v140.exe`. Preparation preserves 1.31 as `ORION131.EXE`, copies the verified b23 executable to the managed `ORION2.EXE`, and applies only the documented exact-hash RKERNEL LAN correction. The source ZIP is never modified.

## First use

Keep a separate copy of your currently working harness. Extract the **0.4.6 portable integration ZIP** into `C:\Games\MOO2-SGC`, with no extra nested directory. It adds `START-MOO2-SGC.cmd` and separate `SGC-*.cmd` helpers; it does not replace the handoff's PowerShell preparation scripts or runtime.

Keep the owned baseline ZIP in that same root and the complete original harness runtime in `runtime\windows`. Open **START-MOO2-SGC.cmd**. On first use, it installs the signed launcher from the kit's offline package. Subsequent starts verify and open the installed application without reinstalling an older release.

The default profile is **Portable baseline — 1.40b23**. Leave the source input blank and select **Prepare game for play**. After verification, choose **Launch MOO2**. Baseline launch uses the original harness flags and direct program argument; the default is an 800×600 window that does not pause or mute when inactive.

First Windows acceptance: title/version, sound/video/input, new game, several turns, save, exit, restart and reload. Then test repair/save preservation. Only afterward choose **Community — standard** to construct a separate 1.50.26 environment.

## Optional versions and mods

1.40b23 is the default runtime, not a hidden intermediate that is immediately overwritten by 1.50.26. The named portable baseline profile is fixed to b23. Duplicate it or select another profile to use a different engine or rule set. Historical CD 1.2 → 1.31 → b23 construction remains available, but is not repeated when importing the approved b23 source.

Community 1.50.26 and its bundled rulesets/add-ons remain supported in independent environments. PRSL and the proposed Chat extension remain separately disabled. Selectable IPX transport includes: Direct/LAN or the third-party `moo2.thedopefish.com` public IPX service. A disabled MOO2-SGC Online entry reserves the future first-party matchmaking/relay boundary; no first-party lobby backend is deployed yet.

## Publish

Extract the complete **repository ZIP** directly into your existing clone, preserving `.git`. Commit and push with GitHub Desktop. The prepared files in `release/0.4.6/` are signed using the existing owner-controlled signing identity. The release workflow publishes `v0.4.6`; no new signing secret is required for these already-signed files. Never edit signed assets in place.

The separate standalone setup now defaults to `C:\Games\MOO2-SGC` on Windows. Run it after the release is published. The offline integration kit can be tested before publication. An update initiated from an old `%APPDATA%` installation remains there: **0.4.6 does not silently move, delete, or merge old user data.**

See [START-HERE](START-HERE.md), [portable design](docs/PORTABLE-HARNESS.md), [Windows acceptance](docs/TEST-CHECKLIST.md), [test report](docs/TEST-REPORT.md), and [release notes](docs/RELEASE-NOTES-0.4.6.md).
