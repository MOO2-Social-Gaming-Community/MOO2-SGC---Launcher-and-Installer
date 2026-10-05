# MOO2-SGC Distribution Infrastructure & Bootstrap Architecture Handoff

## Status

This document records the current infrastructure and distribution decisions for the **Master of Orion II Social Gaming Community (MOO2-SGC)** software ecosystem.

Treat these as the current architectural requirements when developing the installer, bootstrapper, launcher, updater, patcher, configurator, mod manager, PRSL integration, and future community-mod support.

The goal is to create a **modular, resilient, mostly free, low-maintenance distribution system** that can survive beyond the active involvement of its original developer/maintainer.

---

# 1. Core Architectural Principle

Do **not** build MOO2-SGC as one giant monolithic installer that must be redownloaded every time something changes.

Instead use:

```text
Small Bootstrap Installer
        │
        ▼
MOO2-SGC Launcher
        │
        ▼
Signed Distribution Manifest
        │
        ├── Core components
        ├── Patches
        ├── PRSL
        ├── Other mods
        ├── Configuration resources
        └── Future community components

```

The initial download should be as small and self-contained as practical.

Its purpose is to:

1. bootstrap the installation;
2. retrieve the current launcher;
3. verify what it downloads;
4. install/update the launcher;
5. hand control to the full launcher.

The **launcher then becomes the long-term orchestrator** for the rest of the system.

---

# 2. Infrastructure Providers

The initial production infrastructure will use **three services only**.

## Neocities Free

Purpose:

- Public MOO2-SGC website
- Community-facing front end
- Documentation
- Download page
- Community information
- Systems/local-group information
- News/changelog links
- Links to GitHub releases
- Links to alternative downloads if needed

Important:

**Neocities Free is NOT part of the binary-distribution backend.**

Do not depend on Neocities for `.exe`, ZIP/package hosting, launcher updates, or critical manifests.

The site should remain a portable static website.

Its canonical source should also be stored in GitHub so it can be redeployed elsewhere if Neocities ever disappears.

---

## GitHub

GitHub is the **canonical technical backend and primary source of truth**.

The user's existing **Data Pioneer** GitHub account will administer a new:

```text
MOO2-SGC
GitHub Organization

```

Expected initial repositories:

```text
MOO2-SGC/
├── website
├── launcher
└── prsl

```

Additional substantial mods may receive their own repositories later.

Do NOT split the launcher ecosystem unnecessarily into separate repositories for:

```text
bootstrapper
installer
updater
patcher
configurator
mod-manager

```

For now these should remain components/modules of the main:

```text
MOO2-SGC/launcher

```

repository unless a genuine engineering reason later requires separation.

GitHub will be used for:

- canonical source code;
- release history;
- version tags;
- issue tracking;
- GitHub Releases;
- primary bootstrap downloads;
- primary component/package downloads;
- canonical manifests;
- checksums;
- signatures;
- source for automated mirroring.

GitHub should be considered the **authoritative release source**.

---

## Cloudflare

Cloudflare will use a **completely separate MOO2-SGC Cloudflare account**.

It will NOT share the user's Data Pioneer / Pioneer DataWorks Cloudflare account.

The separation is deliberate to minimize cross-project operational, administrative, legal, billing, or security risk.

Initial Cloudflare purpose:

```text
Cloudflare R2
└── secondary release/package mirror

```

Possible later uses:

- custom download domain;
- DNS;
- CDN;
- Workers;
- Pages;
- other MOO2-SGC infrastructure.

However:

**MOO2-SGC must not require Cloudflare to function.**

Cloudflare is a secondary/failover distribution source.

If Cloudflare disappears or stops being free, GitHub must remain sufficient for normal operation.

---

# 3. Deferred Infrastructure

## SourceForge

SourceForge is appealing because it fits the project's old-school software-community aesthetic and provides software-distribution infrastructure.

However it is **explicitly deferred**.

Do not design the current release process around SourceForge.

It may later become:

- archival distribution;
- additional mirror;
- mature public release channel.

Reason for deferral:

The project should not require updating or maintaining many different services every time a release is made.

Initial infrastructure should stay limited to:

```text
Neocities
GitHub
Cloudflare

```

---

# 4. Distribution Hierarchy

The initial hierarchy should be:

```text
PRIMARY
GitHub Releases

SECONDARY / FAILOVER
Cloudflare R2

```

For software payloads.

For the human-facing website:

```text
PRIMARY WEBSITE
Neocities Free

CANONICAL WEBSITE SOURCE
GitHub

```

Conceptually:

```text
                     MOO2-SGC
                         │
             ┌───────────┴───────────┐
             │                       │
          WEBSITE                 SOFTWARE
             │                       │
        Neocities Free              GitHub
             │                       │
             │                 Primary Releases
             │                       │
             │                       ▼
             │                  Cloudflare R2
             │                  Backup Mirror
             │
             └── website source also stored on GitHub

```

---

# 5. Bootstrap Installer

The initial download should be called something similar to:

```text
MOO2-SGC-Setup.exe

```

Internally it should be treated as a **bootstrapper**, not as the complete application.

Target size:

Keep it reasonably small, but reliability is more important than chasing an arbitrary minimum size.

A self-contained bootstrapper in the approximate range of:

```text
5–30 MB

```

would be excellent.

Even a somewhat larger bootstrapper is acceptable if required to eliminate fragile external runtime dependencies.

Do NOT sacrifice robustness just to reach a very small binary size.

---

# 6. Bootstrapper Responsibilities

The bootstrapper should perform only the minimum reliable setup necessary.

Required responsibilities:

```text
1. Establish HTTPS connection.
2. Retrieve distribution metadata.
3. Verify the manifest cryptographically.
4. Determine current launcher release.
5. Select an available mirror.
6. Download the launcher.
7. Support resumable/ranged downloads where available.
8. Verify byte count.
9. Verify SHA-256.
10. Verify package signature.
11. Stage installation safely.
12. Install/update the launcher.
13. Roll back safely if installation fails.
14. Launch the full MOO2-SGC Launcher.
15. Produce useful logs/error information.

```

The bootstrapper should contain enough functionality to recover from:

- GitHub temporary failure;
- Cloudflare temporary failure;
- interrupted downloads;
- corrupted files;
- incomplete installs;
- incompatible existing launcher installations.

It should **never blindly execute downloaded content merely because the HTTP request succeeded**.

---

# 7. Full Launcher Responsibilities

Once installed, the full MOO2-SGC Launcher becomes the orchestrator.

It should eventually manage:

```text
MOO2-SGC Launcher
│
├── Game detection
├── Existing-install verification
├── Version selection
├── Patch installation
├── Patch removal/rollback
├── Mod installation
├── Mod enabling/disabling
├── Mod version selection
├── PRSL management
├── Future community mods
├── Configuration profiles
├── Repair
├── Updates
├── Dependency resolution
├── Compatibility checking
├── Downloads
├── Hash/signature verification
└── Game launch

```

The launcher must preserve the ability to play:

- vanilla MOO2;
- supported patched versions;
- patched version without PRSL;
- patched version with PRSL;
- other mod combinations where compatible.

PRSL is a **separate mod/component**, not something permanently fused into the launcher.

---

# 8. Modular Package Architecture

Avoid a monolithic:

```text
MOO2-SGC-Everything.zip

```

whenever possible.

Prefer modular components such as:

```text
bootstrap/
launcher/
patches/
mods/
configuration/
resources/
documentation/

```

Example:

```text
launcher/
    launcher-0.1.0.zip

patches/
    moo2-15026.zip

mods/
    prsl-0.1.0.zip
    future-chat-mod.zip

resources/
    optional-resource-pack.zip

```

Benefits:

- small updates;
- less bandwidth;
- easier rollback;
- easier repair;
- easier version pinning;
- independent mod evolution;
- easier mirrors;
- better diagnostics.

If PRSL changes by 800 KB, users should not need to redownload hundreds of megabytes.

---

# 9. Manifest Architecture

The launcher/bootstrapper should use a machine-readable release manifest.

Suggested concept:

```json
{
  "schema": 1,
  "release": "0.1.0",

  "packages": {
    "launcher": {
      "version": "0.1.0",
      "size": 12345678,
      "sha256": "...",
      "signature": "...",
      "mirrors": [
        {
          "provider": "github",
          "priority": 10,
          "url": "..."
        },
        {
          "provider": "cloudflare-r2",
          "priority": 20,
          "url": "..."
        }
      ]
    }
  }
}

```

The exact schema may evolve.

Design it to be **versioned and backward-compatible**.

Include a manifest schema version from the beginning.

---

# 10. Mirror Selection

Initial mirror priority:

```text
1. GitHub
2. Cloudflare R2

```

Behavior:

```text
Try GitHub
    │
    ├── success → download
    │
    └── failure
           ↓
Try Cloudflare
    │
    ├── success → download
    │
    └── failure
           ↓
Return meaningful recovery/error state

```

The downloader should distinguish between:

- HTTP failure;
- timeout;
- DNS failure;
- incomplete transfer;
- corrupted file;
- hash mismatch;
- signature mismatch;
- unsupported package;
- insufficient disk space;
- permission failure.

A mirror that recently failed can temporarily be deprioritized during the current update session.

---

# 11. Manifest Redundancy

Do not make GitHub the only place from which update metadata can ever be retrieved.

Eventually mirror the signed manifest itself:

```text
GitHub
    manifest.json
    manifest.sig

Cloudflare
    manifest.json
    manifest.sig

```

The bootstrapper should contain trusted bootstrap endpoints.

It should also contain or otherwise securely obtain the **public key used to verify manifest signatures**.

Trust must come from cryptographic verification, not simply from trusting whichever URL returns JSON.

---

# 12. File Integrity and Supply-Chain Security

Every downloadable package should ultimately have:

```text
filename
version
size
SHA-256
cryptographic signature

```

Recommended verification sequence:

```text
DOWNLOAD
    ↓
verify byte size
    ↓
verify SHA-256
    ↓
verify signature
    ↓
stage
    ↓
install
    ↓
verify installed state
    ↓
commit

```

Never install a failed or unverified package.

Prefer staging + atomic replacement over modifying live installations directly.

Keep rollback data where practical.

---

# 13. Windows Code Signing

Eventually the Windows bootstrapper and launcher should be code-signed.

This is not required to complete the earliest development builds, but the architecture should anticipate production signing.

This helps with:

- Windows SmartScreen;
- publisher identity;
- tamper detection;
- community trust.

Do not confuse Windows executable signing with package/update signing.

Both have separate purposes.

---

# 14. Release Automation

The maintainer should **publish once**, not manually update several hosting services.

Desired future workflow:

```text
Create Release
     │
     ▼
Build components
     │
     ▼
Run tests
     │
     ▼
Generate hashes/signatures
     │
     ▼
Generate manifest
     │
     ▼
Publish GitHub Release
     │
     ▼
Automatically mirror identical artifacts to Cloudflare R2
     │
     ▼
Verify uploaded copies
     │
     ▼
Publish signed manifest

```

GitHub is the source of truth.

Cloudflare should be treated as an automated replica/mirror rather than an independently maintained release repository.

If Cloudflare mirroring fails, the GitHub release should still remain usable.

---

# 15. Website Deployment

The current MOO2-SGC website is on **Neocities Free**.

The GitHub repository:

```text
MOO2-SGC/website

```

should contain the canonical website source.

The deployed Neocities site should therefore be reproducible from GitHub.

If Neocities ever disappears, the site should be portable to:

- GitHub Pages;
- Cloudflare Pages;
- another static host;
- a future custom-domain host.

Neocities should never contain unique project data that exists nowhere else.

---

# 16. Future Custom Domain

A custom domain may eventually be purchased.

Example concept:

```text
moo2-sgc.org

```

or equivalent.

A custom domain should be treated as a **portable identity layer**, not as a technical dependency.

Possible future endpoint:

```text
downloads.<domain>

```

could point to Cloudflare or another provider.

However the underlying GitHub URLs should continue to function independently.

The system should remain usable even if the domain expires.

---

# 17. Sustainability Requirement

A major project goal is:

> MOO2-SGC should be capable of surviving without the founder continually paying recurring hosting bills.

Therefore:

- prioritize free/open infrastructure;
- avoid unnecessary SaaS dependencies;
- avoid proprietary control planes where possible;
- preserve all source and manifests;
- keep deployment reproducible;
- make mirrors replaceable;
- prevent any single nonessential provider from becoming mandatory.

Initial infrastructure should operate without required recurring hosting costs while usage remains within free-service limits.

---

# 18. Account/Ownership Structure

## GitHub

Existing personal GitHub identity:

```text
Data Pioneer

```

will create/administer:

```text
MOO2-SGC GitHub Organization

```

Repositories belong to the organization rather than directly to the personal account.

This enables future community maintainers without exposing unrelated personal repositories.

---

## Cloudflare

Cloudflare will use a **new, dedicated MOO2-SGC account** created with the MOO2-SGC community identity/email.

It should be completely separated from:

```text
Data Pioneer
Pioneer DataWorks

```

This is intentional.

Do not assume access to unrelated Cloudflare infrastructure.

---

# 19. Repository Structure

Initial repositories:

```text
MOO2-SGC/
│
├── website
│
├── launcher
│
└── prsl

```

Potential future repositories:

```text
chat-mod
other-independent-mod
.github
documentation
development-tools

```

but only create them when they become materially useful.

Avoid repository sprawl.

Within `launcher`, favor modular internal structure:

```text
launcher/
│
├── bootstrapper/
├── launcher/
├── updater/
├── patcher/
├── configurator/
├── mod-manager/
├── shared/
├── manifests/
├── packaging/
├── tests/
└── docs/

```

Exact names may evolve.

---

# 20. SourceForge

Do not implement SourceForge integration yet.

It may be added later once:

- the project is more mature;
- licensing is settled;
- release automation exists;
- there is actual value from another public distribution network.

SourceForge is a **future option, not a current dependency**.

---

# 21. Commercial Base-Game Boundary

The public MOO2-SGC distribution infrastructure should only distribute files for which MOO2-SGC has appropriate redistribution rights.

The original commercial Master of Orion II game data should remain conceptually separate unless redistribution rights are established.

Therefore the launcher should be designed to:

```text
Detect existing MOO2 installation
        │
        ├── Steam
        ├── GOG
        ├── other legitimate installation
        └── manually selected installation

```

and then manage patches/mods around that installation.

Do not architect the public release system on the assumption that the original commercial game files will be redistributed.

---

# 22. Expected Initial User Experience

The public user flow should eventually resemble:

```text
User visits MOO2-SGC Neocities website
                │
                ▼
       DOWNLOAD MOO2-SGC
                │
                ▼
        GitHub Releases
                │
                ▼
       MOO2-SGC-Setup.exe
                │
                ▼
          Bootstrapper
                │
                ▼
      Signed release manifest
                │
        GitHub / Cloudflare
                │
                ▼
      Install current Launcher
                │
                ▼
        Launcher starts
                │
                ▼
      Detect MOO2 installation
                │
                ▼
      Select version / patches
                │
                ▼
       Select desired mods
                │
                ▼
        Download components
                │
                ▼
        Verify everything
                │
                ▼
          Configure game
                │
                ▼
            PLAY MOO2

```

---

# 23. Development Priorities From Here

When integrating this infrastructure into the current executable-development work, prioritize:

1. **Stable bootstrapper/launcher boundary**
2. **Versioned manifest schema**
3. **Downloader abstraction**
4. **GitHub mirror/provider implementation**
5. **Cloudflare mirror/provider implementation**
6. **SHA-256 verification**
7. **Cryptographic manifest/package signatures**
8. **Transactional/staged installation**
9. **Rollback**
10. **Logging**
11. **Installer repair mode**
12. **Launcher self-update**
13. **Mod/component dependency management**
14. **Release automation**

Do not couple PRSL implementation directly to the bootstrapper.

Do not couple any one hosting provider deeply into launcher business logic.

Use abstractions such as:

```text
IReleaseSource
IMirrorProvider
IPackageDownloader
IManifestProvider
IPackageVerifier
IInstaller

```

or equivalent appropriate abstractions for the chosen implementation language.

Providers should be replaceable.

---

# 24. Architectural Requirement

The final system should tolerate scenarios such as:

```text
GitHub available
Cloudflare unavailable
→ system works

```

```text
GitHub temporarily unavailable
Cloudflare available
→ system can recover/fail over

```

```text
Neocities disappears
→ software distribution still works
→ website can be redeployed from GitHub

```

```text
Cloudflare pricing changes
→ disable/remove Cloudflare mirror
→ GitHub continues operating

```

```text
future maintainer takes over
→ repositories and infrastructure can be transferred/administered
→ no dependence on unrelated Data Pioneer infrastructure

```

That resilience is a deliberate project requirement.

---

# 25. Current Infrastructure Decision — Freeze for Initial Development

For the present development cycle, treat this as the settled architecture:

```text
WEBSITE
Neocities Free

CANONICAL WEBSITE SOURCE
GitHub

CANONICAL DEVELOPMENT HOME
GitHub MOO2-SGC Organization

PRIMARY SOFTWARE DISTRIBUTION
GitHub Releases

SECONDARY SOFTWARE MIRROR
Cloudflare R2

FUTURE OPTIONAL DISTRIBUTION
SourceForge

BOOTSTRAP MODEL
Small self-contained MOO2-SGC-Setup.exe
        ↓
downloads/verifies/installs current launcher
        ↓
launcher orchestrates everything else

```

Do not add additional hosting providers at this stage unless a concrete technical limitation requires one.

The immediate engineering objective is therefore:

> **Refactor/build the distribution architecture so that MOO2-SGC begins with a small bootstrap installer, while the full launcher independently orchestrates installation, updating, patching, configuration, mod management, verification, rollback, and mirror-aware downloading of the remaining components from GitHub Releases with Cloudflare R2 failover.**