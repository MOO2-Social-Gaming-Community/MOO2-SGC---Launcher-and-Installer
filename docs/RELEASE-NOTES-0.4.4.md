# MOO2-SGC 0.4.4

This maintenance release addresses the first real Windows gameplay acceptance findings.

## Fixed

- **Owned-game ZIP preparation:** selecting a recognized `Master of Orion 2 - v1_40b23.zip` (or another supported fingerprinted source ZIP) in **Prepare your owned game** now imports it through the same content-verified source path used by folder imports and portable auto-discovery.
- **Windowed scaling:** the bundled DOSBox Staging configuration now explicitly uses `viewport = fit` and `integer_scaling = off`, so MOO2 scales to the available 4:3 window area instead of leaving avoidable internal padding. Fullscreen continues to preserve the game's aspect ratio.

## Unchanged

- The owned Steam/GOG/CD source remains untouched.
- Portable baseline remains **1.40b23**.
- Community **1.50.26** remains an optional separate managed environment.
- PRSL and the new Chat extension remain unavailable pending live integration testing.
- No commercial MOO2 data, DOSBox runtime, or private signing key is included in the public release.
