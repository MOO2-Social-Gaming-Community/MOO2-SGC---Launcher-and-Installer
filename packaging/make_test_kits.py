#!/usr/bin/env python3
"""Package explicitly development-only, per-platform offline acceptance kits.
No game data/runtime/private key is read. This is not the production release path.
"""
from __future__ import annotations
import argparse, hashlib, json, pathlib, shutil, stat, zipfile
ROOT=pathlib.Path(__file__).resolve().parents[1]
LABELS={'windows-amd64':'WINDOWS','linux-amd64':'LINUX','darwin-amd64':'MAC_INTEL','darwin-arm64':'MAC_APPLE_SILICON'}

def archive(folder:pathlib.Path,destination:pathlib.Path)->None:
    with zipfile.ZipFile(destination,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
        for p in sorted(folder.rglob('*')):
            if not p.is_file():continue
            info=zipfile.ZipInfo(str(pathlib.PurePosixPath(folder.name)/p.relative_to(folder).as_posix()),(2026,10,4,0,0,0))
            info.create_system=3;info.external_attr=(stat.S_IFREG | (0o755 if p.stat().st_mode & 0o111 else 0o644))<<16
            z.writestr(info,p.read_bytes(),compress_type=zipfile.ZIP_DEFLATED,compresslevel=9)
    with zipfile.ZipFile(destination) as z:
        if z.testzip():raise RuntimeError('ZIP integrity failed')

def main()->None:
    ap=argparse.ArgumentParser();ap.add_argument('--release',type=pathlib.Path,required=True);ap.add_argument('--out',type=pathlib.Path,required=True);a=ap.parse_args()
    r=a.release.resolve();out=a.out.resolve();out.mkdir(parents=True,exist_ok=True)
    m=json.loads((r/'manifest.json').read_text());trust=json.loads((r/'trust-public.json').read_text())
    if not trust['development'] or trust.get('endpoints'):raise SystemExit('Test kit packaging expects offline DEVELOPMENT trust only')
    version=m['release']; records=[]
    for target,label in LABELS.items():
        kit=out/f'MOO2_SGC_{version}_{label}_TEST_KIT'
        if kit.exists():raise SystemExit('Refusing to overwrite '+str(kit))
        rel=kit/'release';rel.mkdir(parents=True)
        for n in ('manifest.json','manifest.sig','trust-public.json'):shutil.copy2(r/n,rel/n)
        ps=[p for p in m['packages'] if p['kind']=='launcher' and p['platform']==target]
        if len(ps)!=1:raise SystemExit('ambiguous launcher')
        p=ps[0];source=r/p['filename'];b=source.read_bytes()
        if len(b)!=p['size'] or hashlib.sha256(b).hexdigest()!=p['sha256']:raise SystemExit('package mismatch')
        shutil.copy2(source,rel/p['filename'])
        win=target.startswith('windows');exe='MOO2-SGC-Setup.exe' if win else 'MOO2-SGC-Setup'
        shutil.copy2(r/'binaries'/target/exe,kit/exe);(kit/exe).chmod(0o755)
        cmds={'START-SETUP':'--offline RELEASE','OPEN-LAUNCHER':'--command launch-installed','REPAIR-LAUNCHER':'--command repair --offline RELEASE --no-launch','VERIFY-LAUNCHER':'--command verify','APPLY-STAGED':'--command apply-staged','ROLLBACK-LAUNCHER':'--command rollback'}
        for name,args in cmds.items():
            if win:
                body='@echo off\r\nsetlocal\r\ncd /d "%~dp0"\r\n"%~dp0'+exe+'" '+args.replace('RELEASE','"%~dp0release"')+'\r\nset "RESULT=%ERRORLEVEL%"\r\nif not "%RESULT%"=="0" (\r\n  echo.\r\n  echo Setup stopped. Preserve the message above; do not disable security software.\r\n  pause\r\n)\r\nexit /b %RESULT%\r\n'
                (kit/(name+'.cmd')).write_bytes(body.encode('utf-8'))
            else:
                ext='.command' if target.startswith('darwin') else '.sh'
                body='#!/bin/sh\nset -eu\nHERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)\nexec "$HERE/'+exe+'" '+args.replace('RELEASE','"$HERE/release"')+'\n'
                f=kit/(name+ext);f.write_text(body);f.chmod(0o755)
        shutil.copy2(ROOT/'START-HERE.md',kit/'START-HERE.md');shutil.copy2(ROOT/'THIRD-PARTY-NOTICES.md',kit/'THIRD-PARTY-NOTICES.md')
        docs=kit/'docs';docs.mkdir()
        for n in ['TEST-REPORT.md','TEST-CHECKLIST.md','SECURITY-AND-TRUST.md','DISTRIBUTION-ARCHITECTURE.md','GO-LICENSE.txt']:
            shutil.copy2(ROOT/'docs'/n,docs/n)
        shutil.copy2(ROOT/'manager/evidence/launcher-desktop.png',docs/'LAUNCHER-PREVIEW.png')
        (kit/'PLATFORM.txt').write_text(f'{target}\nVersion {version}\nDevelopment signing trust; unsigned/notarization unavailable. Only Linux executed in the development environment.\nManifest expires {m["expires"]} for NEW installation/update; verified installed launcher remains available offline.\n')
        hashes=[hashlib.sha256(f.read_bytes()).hexdigest()+'  '+f.relative_to(kit).as_posix() for f in sorted(kit.rglob('*')) if f.is_file()]
        (kit/'SHA256SUMS.txt').write_text('\n'.join(hashes)+'\n')
        dest=out/(kit.name+'.zip');archive(kit,dest)
        records.append({'platform':target,'file':dest.name,'bytes':dest.stat().st_size,'sha256':hashlib.sha256(dest.read_bytes()).hexdigest(),'setup_bytes':(kit/exe).stat().st_size,'launcher_zip_bytes':p['size'],'game_data':False,'development_trust':True})
    (out/f'MOO2_SGC_{version}_KIT-RESULTS.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(records,indent=2))
if __name__=='__main__':main()
