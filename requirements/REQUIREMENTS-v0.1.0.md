# Modular MOO2 Launcher — Requirements Addendum v0.1.0

**Status:** Accepted product direction from the user's latest request; proposed implementation contract. This document does not add a graphical interface or implement a general mod resolver. The software remains PRSL Lab v0.2.0.

**Basis:** Inspection of `prsl/cli.py`, `prsl/packages.py`, and `README.md` in the supplied `moo2_prsl_lab_v0.2.0.zip`, plus the community's public modding and installation documentation. No existing source package was modified.

## 1. Product boundary

The launcher is a neutral MOO2 environment and mod manager. PRSL is one independently versioned, optional mod. The proposed Chat extension is another independently versioned, optional mod. Neither the launcher nor unrelated mods may require PRSL merely because PRSL was developed first.

No new product or community name is adopted by this addendum. "Launcher" and "environment manager" are descriptive terms. Preserve the user's existing community name separately.

User requirements:

- Show an initial screen before game launch to choose a profile, engine/patch version, community ruleset, and optional mods.
- Permit ordinary community play with PRSL disabled.
- Permit other community mods in supported combinations, not only the project's own mods.
- Keep Chat separate from PRSL; additional features can become independent packages when that separation has a real purpose.
- Preserve clean owned-game sources and versioned, reconstructable workspaces.
- Target Windows, Linux, and macOS. Platform support must be tested, not inferred from a cross-platform language.

## 2. Architecture

```text
Launcher / environment manager
    |
    +-- owned base-game importer
    +-- package catalog and verified local cache
    +-- profile selection, dependency resolution, and lockfiles
    +-- workspace builder, verifier, updater, and launch orchestration
    |
    +-- selected engine / community patch
    +-- selected community Core ruleset and other exclusive groups
    +-- selected community add-ons
    +-- optional PRSL package
    +-- optional Chat package
    +-- other individually supported extensions
    |
    +-- optional shared integration services and exact-engine adapters
        (only when selected features actually require them)
```

Keep three responsibilities distinct:

1. **Manager:** installation, selection, verification, networking setup, and process lifecycle.
2. **Extension integration:** engine observation, UI attachment, message transport, and hook arbitration, where needed and validated.
3. **Features:** PRSL's reversible readiness/commit behavior, Chat's messaging interface, and unrelated community functionality.

A shared integration layer must not itself implement pre-ready behavior. PRSL controls pre-ready policy only when PRSL is selected. Chat must not request PRSL as a hidden dependency simply to obtain networking or UI services.

Do not build a broad plugin SDK before concrete needs justify it. Define narrow interfaces first; extract shared code when PRSL and Chat genuinely need the same capability. Shared engine hooks require a single verified dispatcher, not several independently applied patches to the same instruction.

## 3. Startup screen contract

Show the selection screen on normal startup. On first use, an owned-game import/setup step can precede it. Loading the last-used profile must not launch it without the user's explicit action.

Proposed fields:

| Control | Meaning |
|---|---|
| Profile selector | A named, reusable set of choices |
| Engine/patch selector | Exact installed version or explicitly selectable supported download |
| Core ruleset selector | One choice in an exclusive Core group |
| Other exclusive groups | At most one choice per group, according to the supported package model |
| Optional mod rows | Enable checkbox, version selector, author/source, status, and settings |
| Dependency preview | Packages required by selected choices and why |
| Compatibility summary | Missing dependencies, unsupported engine, conflicts, untested combinations |
| Actions | Save profile, duplicate profile, verify/repair, prepare, play/host/join |

Version selection must refer to real catalog entries. Do not invent a release number for a planned mod. A planned Chat package can be displayed as `Planned — unavailable`; the current PRSL research code must be displayed as unavailable for live-game activation.

There are two separate user actions:

- **Prepare/update environment:** review downloads, dependency changes, source/trust details, and build a workspace.
- **Play/host/join:** start an already checked environment, after resolving required multiplayer agreement.

Combining these into one guided flow is acceptable; silently changing a saved profile at launch is not.

## 4. Installed, selected, resolved, and running are different states

- **Installed:** package bytes exist in the local cache.
- **Selected:** the user explicitly enabled a package in a profile.
- **Resolved:** the manager derived the full compatible dependency set and exact versions.
- **Running:** a verified workspace containing that resolved set is active.

Unchecked packages must not contribute configuration includes, injected code, UI hooks, background feature processes, or packets to that game session. Being present in the cache does not make a package active.

An unselected dependency can still be active only if another selected mod requires it and the dependency preview explains that relationship. This cannot silently enable PRSL: selecting Chat must not turn on readiness semantics.

When disabling a package that another selected package requires, show the dependent packages and require an explicit choice to disable them too or cancel the change. Do not silently re-enable the unchecked feature.

## 5. Package model and versioning

Independent version streams:

- launcher application;
- community engine/patch;
- community rulesets and add-ons;
- PRSL;
- Chat;
- optional integration services;
- version-specific engine adapters;
- emulator/runtime per supported operating system and architecture.

A proposed package descriptor needs these concepts (schema not yet implemented):

| Field | Purpose |
|---|---|
| Stable package ID and display name | Avoid collisions and retain authorship |
| Version / upstream release label | Preserve identifiers such as historical suffixes; do not assume every mod uses SemVer |
| Package type and exclusive group | Engine, Core, map, add-on, runtime extension, adapter, etc. |
| Source and integrity | Provenance, verified digests, package size, redistribution terms |
| Supported engine builds | Exact hashes for binary hooks; explicit tested compatibility for config/data packages |
| Platform support | OS/architecture restrictions for native tools and adapters |
| Dependencies / provided capabilities | Required versions and why they are needed |
| Conflicts and ordering | Known incompatible choices, required load order, and explicit overrides |
| Declared resources | Config includes, assets, scripts, and binary-hook claims |
| Multiplayer policy | Session-required, negotiated subset, or validated local-only |
| Settings schema | Which options are local and which are session-wide |
| Save/environment compatibility | Relevant restore requirements and known migration restrictions |
| Validation status | Recognized, structurally checked, runtime-tested, multiplayer-tested, or unsupported |

Never equate a trusted download, successful dependency resolution, or valid checksum with gameplay compatibility. Those are different checks.

The manager must retain the exact resolved set, configuration digest, and relevant load order in a profile lockfile. Store a user-readable requested profile separately so that adding a dependency does not disguise the user's original selections.

## 6. Community mod interoperability

The community's existing 1.50 launcher already recognizes mod directories under `150/mods` and descriptors such as `mod_id`, `mod_name`, `mod_class`, and `mod_order`. Its documentation distinguishes independent checkbox add-ons from exclusive groups such as Core. Use that model rather than treating every package as a freely stackable checkbox.

Import supported descriptor information while preserving upstream IDs, authorship, ordering, settings, and assets. Store additional manager metadata alongside it, rather than inserting unknown launcher fields into upstream CFG files.

"No PRSL" does not mean "unpatched original game." With PRSL off, a user can still choose a community patch, one community Core ruleset, and supported add-ons. A stock-engine profile is a separate target and is not currently implemented by the v0.2.0 launcher path.

Old mods that replace executables or original assets need explicit recipes and compatibility investigation. They may require mutually exclusive environments. Do not promise arbitrary historical mod stacking or infer compatibility merely because two mods load.

Unknown local mods may be inventoried for review, but discovering them must not execute their scripts or installers. Native extensions and adapters need stronger trust and compatibility checks than metadata-only packages.

## 7. PRSL and Chat relationship

Required independent combinations, once implemented and tested:

| PRSL | Chat | Intended result |
|---|---|---|
| Off | Off | Ordinary selected community game; no custom feature code runs |
| On | Off | Pre-ready behavior and its required status/timer controls, without the new Chat extension |
| Off | On | Chat functionality without altered native End Turn semantics |
| On | On | Both features, with an optional integration between Chat UI and pre-ready status |

The original game's own chat is not removed when the proposed Chat extension is disabled.

Chat's actual implementation is unresolved. A companion/launcher chat tab and an in-game chat tab have different engine-integration requirements. Reusing the native turn-monitor chat requires verifying and separating any native submission side effects. This document does not assert that decoupling has been implemented.

If Chat wants to display pre-ready status, use optional capability detection. Hide that panel when PRSL is absent. Do not require PRSL just to exchange messages. Similarly, PRSL must present essential readiness and timer information without requiring Chat.

The PRSL v0.2.0 prototype includes coordinator and research work, not a certified live adapter. The Chat extension has not been built. Neither may appear as playable in the released selector until validated.

## 8. Multiplayer agreement

Optional selection is per profile before the session; it is not permission for every player to use conflicting session semantics.

For an initial PRSL-enabled session, require compatible PRSL and adapter builds for every participating human. All players agree on the required engine, gameplay content, and session settings. Do not silently downgrade PRSL or mix native and pre-ready submission policies.

A genuinely local-only UI preference need not force every client to match. That classification must be established by implementation and tests, not asserted because a feature sounds cosmetic.

Chat may be optional for a subset only if its protocol and host/session policy support that arrangement. Advertise availability; do not imply that players without Chat receive or can answer its messages.

When a joining player's profile differs, show a proposed temporary session profile and the precise changes before downloading/activating them. Preserve the player's personal profile. Pin resolved versions during the game and save the environment lockfile alongside campaign information without changing the native .GAM format merely for PRSL bookkeeping.

## 9. Update, removal, and rollback

Choosing a different mod version, disabling a mod, or removing a community patch creates a different desired environment. Reconstruct from verified sources instead of attempting arbitrary reverse patches.

The current v0.2.0 code only builds a fresh destination; reconstruction/switching/repair orchestration remains work to implement. Do not advertise those as completed operations.

Maintain separate display states for latest upstream, latest supported by this manager/adapter set, and version pinned to the selected profile. An upstream release must not automatically update a session into an incompatible adapter combination.

Retain old packages while installed profiles/campaigns reference them. A cache cleanup operation must distinguish unreferenced package removal from disabling a mod in one profile.

Never hot-remove engine hooks while MOO2 is running. Changes take effect on a fresh launch. Preserve source archives and back up relevant saves/preferences before migration or workspace replacement.

## 10. Code transition from v0.2.0

Observed present implementation:

- `prsl.cli launch` accepts `--mode community|prsl`.
- `prsl.packages.launch` refuses any mode other than community.
- `build` recognizes only the supplied base and 1.50.26 archive hashes.
- The manifest contains PRSL research metadata even for community workspaces.
- There is no general mod catalog, profile resolver, mod-picker GUI, or implemented Chat feature.

Proposed next refactor:

1. Introduce neutral profile/package models and a pure resolver, separate from feature code.
2. Add a startup selector that invokes that same resolver and displays real support status.
3. Move general environment-management implementation out of the PRSL feature namespace when practical.
4. Preserve the old `python -m prsl.cli` entry point as a compatibility shim during migration; retain its fail-closed behavior.
5. Move PRSL research/adapter metadata into its own optional package descriptor and research views.
6. Add supported community config/ruleset selection before attempting arbitrary legacy binary patches.
7. Connect the certified live PRSL adapter only after the pending in-game tests pass.

These are implementation requirements, not changes made by this addendum.

## 11. Acceptance tests to implement (not executed here)

| Test | Required outcome |
|---|---|
| Startup | Selector appears before MOO2 starts |
| Community profile | No PRSL hooks, coordinator, generated PRSL config, or PRSL packets |
| Installed but disabled | Cached PRSL/Chat bytes have no effect on the running session |
| PRSL-only / Chat-only / both / neither | Each supported combination behaves according to Section 7 |
| Core conflict | A second conflicting Core choice is rejected or explicitly replaces the first |
| Engine mismatch | An unsupported adapter cannot be enabled by editing UI/profile metadata |
| Missing dependency | Explanation and explicit resolution; no silent enablement |
| Shared hook conflict | Overlapping native hooks blocked without a verified composition mechanism |
| Version switch | A new exact environment is built without mutating the old one |
| Save preservation | Changes/removal preserve user saves and original sources |
| Multiplayer mismatch | Session-required differences are caught before play |
| Local-only differences | Validated harmless differences do not unnecessarily block joining |
| Game running | Selection changes do not modify the active workspace |
| Unsupported legacy mod | Clear unsupported status, not an attempted arbitrary overlay |
| Cross-platform | Real Windows, Linux, and Mac startup/selection/install/launch testing |

## 12. Evidence and source notes

Local implementation inspected: `moo2_prsl_lab_v0.2.0.zip`, specifically `README.md`, `prsl/cli.py`, and `prsl/packages.py`. No claims of new software execution or runtime integration are made by this document.

Primary upstream references consulted:

- The MOO2 Book / 1.50 / Modding — https://moo2mod.com/doc/150/modding.html — mod descriptors, checkbox versus exclusive classes, ordering, and configuration model.
- The MOO2 Book / Installation — https://moo2mod.com/doc/dist/installation.html — upstream installation workflow.

Product requirements and proposed architecture above are project decisions, not assertions that upstream exposes all of these capabilities.
