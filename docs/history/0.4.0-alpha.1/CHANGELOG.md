# Change log

## 0.4.0-alpha.1 — distribution integration

Introduces a separate self-contained setup executable and signed modular launcher ZIPs; schema-1 manifest with feed/channel/revision/expiry and file indexes; Ed25519 manifest/package verification; GitHub-primary/R2-secondary HTTPS downloads with validated resume; staged immutable launcher versions, rollback and installed-file verification; shared application locks; exact-archive local imports; explicit manager update check/stage and bootstrap apply-staged; release build/sign/mirror scaffolding; and documentation/tests for the supplied infrastructure handoff.

Retains schema-1 game profiles and prior game-workspace semantics. Does not enable PRSL or Chat. Does not include commercial game payloads or DOSBox. Does not deploy production services, create accounts, introduce SourceForge, or use Neocities as a software backend. The prior all-in-one private test archive is superseded as the distribution model, not deleted as a useful local source fixture.
