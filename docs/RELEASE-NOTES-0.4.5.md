# MOO2-SGC 0.4.5

This release is driven by the first successful Windows 11 gameplay acceptance of 1.50.26 and the first multiplayer failure observed afterward.

## Fix: RKERNEL.COM is now a required network prerequisite

MOO2 1.50.26 terminated when Network Game was selected because an older managed community environment did not contain `RKERNEL.COM`. 0.4.5 verifies the exact LAN-fixed 1.40b23 kernel before launching any 1.40b23 or 1.50.26 profile. Existing environments missing it are refused before MOO2 starts and must be repaired/rebuilt. The repaired copy is constructed from the verified owned 1.40b23 baseline; Steam/GOG and the source ZIP remain untouched.

Expected LAN-fixed kernel SHA-256:

`18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97`

## Network service selection

The launcher now separates the MOO2 in-game role from the IPX transport:

- **Direct / LAN** — Create starts a local DOSBox IPX server; Join connects to the supplied host/address and UDP port.
- **moo2.thedopefish.com** — both Create and Join connect to the long-running third-party public IPX service on UDP 213. Players then use MOO2's Multiplayer → Network screens to create or join a named game.
- **MOO2-SGC Online** — visible but disabled. This reserves the UI/data-model boundary for a future first-party lobby/matchmaking/relay service without coupling today's launcher to an unfinished backend.

The Dopefish endpoint is third-party infrastructure and availability is not controlled by MOO2-SGC.

## Scope

PRSL and the new Chat extension remain unavailable. The existing 1.40b23 baseline, 1.50.26 optional community environment, portable DOSBox Staging runtime, signed updates, rollback, and source isolation remain unchanged.
