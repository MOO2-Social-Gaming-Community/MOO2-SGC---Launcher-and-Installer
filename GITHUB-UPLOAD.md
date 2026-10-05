# GitHub Desktop upload — 0.4.1

Extract the complete repository ZIP directly into the cloned repo folder. Keep `.git`. Commit and push the supplied root files, `.github`, `manager`, `packaging`, `tests`, `release`, `manifests`, documentation and public evidence. The ignore rules explicitly permit the reviewed release EXEs/ZIPs, while attributes keep signed metadata byte-identical on Windows.

Push to the default branch of `MOO2-Social-Gaming-Community/MOO2-SGC---Launcher-and-Installer`.

**This push automatically publishes the prepared v0.4.1 software Release after the publishing workflow's checks succeed.** It does not wait for your later manual game test. The release notes disclose the test limitations. Ordinary source commits without a new prepared release do not sign or manufacture a new version.

Then open GitHub → Actions → Publish prepared release. Once successful, open Releases → 0.4.1, download `MOO2-SGC-Setup.exe`, and run it from Downloads without adjacent manifests or launcher ZIPs. The executable uses the configured online endpoints by default.

A repository push is not itself a Release. If there is no Release, inspect the workflow status. Organization policies may require enabling Actions/write permissions. A private repository or draft Release cannot be accessed by this anonymous public installer. No account token is embedded or requested.

Do not upload commercial MOO2 archives, prior PRIVATE_TEST packages, or the separately supplied PRIVATE maintainer signing backup. The prepared release contains no commercial game payloads.

The exact source inputs are fingerprinted inside signed launcher packages. Do not alter application/build/workflow source before initially pushing this prepared release; its verifier will correctly require a rebuilt and newly signed package if code changes. Editing documentation alone does not invalidate binary correspondence.
