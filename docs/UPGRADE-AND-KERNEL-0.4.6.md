# Investigation: stale launcher and missing network driver

## Established from code and controlled execution

0.4.5 FetchManifest returned the first authenticated, unexpired manifest. Its newest-trusted revision prevented rollback relative to previously accepted metadata, but there was no minimum relative to the setup executable. Thus an old yet valid release was acceptable in a fresh root or a root still trusting revision 42. The pinned fallback was not examined once latest succeeded.

A 0.4.2 imported source snapshot intentionally excluded RKERNEL.COM. The old environment verifier did not require that file; launching the title screen could succeed without proving the later network path. The new game-only fix could not help a user who was actually still executing the old launcher.

The current regression suite recreates both conditions with prior signed binaries and real owned archive data. This does not establish which event left the live repository latest pointer old: Actions failure, incomplete publication, cached responses, or reuse of an older entrypoint must be distinguished from live logs. No remote publication state is presumed.

## Changed contracts

- Trust signature and metadata freshness are necessary but not sufficient. Setup's local semantic-version floor applies before acceptance; stale latest falls through to the pinned endpoint. Existing verified versions newer than setup are not silently downgraded.
- Offline new installation and direct launch-installed also reject older launchers. An explicit rollback is intentionally distinct.
- Every launch records and checks the actual executable directory. 1.40b23/1.50.26 require the pinned RKERNEL.COM there even for a standalone role.
- Kernel import only populates a private verified cache. Application to an environment occurs via ordinary staged reconstruction, preserving saves and prior generations. Unknown bytes are never patched blindly.
- Older source receipt schemas stay readable; their source bytes are not mutated to claim completeness. New built game manifests include the required driver as an immutable pinned component.

## Test boundaries

Linux managers and signed installers actually execute in the tests. The HTTP fixture serves real signed releases over a local TLS proxy; it is not a live GitHub test. Game process tests deliberately substitute a small labelled shell process for DOSBox. File hashes, cwd and configuration coverage are verified, but DOS guest semantics, Windows loading, live networking, audio/video and semantic game saves remain external acceptance tasks.
