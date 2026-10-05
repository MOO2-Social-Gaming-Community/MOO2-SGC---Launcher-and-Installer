# Architecture and implementation decision

## Boundaries preserved

The manager is not the PRSL mod. The independent PRSL research code remains unchanged in the source handoff and was regression-tested again. No generated binary gate, experimental memory API, readiness coordinator or chat process runs in this release.

The current source is a native Go manager with an embedded HTML/CSS/JavaScript dashboard served on a random loopback port. This implementation was selected for the testing release because it compiles to Windows, Linux and macOS executables without requiring users to install Python, Node, Tcl or a webview runtime. The earlier Python lab remains a separate research module; its speculative gate is not translated or installed here. This is a delivery implementation decision, not a change to PRSL's behavior contract.

The browser is a local UI, not a hosted service. It uses a random per-process bearer token and an exact Host/origin policy. No public listening interface, remote assets, account system, telemetry, or external web chat is enabled. The game process is launched directly with argument arrays; no shell commands are assembled from profile fields. The local loopback URL is never included in diagnostic exports.

## Package model in this alpha

- Base: exact supplied archive, SHA-256 pinned; never modified.
- Engine: original 1.31 or exact supplied community 1.50.26.
- Community content: descriptors extracted from the exact patch, preserving IDs, exclusive groups, order and direct/recursive CFG-enabled dependencies. Twenty-five selectable descriptors remain after excluding the internal USER descriptor.
- Features: PRSL and new Chat are unavailable; the resolver and launcher fail closed for their IDs.
- Runtime: existing user-selected DOSBox, or an official 0.83.0 download whose SHA-256 is pinned in source.
- State: profile request, resolution, managed file records, effective configuration fingerprint and generation pointer.

This is NOT a generic plugin SDK or an arbitrary mod importer. Bundle versions are tied to the patch distribution until per-mod package recipes and compatibility records exist. Known exclusive-group conflicts are rejected; structural resolution does not certify gameplay compatibility.

## Workspace lifecycle

1. Resolve and validate the requested profile.
2. Verify source archive SHA-256 hashes.
3. Extract to an operation-owned staging directory, rejecting unsafe names, symlinks, duplicate/case aliases, overwrite attempts and oversized archives.
4. Omit archived user-state seeds from fresh installs.
5. Apply the community overlay only for the 1.50.26 engine selection.
6. Generate ENABLE.CFG for the resolved selection, leaving native root machinery intact.
7. Preserve known user files from the active generation only when the engine version is unchanged.
8. Hash the generated workspace and verify the exact native engine hash.
9. Rename the staged generation into the retained generation set.
10. Atomically replace the active-generation pointer. Older generations remain untouched.

The separate active pointer and requested-profile file are not a multi-file crash transaction. If a crash occurs between those writes, launch still compares the requested resolution with the prepared resolution and refuses a mismatch; it does not guess. Windows file replacement uses MoveFileExW with replace/write-through flags. Only Linux behavior was executed here.

## Mutability policy

Known user files include saves, race/score/settings files, DIG.INI, MDI.INI, SOUND.LBX, 150/USER.CFG and build-list CFGs. Other CFG/LUA/EXE/LBX content is not broadly treated as mutable. User CFG contents participate in the effective compatibility fingerprint. User-provided configuration is not interpreted as harmless merely because it is editable; users must agree on it before a multiplayer test.

Repair preserves the current known user files, including custom settings. Therefore it is not a universal reset of those files. Older generations remain available for recovery. The app does not automatically clean them up or migrate saves between engine versions.

## Updating

Implemented: explicit upstream-version discovery, explicit download/refresh of the pinned 1.50.26 package, runtime fetching, digest verification before cache replacement, bounded downloads, HTTPS host/redirect restrictions, and reconstruction-based activation.

Not implemented: a production signed release feed, hosted PRSL adapters, automatic activation of unknown versions, scheduling, manager binary self-replacement, key rotation, or a complete secure-update framework. The older Python laboratory includes signature-validation experiments, but those are NOT wired into this manager and must not be advertised as production updater coverage.

## Known operational limits

- Windows and Mac binaries were cross-compiled, not executed.
- Actual official DOSBox runtime archives could not be downloaded here. ZIP, Linux tar and Mac DMG packaging assumptions need real-machine validation.
- A process-start acknowledgement does not mean MOO2 loaded successfully or IPX connected.
- LAN discovery, invitation codes, cross-machine fingerprint negotiation, NAT traversal, relays and firewall management are absent.
- Configs are generated using standard DOSBox IPXNET commands. The native game's menus still create/join the multiplayer game.
- Every generated environment retains its own saves. Loading a save can restore the game's saved configuration.
- A same-user malicious process is not an isolation boundary. This alpha must not be installed or run as a privileged service.
- Disk-space cleanup and robust interrupted-download resume are future work. Failed or partial downloads are not activated.

## Next engine milestone

Retain the earlier exact-build research: Human_Hit_Next_Turn_ is a getter, not submission. The candidate gate is at the manual event boundary, and release needs a fresh caller frame. It remains unvalidated inside the live event loop; the alternate automatic path is uncovered. Nothing in this manager release changes that status.
