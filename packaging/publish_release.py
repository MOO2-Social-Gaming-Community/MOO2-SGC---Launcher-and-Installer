#!/usr/bin/env python3
"""Publish an already signed release once: GitHub canonical, R2 best-effort.
Default is a dry run. Requires explicit --execute for remote writes.
No signing keys or cloud tokens are read from a file in the release output.
"""
from __future__ import annotations
import argparse, hashlib, json, os, pathlib, subprocess, tempfile
ROOT=pathlib.Path(__file__).resolve().parents[1]

def cmd(args:list[str])->str:
    return subprocess.check_output(args,text=True).strip()

def main()->None:
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--release',type=pathlib.Path,required=True)
    p.add_argument('--repository',default='MOO2-Social-Gaming-Community/MOO2-SGC---Launcher-and-Installer')
    p.add_argument('--tag',required=True)
    p.add_argument('--r2-bucket')
    p.add_argument('--r2-endpoint',help='Account-specific S3 endpoint for the dedicated MOO2-SGC account')
    p.add_argument('--execute',action='store_true')
    a=p.parse_args();r=a.release.resolve()
    trust=json.loads((r/'trust-public.json').read_text())
    m=json.loads((r/'manifest.json').read_text())
    files=[r/p['filename'] for p in m['packages']]
    # Bootstrap initial downloads are separate from signed launcher ZIPs.
    upload=[]
    for pth in sorted((r/'binaries').glob('*/MOO2-SGC-Setup*')):
        target=pth.parent.name
        public_name='MOO2-SGC-Setup-'+target+('.exe' if pth.suffix=='.exe' else '')
        upload.append((pth,public_name))
    upload += [(f,f.name) for f in files]
    upload += [(r/name,name) for name in ('manifest.json','manifest.sig','SHA256SUMS.txt','trust-public.json')]
    plan={'canonical_repository':a.repository,'tag':a.tag,'development_trust':trust['development'],
          'github_assets':[n for _,n in upload], 'r2_optional':bool(a.r2_bucket),
          'steps':['verify signatures and local bytes','create GitHub draft','upload all assets','download and SHA-256 verify GitHub copies',
                   'publish GitHub release','optionally mirror artifacts to R2','verify R2 copies','publish identical signed metadata to R2 last'],
          'remote_writes':a.execute}
    print(json.dumps(plan,indent=2))
    if not a.execute:return
    if trust['development'] or not trust.get('endpoints'):
        raise SystemExit('Refusing public publication of disposable development trust / unconfigured feed')
    subprocess.check_call(['go','run','./cmd/releasectl','--command','verify','--trust',str(r/'trust-public.json'),'--dir',str(r)],cwd=ROOT/'manager')
    if bool(a.r2_bucket)!=bool(a.r2_endpoint):raise SystemExit('R2 bucket and endpoint must be supplied together')
    # Refuse to replace an existing published tag; no --clobber path.
    subprocess.check_call(['gh','release','create',a.tag,'--repo',a.repository,'--draft','--title',a.tag,
                           '--notes','MOO2-SGC modular distribution. Signed launcher packages; no commercial game data.'] + (['--prerelease'] if '-' in a.tag else []))
    with tempfile.TemporaryDirectory(prefix='sgc-publish-') as tmp:
        stage=pathlib.Path(tmp)
        import shutil
        for src,name in upload:shutil.copy2(src,stage/name)
        subprocess.check_call(['gh','release','upload',a.tag,'--repo',a.repository,*[str(stage/n) for _,n in upload]])
        verify=stage/'verify';verify.mkdir()
        subprocess.check_call(['gh','release','download',a.tag,'--repo',a.repository,'--dir',str(verify)])
        for _,name in upload:
            if hashlib.sha256((verify/name).read_bytes()).digest()!=hashlib.sha256((stage/name).read_bytes()).digest():
                raise SystemExit('GitHub copy mismatch; draft remains unpublished: '+name)
        subprocess.check_call(['gh','release','edit',a.tag,'--repo',a.repository,'--draft=false'])
        if a.r2_bucket:
            try:
                ordered=[n for _,n in upload if n not in ('manifest.json','manifest.sig')]+['manifest.json','manifest.sig']
                for name in ordered:
                    key=f's3://{a.r2_bucket}/{a.tag}/{name}'
                    subprocess.check_call(['aws','--endpoint-url',a.r2_endpoint,'s3','cp',str(stage/name),key,'--only-show-errors'])
                    copy=stage/('r2-verify-'+name)
                    subprocess.check_call(['aws','--endpoint-url',a.r2_endpoint,'s3','cp',key,str(copy),'--only-show-errors'])
                    if hashlib.sha256(copy.read_bytes()).digest()!=hashlib.sha256((stage/name).read_bytes()).digest():
                        raise RuntimeError('R2 copy differs: '+name)
                # Optional mirrored channel head; configure its public URL in the
                # embedded trust endpoints. Versioned metadata remains archival.
                for name in ('manifest.json','manifest.sig'):
                    head=f's3://{a.r2_bucket}/channels/{m["channel"]}/{name}'
                    subprocess.check_call(['aws','--endpoint-url',a.r2_endpoint,'s3','cp',str(stage/name),head,'--only-show-errors'])
                    copy=stage/('head-verify-'+name)
                    subprocess.check_call(['aws','--endpoint-url',a.r2_endpoint,'s3','cp',head,str(copy),'--only-show-errors'])
                    if hashlib.sha256(copy.read_bytes()).digest()!=hashlib.sha256((stage/name).read_bytes()).digest():
                        raise RuntimeError('R2 channel head differs: '+name)
            except (subprocess.CalledProcessError,RuntimeError) as e:
                print('R2 mirror incomplete; GitHub release remains usable. Mirror retry required:',type(e).__name__)
    print('Canonical release published. No Neocities or SourceForge dependency was used.')
if __name__=='__main__':main()
