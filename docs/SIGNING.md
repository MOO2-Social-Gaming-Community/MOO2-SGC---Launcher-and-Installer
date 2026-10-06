# Release signing — 0.4.6

Uses the existing `moo2-sgc-release-1` public identity, feed `moo2-sgc`, channel `stable`, revision **46**. No replacement private key is distributed. The previously stored private backup remains the maintainer's signing backup; never commit it, even temporarily.

Prepared metadata expires **2027-01-04T06:34:38Z**. Existing verified installations can run offline; new installation/update authorization needs fresh signed metadata after expiration. Never edit signed JSON or suppress signature/freshness failures.

The private signing key was kept outside the repository and public output while building these supplied packages. GitHub publication verifies and uploads already-signed artifacts; it does not need that key. Future builds must be signed by the owner-controlled identity, unless an explicit trust migration is performed.

Ed25519 release signatures are not Windows Authenticode or macOS notarization. These Windows/Mac executables remain unsigned by those operating-system schemes. Do not disable system protection to force acceptance.

See `DEPLOYMENT.md` and `UPGRADE-AND-KERNEL-0.4.6.md` for the separate local minimum-version policy. Explicit rollback is distinct from silent fallback to an older launcher.
