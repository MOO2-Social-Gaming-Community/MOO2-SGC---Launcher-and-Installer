# Master of Orion II Social Gaming Community — Launcher & Installer

**Version 0.4.7. In development and trusted-user testing; not a general multiplayer release.**

A small signed bootstrapper installs the environment manager. The manager imports an owned game without changing the original Steam/GOG/CD source, preserves an independent **1.40b23 baseline**, prepares the optional **1.50.26** fan-patch profile, manages supported rulesets and configures DOSBox for local, LAN or public-IPX play.

## 0.4.7 focus

The launcher remembers the selected profile across restarts and changing local-browser ports. Historical engines no longer show a misleading 1.50 Core label. The network summary shows the exact selected game executable and transport command. **Check Dopefish connection (no game)** opens a separate CONNECT/STATUS diagnostic, with no mounted game and no fabricated success indicator. Existing kernel, signature, minimum-version and repair safeguards remain in place.

Portable Windows root: **C:\Games\MOO2-SGC**. See [START-HERE](START-HERE.md), [network and profile details](docs/NETWORK-AND-PROFILES-0.4.7.md), and [test report](docs/TEST-REPORT.md).

## Repository/publishing

Extract the complete repository package directly into the GitHub Desktop clone, preserve `.git`, commit and push the default branch. The **Publish prepared release** workflow verifies and publishes the already-signed `release/0.4.7` assets. Do not edit signed packages or metadata. GitHub is primary software hosting; optional R2 failover is not provisioned in this release. Website availability is not a software requirement.

The prepared offline portable integration kit can be tested before publication. The standalone setup's online installation needs the published Release, not merely committed source. [GitHub upload guide](GITHUB-UPLOAD.md).

## Boundaries

No commercial MOO2 files, DOSBox runtime, RKERNEL binary, or private signing key belongs in this repository/public release. Import your licensed files locally. PRSL and the new Chat extension are independent, unavailable future components; the future SGC matchmaking/relay service remains disabled. Selecting a profile does not rewrite the baseline or activate a patch automatically.

Windows and Mac builds are compiled, not locally executed by the build environment. The connection checker must be observed on your machine; it is not proof that Dopefish is reachable until its DOSBox status actually reports a connection. Native code signing/notarization is distinct from the implemented package signatures and is not supplied.
