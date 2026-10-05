# Distribution architecture — 0.4.2

Small native Setup → signed GitHub manifest → signed platform launcher ZIP → immutable application generation → local launcher UI → owned game import + independent pinned community patch/runtime → profile-specific game generation.

The Setup executable contains its public trust configuration. Its default operation is online; offline installation requires an explicit flag. The launcher package includes an identical, signed-indexed Setup helper for update restart. The helper waits for the application lock; it never removes a live lock or patches an executable in place.

User data is outside application versions. Game reconstruction preserves same-engine saves and retains old environments. Removing a game environment deactivates it rather than deleting files. PRSL and Chat are disabled, independent capabilities; no engine offsets or PRSL protocol are added to the bootstrapper.

The public signed release manifest remains schema 1, so installed 0.4.1 clients can authenticate 0.4.2 using the unchanged key. Private source snapshots and newly prepared environment receipts use schema 2; existing schema-1 profiles/environments are read for forward migration. Supported CD 1.2, Steam 1.40b23 and legacy 1.31 fingerprints coexist. See SOURCE-AND-LINEAGE.md for the distinct game installation recipes. Detection yields candidates, not certification of all retail editions. No commercial-game package is published.

The exact live-game interception, strict turn timers, zero-config LAN discovery, internet relay, arbitrary external mod import and all-version compatibility remain outside this release's implemented scope.
