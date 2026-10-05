# GitHub Desktop upload — 0.4.2

Extract the ZIP contents directly into the clone, preserving `.git`. Commit and push to the default branch. The existing Publish prepared release workflow verifies signed artifacts and publishes v0.4.2. No signing key or additional account configuration is required for these prepared files. Keep previous published releases unchanged.

The files under `release/0.4.2/` are already signed. Do not rename/repackage launcher ZIPs or edit the manifest. The root source input fingerprints must remain consistent with the binaries. Existing v0.4.1 source inputs are historical; current publication uses VERSION.

After the v0.4.2 release exists, update through the existing launcher or run the new standalone setup. Follow START-HERE.md to test Steam recognition and game preparation. No commercial archives, private keys or local user data belong in the repository.
