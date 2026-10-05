# MOO2-SGC 0.4.2

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
