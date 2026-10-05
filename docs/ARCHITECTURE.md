# Application architecture — 0.4.2

The distribution design is specified in [DISTRIBUTION-ARCHITECTURE.md](DISTRIBUTION-ARCHITECTURE.md); trust and rollback details are in [SECURITY-AND-TRUST.md](SECURITY-AND-TRUST.md).

The full manager remains a Go executable with embedded HTML, CSS and JavaScript, served on a random localhost port. It uses a per-process bearer token and an exact Host/origin policy. No publicly bound dashboard, account system, telemetry, or external web assets are enabled. The game process is launched with argument arrays rather than assembled shell commands. Local authenticated dashboard URLs are excluded from exported diagnostics.

A bootstrap installation gives the manager an immutable executable location and a stable `userdata` directory. The bootstrap and running manager share an application lock; each manager also owns its data-directory lock. Active executables are not replaced. Stale locks require an operator to confirm that no related process remains before removing them. A browser tab closing is not equivalent to the manager process exiting; use Exit launcher.

## Existing game environment contract

Owned sources are verified against per-edition DOS executable and asset fingerprints. The Steam 1.40b23 baseline and CD 1.2 lineage are specified in [SOURCE-AND-LINEAGE.md](SOURCE-AND-LINEAGE.md). Fresh stages are created without importing personal saves/store wrappers; known user files are retained during same-engine reconstruction. Compiled fingerprints and explicit installation steps are recorded in schema-2 receipts, and the active-generation pointer changes only after verification. Original files and previous generations remain untouched.

The descriptor catalog has 25 selectable upstream entries; Core and other exclusive groups cannot be combined arbitrarily. Dependency selection is structural, not certification that every combination has been played. Arbitrary external mod imports and independently fetched versions of each mod are not implemented.

PRSL and the future Chat extension have separate unavailable controls. Neither can inject code or start a coordinator. Native in-game chat is unchanged. The prior PRSL instruction research and unresolved live-game adapter remain independent of this release.

## Updates

SGC launcher updates now use shared Ed25519-signed manifests/packages, approved-host HTTPS transport, ranged resume, GitHub/R2 mirror failover, content-addressed caching and staged generations. The default metadata source is the configured public GitHub Release feed; an explicitly selected local signed release is also supported. Explicit check/stage is available from the manager; applying it requires a shutdown and Setup's apply-staged operation.

The existing upstream community-patch and DOSBox fetch paths still use exact pinned hashes and official sources. They are not secretly repackaged as signed SGC components. Downloadable mod/package resolution beyond the launcher remains incomplete.

## Limits

Only Linux executables were run. DOSBox installation from actual upstream packages, MOO2 gameplay, cross-machine multiplayer, native Windows/macOS behavior and live PRSL are not tested. LAN discovery, room codes, NAT traversal, relay operation and automatic firewall changes are absent. Save preservation tests use sentinels and byte comparisons; they do not prove every real saved campaign loads under a different engine or mod set. No cross-engine save migration is promised.
