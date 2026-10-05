# Development and test report — 0.3.0-alpha.1

## Executed in this cycle

| Suite | Result | Scope |
|---|---|---|
| New Go manager tests | 49 top-level tests passed; 78 test/subtest results including table cases | Resolver, configuration, archive/path safety, local API authorization, updater/cache behavior, profile persistence |
| Go race detector | Passed for those tests | Not a proof of absence of all races |
| Go unit statement coverage | 37.7% | Integration tests are separate processes and not included in this coverage percentage |
| Go vet | Passed | Static analysis, not a security audit |
| JavaScript syntax | Passed | Node syntax check |
| Real fixture/API plus offline browser integration | 37 checks passed | Details in integration-tests.json; not 37 game tests |
| Prior PRSL Python regression suite | 52 tests passed again | Includes exact engine structure, coordinator, older lab package/update tests |
| Prior isolated original-x86 probe | 30 checks passed again | Same lab scope: original instructions, stubs and synthetic state; NOT DOSBox or full-game execution |
| Binary build targets | Windows x64, Linux x64, macOS Intel, macOS ARM64 compiled | Only the Linux manager was executed |

No test suite is being presented as full coverage. The manager's branches for official runtime downloads and platform-specific extraction remain unexecuted in this environment.

## Real installation result

A fresh 1.50.26 workspace verified **931 files**, with no integrity failures. The earlier lab's 944-file count included 13 historical personal-state files. This new manager deliberately does not seed fresh installs with those old saves/race/score/settings files. The source ZIP is unchanged and still contains them.

Real integration checks covered mod-switch reconstruction, save and USER.CFG preservation, native engine corruption detection, refusal to launch a corrupt engine, reconstruction repair, rejection of a corrupt cached dependency without changing the active generation, diagnostic export, deactivation without save deletion, retained-generation activation, and an original 1.31 environment without the 1.50 executable or tree.

Both original uploaded archives remained byte-identical.

## Browser testing boundary

System Chromium was available, but its environment policy blocked navigation to the loopback launcher URL with ERR_BLOCKED_BY_ADMINISTRATOR. That policy was not disabled.

The UI was therefore rendered as an **offline DOM fixture** in Chromium with actual application HTML/CSS/JS and precomputed API response fixtures. Startup controls, dependencies, conflict display, original-engine control disabling, narrow-layout overflow and JavaScript errors were checked. The screenshot shows this offline fixture. Separately, Python exercised the actual live loopback HTTP API with normal requests.

This is NOT a browser-to-live-backend end-to-end acceptance test. That test belongs in the user's local test cycle.

## Subprocess testing boundary

A deliberately labeled `dosbox` shell TEST DOUBLE was used to test argument generation, process startup, log capture, wait/reaping and status changes. It was removed afterward and is not part of the private distribution. It did not emulate DOS, execute MOO2 or establish IPX.

## Not executed

- Actual DOSBox startup or game title screen.
- Single-player turns, save/load or combat.
- Two-client IPX game or synchronization.
- Live PRSL interception, release, timer, player roster or Chat.
- Windows or macOS manager execution.
- Official runtime download and platform-specific extraction.
- Hosted launcher update-feed or automatic binary replacement.
- Gameplay testing of every resolved community combination.

## Development issues found and handled

- Exact community metadata includes internal USER and `mod_class = 0`; USER is not a selectable gameplay package, and class 0 is normalized to independent add-ons.
- Core rulesets auto-enable content; those dependencies are exposed and exclusive-group conflicts are rejected.
- Historical personal saves were previously entering fresh workspaces; fresh builds now omit them while retaining source bytes.
- Default interruption would leave a stale manager lock; explicit shutdown and ordinary idle SIGINT/SIGTERM now close the server and remove the lock. Interrupting during a running game/operation is refused to avoid dropping the owner process.
- JSON reads now enforce file-size limits before decoding as well as rejecting unknown fields/trailing objects.
- The first browser check was blocked by environment policy; coverage was split into offline UI and actual HTTP integration rather than bypassing that policy.
- A first offline UI test tried to select a checkbox in a collapsed details section; the test was corrected to open the section, as a user would.

## Evidence

The source handoff's `manager/evidence` directory contains the test logs, JSON results, source/build records and screenshots. Token-bearing server startup logs are excluded. Integration fixtures and deliberately corrupted generations are not included in the user's runnable bundle.
