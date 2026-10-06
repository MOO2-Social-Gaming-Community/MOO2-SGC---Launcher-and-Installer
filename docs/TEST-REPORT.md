# MOO2-SGC 0.4.5 test report

## Trigger

Windows 11 acceptance proved 1.50.26 could launch and play, but selecting Network Game terminated with `RKERNEL.COM not found`. Fullscreen rendering was confirmed working. 0.4.5 therefore treats the LAN-fixed kernel as a required runtime file and adds selectable IPX services.

## Internal tests

- Go unit/integration suite with race detector: PASS.
- `go vet`: PASS.
- New network-service tests: Direct Host/Join, Dopefish Create/Join, hostname validation, and disabled future SGC service: PASS.
- Real owned-file portable integration reached and passed the 1.50.26 community build with exact `ORION150.EXE`, exact LAN-fixed `RKERNEL.COM`, and generated `IPXNET CONNECT moo2.thedopefish.com 213` configuration. The longer historical-chain test exceeded the execution-window after these checks; a shorter regression repeated the same critical checks.
- Actual DOSBox/MOO2 networking was not executed in this Linux environment. The user's Windows test remains the acceptance boundary for the external Dopefish service and real multiplayer.

## Safety / migration

Existing pre-0.4.5 1.40/1.50 workspaces missing `RKERNEL.COM` now fail verification before game launch instead of reaching MOO2's fatal Network Game error. Portable preparation can refresh an older source receipt from the verified canonical 1.40b23 ZIP without modifying the owned source. Repair rebuilds the selected managed environment.

## Scope

PRSL and the new Chat extension remain disabled. MOO2-SGC Online is a reserved service selector only; there is no first-party lobby or relay backend in 0.4.5.
