# Start here — MOO2-SGC 0.4.2

## Publish

1. Extract the complete repository ZIP into the cloned repository root. Preserve `.git` and keep your private signing backup outside the clone.
2. Commit and push with GitHub Desktop. Wait for **Publish prepared release** to finish and confirm a published `v0.4.2` Release.
3. Update from the existing launcher, or run `MOO2-SGC-Setup.exe` from the new Release. Do not run the new setup before the Release is published and expect its download to resolve to 0.4.2.

## Steam test first

Confirm the launcher shows **0.4.2**. Choose **Community — standard**. Under **Prepare your owned game**, select or paste the folder containing Steam's `Orion2.exe` and `.LBX` files. In the reported test this was:

```
B:\SteamLibrary\steamapps\common\Master of Orion 2
```

Click **Prepare game for play**. Expected: source recognized as **Steam English DOS 1.40b23**, isolated copy prepared, fan patch 1.50.26 applied and verified, DOSBox reused or downloaded. It does not start the game automatically. It does not rerun the CD/1.31 patches on Steam.

Click **Launch MOO2**. Check actual title screen, keyboard/mouse, sound, display and version. Start a new single-player game, take several turns, save, exit and reload. A created process or successful hash check is not a substitute for this test.

## Baseline test

Select **SGC baseline — 1.40b23**, retain the same Steam source, and prepare/launch. This separate profile runs the recognized baseline without 1.50 rulesets. The current community environment remains retained.

## Fresh CD test

Select an independent profile (duplicate Community or use Original CD) and provide the original `Master of Orion 2 - v1_2(1).zip` or its extracted game folder. For target 1.50.26, the internal sequence is:

```
CD 1.2 -> official 1.31 -> baseline 1.40b23 -> fan 1.50.26
```

The official 1.31 prerequisite is downloaded when needed and each installed member is checked against compiled hashes. If that host is unavailable, choose **Official prerequisite — English 1.31 ZIP** in Packages and import your archived original patch ZIP. Steam tests do not require that download.

To play original 1.2, provide the CD source. The manager cannot reconstruct original CD bytes from the later Steam engine and will not fake a downgrade. Newly imported sources become the current preparation source; existing prepared environments remain intact.

## Failure reporting

Do not edit generated game files or disable integrity checks to bypass an error. Record the first failing action and exact error; use **Export diagnostics**, retaining the relevant game-process log. Launcher/bootstrap data remain under:

```
%APPDATA%\MOO2-SGC\stable
```

Do not post commercial game files, an active dashboard URL/token, or signing material. The unsigned Windows executables may still cause Windows reputation warnings; update signatures are not Authenticode.
