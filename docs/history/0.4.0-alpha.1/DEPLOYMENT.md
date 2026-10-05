# Production deployment handoff — not deployed in this cycle

## What remains with the project owner

Create/administer the planned `MOO2-SGC` GitHub organization and `website`, `launcher`, `prsl` repositories. Use the dedicated MOO2-SGC Cloudflare account for R2, never the unrelated Data Pioneer/Pioneering DataWorks account. This run created no accounts, repositories, buckets, domains, secrets or public releases.

The source ZIP represents the launcher repository. Bootstrapper/updater/patcher/configurator remain modules of that repository. The prior PRSL research handoff stays with `prsl`; website source stays with `website`. Neocities remains the human-facing portable site, not a binary or metadata backend. SourceForge remains deferred.

## Keys and metadata endpoints

Start with `manifests/trust.production.EXAMPLE.json`. It intentionally contains no keys/endpoints and cannot authorize downloads until configured.

Generate a private key OUTSIDE the repository and release output:

```sh
cd manager
go run ./cmd/releasectl --command keygen --key /secure/location/release.key --key-id maintainer-1
```

Store the printed public key in the trust file. Never upload the private key or bake it into an executable. Add metadata endpoint pairs (`url`, `signature_url`) with provider `github`, priority 10; optionally add `cloudflare-r2`, priority 20. Every destination and redirect hostname must be in the exact allowed-host list. URLs must use HTTPS.

Stable releases may use GitHub's latest-release asset URLs. A prerelease channel must use an explicitly maintained discoverable channel endpoint; **do not assume GitHub's latest-release shortcut tracks alpha releases**. The included publisher marks hyphenated tags as prereleases. A production alpha-channel head on GitHub has not been implemented/deployed here. Final endpoint topology is an operator provisioning decision, not an invented existing service.

R2 package objects are mirrored under `<release_tag>/<filename>`. Mirrored channel metadata is published last under `channels/<channel>/manifest.json` and `manifest.sig`. Configure a real public-download endpoint mapping those keys; the S3 administrative endpoint is not a credential-free client download URL. Verify actual public reachability, ranges, content length, cache behavior and complete outage fallback before embedding production URLs.

## Build and sign

```sh
python3 packaging/build_release.py \
  --out /release/staging \
  --trust /secure/public-trust.json \
  --key /secure/release.key \
  --key-id maintainer-1 \
  --revision 100 \
  --github-base https://github.com/MOO2-SGC/launcher/releases/download/v0.4.0/ \
  --r2-base https://YOUR-PUBLIC-MIRROR/v0.4.0/
```

Those are configuration examples, not live project URLs. Use the real version in source and the matching tag. Omit R2 entirely when not provisioned. Revision must increase monotonically per feed/channel; it is not the game patch version or semantic release version.

The build emits four native setup executables, four launcher ZIPs, signed metadata, checksums and build records. It never reads base-game archives. Windows executable signing/macOS signing should be incorporated before packaging when configured; update signatures must be regenerated if a packaged executable changes afterward.

## Publish once

Dry run is the default:

```sh
python3 packaging/publish_release.py --release /release/staging --tag v0.4.0 --repository MOO2-SGC/launcher
```

After operator review, `--execute` performs GitHub writes. It verifies local signed packages, creates a draft, uploads artifacts, downloads them again and checks their hashes, then publishes GitHub. Optional R2 copying happens afterward and cannot withdraw a successful GitHub release. Metadata is mirrored after payload verification; failures are reported for retry. Development-trust kits are explicitly refused by the publication path before any network write.

The GitHub CLI and AWS CLI must be available for remote publication. R2 credentials belong only in the release job/maintainer environment. The clients need no GitHub token, R2 token, cloud secret, or account login. Remote commands and credentials were not exercised here.

## Workflow scaffolds

`.github/workflows/test.yml` and `release.yml` are provided. The release job is manual and names a protected `release-signing` environment. Configure required reviewers and branch restrictions before adding secrets. Pin action references to reviewed full commit SHAs before production, as recommended by GitHub; the supplied `@v7` references are reviewable scaffolds, not immutable supply-chain pins. The action documentation was checked during this cycle. The workflow has not run on GitHub.

Set `SGC_GO_VERSION` to an approved exact toolchain for releases (the template defaults to `stable` for discovery/CI). The local binaries in this handoff were built with the available Go 1.23.2 toolchain; this is a reproducibility record, not a recommendation to keep that toolchain indefinitely. `SGC_TRUST_JSON`, `SGC_KEY_ID`, optional R2 endpoint/bucket/public-base variables, and protected signing/R2 secrets are described directly in `release.yml`. Prefer offline release signing for the long-lived trust authority; the single-key CI scaffold is not a completed root/online-key separation design.

## Acceptance gates before public release

Native Windows/macOS startup; real game session tests; reviewed license and redistribution records; provisioned ownership/recovery; official signing identity; live GitHub-only operation; live R2-only recovery; no dependence on Neocities; repeated interrupted downloads; safe rollback; key rotation plan; and public-metadata discovery for each advertised release channel.

## Public documentation verified

- GitHub releases: https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases
- R2 S3 compatibility and Range support: https://developers.cloudflare.com/r2/api/s3/api/
- GitHub Actions security guidance: https://docs.github.com/en/actions/reference/security/secure-use
- Current action documentation: https://github.com/actions/checkout and https://github.com/actions/setup-go
