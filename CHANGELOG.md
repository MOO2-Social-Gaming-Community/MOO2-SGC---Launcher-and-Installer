# 0.4.2


Fixes the clean Steam import rejection encountered in 0.4.1. Recognizes the verified English Steam DOS 1.40b23 baseline and English CD 1.2 data, without modifying the source installation.

- Steam: 1.40b23 → 1.50.26.
- CD: 1.2 → official 1.31 → 1.40b23 → 1.50.26.
- Separate current, baseline, original CD and legacy profiles.
- Archive import based on known file contents, not ZIP filename/outer hash alone.
- Historical b23 file-transform output matches the supplied clean Steam executable exactly; no XP-era Windows patcher is run.
- Default managed sound setup avoids requiring MT-32 ROMs. Original CD uses an appropriate launch command.
- New schema-2 lineage receipts, compiled file verification, source preservation and same-engine save retention.
- Existing 0.4.1 signing identity and user-data location retained. Public release contains no full game payload or private key.

File/import/build and application tests ran on Linux. Windows/Mac binaries were cross-compiled; Windows game execution and two-client multiplayer remain user acceptance tests. PRSL and new Chat remain disabled. This update does not provide a matchmaking or relay service.

---

# Changelog

## 0.4.1

- Exact user repository wired into signed online bootstrap endpoints.
- Complete GitHub Desktop drop-in tree including reviewed release inputs.
- Automatic verify/upload/readback/publish workflow; repeat runs are safe and non-clobbering.
- A signed copy of Setup in each launcher package enables staged-update restart.
- Per-user Start-menu shortcut and visible Windows double-click errors.
- Explicit stable installation namespace preserves older development installations.
- Known-folder detection, fingerprint-verified local DOS-folder import and one-action preparation.
- Emulated C: data path and matching Sound Blaster IRQ configuration.
- Source-to-binary build-input fingerprint; metadata protected against Git line-ending normalization.
- Three-part version only; existing limitations are not hidden by the version label.

Previous narratives and test evidence are retained under docs/history and manager/evidence. They are not 0.4.1 test results.
