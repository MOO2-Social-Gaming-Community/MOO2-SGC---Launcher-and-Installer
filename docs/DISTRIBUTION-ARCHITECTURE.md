# Distribution architecture — 0.4.1

Small native Setup → signed GitHub manifest → signed platform launcher ZIP → immutable application generation → local launcher UI → owned game import + independent pinned community patch/runtime → profile-specific game generation.

The Setup executable contains its public trust configuration. Its default operation is online; offline installation requires an explicit flag. The launcher package includes an identical, signed-indexed Setup helper for update restart. The helper waits for the application lock; it never removes a live lock or patches an executable in place.

User data is outside application versions. Game reconstruction preserves same-engine saves and retains old environments. Removing a game environment deactivates it rather than deleting files. PRSL and Chat are disabled, independent capabilities; no engine offsets or PRSL protocol are added to the bootstrapper.

The release manifest schema is unchanged at 1. Folder-source snapshots introduce a separate local schema-1 record, not a public commercial-game package. The initial trusted source ZIP and new 408-file folder fingerprint path coexist. Detection yields candidates, not certification of all retail editions.

The exact live-game interception, strict turn timers, zero-config LAN discovery, internet relay, arbitrary external mod import and all-version compatibility remain outside this release's implemented scope.
