# Supported sources and installation lineage — 0.4.2

## Source classification is not a filename guess

| Source fixture | Required copied files | DOS engine SHA-256 |
|---|---:|---|
| English CD 1.2 | 397 | `558c2bb51354fa48ba3189021374d67873ea0e7b37bab77e32f44eda89e135bc` |
| Supplied clean English Steam 1.40b23 | 405 | `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5` |
| Legacy English DOS 1.31 snapshot | 408 | `4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f` |

The current fan engine `ORION150.EXE` remains pinned to `2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c`.

The executable selects a candidate edition, then every required asset is verified against that edition's allowlist. Extra store files and source saved games are not copied. Imported ZIPs may have different names, root directory names, case and packaging, provided all required content matches. Ambiguous executable roots, case collisions, traversal, links, encrypted members and size-limit violations are refused. A known engine does not authorize an arbitrary modified data set.

Private snapshots are content-addressed ZIPs. New snapshots use stored members to avoid expensive recompression of already-large game assets. Installation streams only verified members into the fresh stage, hashing while copying. The source is never opened for writing. Cached source receipts are private local state, not authoritative declarations of a new supported edition.

## Installation paths

CD 1.2 can construct targets 1.2, 1.31, 1.40b23 or 1.50.26. Legacy 1.31 can construct 1.31, 1.40b23 or 1.50.26. Steam 1.40b23 can construct baseline 1.40b23 or current 1.50.26. Downgrades to missing original bytes are refused. The UI does not require the user to operate the intermediate 1.31 patch manually.

1.40b23 is the **project's effective compatibility baseline**, not the final official MicroProse patch. The archived b23 readme explicitly requires English DOS 1.31. The default community profile still executes 1.50.26. The installation lineage is recorded separately from the selected ruleset.

### Official 1.31

The shipped index describes 27 required DOS update files extracted from the owner's archived official update. Its Windows executable and readme are not installed. The fixed primary endpoint is `https://moo2mod.com/patch/MOO2-1.31.en.zip`.

Because publishers may repackage ZIPs, the importer authenticates **every installed output** using hashes embedded in the signed launcher instead of accepting an unknown executable based on HTTPS alone. File names, lengths and hashes must match. Only then is a normalized private cache accepted. Remote availability and content of the current live endpoint were not exercised by the file-based acceptance test; a repackaged archive was tested through a local TLS server. Local original-patch import remains available.

### 1.40b23

The old Windows patcher is not executed on a modern host. Its file-transform routine was inspected and reimplemented: fixed header adjustments, two small table insertions, a 32 KiB zero insertion, and 124 patch writes (3,057 literal bytes). Input and final output SHA-256 hashes are mandatory. The resulting 2,644,842-byte file was compared byte-for-byte with the user's clean Steam DOS executable and matches exactly.

The source contains the small patch-edit recipe, **not** the original or patched full game executable. The patcher itself also includes historical Windows/XP integration behavior which this implementation does not reproduce or execute. This does not change game rules beyond the historical b23 patch.

### 1.50.26

The existing checksum-pinned upstream distribution is added to the managed baseline, and selected compatible 1.50 rulesets are enabled as before. Baseline `ORION2.EXE` remains present; current profiles execute `ORION150.EXE`. Source assets are not arbitrarily rewritten to erase known CD/Steam differences. Matching engine bytes alone are not a multiplayer synchronization certification.

## Runtime configuration

The manager mounts only the managed game directory as DOS C:. Generated `orioncd.ini` points there. New environments receive SB16 digital-audio and SB Pro OPL music settings matching the emulator, avoiding a silent dependency on external MT-32 ROMs. Existing same-engine profile audio settings/saves are retained during rebuild. Original CD 1.2 is launched without assuming support for `/skipintro`.

## Compatibility and migration

0.4.2 reads old profiles, cached legacy 1.31 imports and schema-1 prepared environments. New builds record schema-2 source identity and exact lineage; verification uses compiled recipes rather than trusting a locally edited receipt. Mutable sound/settings files stay mutable; marking executable/data content as mutable in a receipt does not bypass compiled checks.

Old generations and user saves remain retained. New 0.4.2 profiles/receipts need 0.4.2 or later; rolling the application back to 0.4.1 does not teach it the new engines or receipt schema. Do not downgrade the application to operate new profiles. No save conversion across engine versions is performed.

PRSL's research adapter remains specific to its researched engine and is not made compatible with 1.40b23 simply by selecting a baseline. PRSL and new Chat are still disabled.
