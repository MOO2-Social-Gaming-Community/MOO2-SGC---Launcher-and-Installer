# Signing and future versions

The 0.4.1 artifacts are already signed. The push-to-release workflow verifies and publishes these exact bytes using GitHub's built-in workflow token. It requires **no private signing key** for this first upload.

A new project release-signing identity was generated for this package. Its public key is in `manifests/trust.json` and is embedded in both binaries. A separate owner-only backup contains `moo2-sgc-release-1.key`. **Never put that backup or key in this repository, release assets, installation kits, issue attachments or chat logs.** Store it outside the repository in a secure private location. It is not Authenticode, a GitHub token, or a Cloudflare key.

For subsequent builds, keep the same public trust identity and supply the private key from outside the output/repository directory:

```
python packaging/build_release.py --out <EMPTY-OUTPUT> --trust manifests/trust.json --key <PRIVATE-OUTSIDE-REPO>/moo2-sgc-release-1.key --key-id moo2-sgc-release-1 --revision <INCREASED-INTEGER> --github-base https://github.com/MOO2-Social-Gaming-Community/MOO2-SGC---Launcher-and-Installer/releases/download/v<NEXT-VERSION>/ --days 90
```

First update `VERSION` and the compile-time version in `manager/shared/buildconfig/config.go` to matching three-part values. Refresh relevant source, tests and release notes; rebuild all artifacts and publish a new version rather than overwriting an old one. Prefer an up-to-date supported Go toolchain; the included binaries record the compiler available in the development environment.

Signed metadata expires after 90 days. A valid signed renewal/new release with a larger revision is needed for new installs/updates after expiry. Previously installed, verified software can still launch offline. Do not disable expiry/signature checks to work around this. Changing metadata, URLs or ZIP contents requires re-signing.

The current key was generated in this development session, not on the owner's physical machine. Before broader distribution, the owner may choose to generate a replacement locally and rebuild. Existing installers trust only embedded public keys; key rotation is not an automatic trust-on-first-use operation. Do not silently trust a downloaded replacement key.

If the key is lost, old clients cannot authenticate newly signed packages under an unrelated key. Recovery requires a deliberately distributed new bootstrap with the new trust root. If the key is exposed, treat it as compromised and replace/rebuild through a trusted channel.
