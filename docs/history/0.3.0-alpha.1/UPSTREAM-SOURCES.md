# Upstream references and provenance

The game and patch payloads are copies of the user's supplied archives. Their exact digest pins are in the manager source and SOURCE-PAYLOADS.json. No commercial game content is downloaded by the manager.

Primary documentation consulted for this cycle:

- MOO2 Book, Installation: https://moo2mod.com/doc/dist/installation.html
- MOO2 Book, Modding: https://moo2mod.com/doc/150/modding.html
- DOSBox Staging Windows downloads / checksums: https://www.dosbox-staging.org/download/windows/
- DOSBox Staging Linux downloads / checksums: https://www.dosbox-staging.org/download/linux/
- DOSBox Staging macOS downloads / checksums: https://www.dosbox-staging.org/download/macos/
- DOSBox Staging IPX/multiplayer documentation: https://www.dosbox-staging.org/0.83/manual/networking/multiplayer/

Recorded runtime recipes (not downloaded/executed here):

| Platform | Artifact | SHA-256 |
|---|---|---|
| Windows x64 | dosbox-staging-windows-x64-v0.83.0.zip | 725b915e325a6d410ce30a10989fd492fdad07f6611fff40ff77f160478810f4 |
| Linux x64 | dosbox-staging-linux-x86_64-v0.83.0.tar.xz | d3a94f7f1c3e68a47ec88d61145506c7904452adb0c9c5928cb8cfe2331d6c5c |
| macOS universal | dosbox-staging-macOS-v0.83.0.dmg | d8a771adfb8010fa6b5f7fb5351abfba659273ad01c89f03675a92bdbdae8167 |

Runtime downloads go to official GitHub release hosting and approved release-asset redirects; the bytes are checked against these pins before extraction. They do not follow arbitrary URLs from the upstream game homepage. A successful checksum establishes identity against this recorded artifact, not game compatibility or an independent security audit.
