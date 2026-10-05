#!/usr/bin/env python3
"""Verify Git Desktop-style text handling and public/private upload boundaries.
Works only on a disposable local Git copy; never connects to GitHub.
"""
from pathlib import Path
import argparse,hashlib,json,os,shutil,subprocess,sys,tempfile
ROOT=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(ROOT/'packaging'));from source_fingerprint import source_index
p=argparse.ArgumentParser();p.add_argument('--output',type=Path,required=True);a=p.parse_args();checks=[]
def check(name,ok):
 if not ok:raise AssertionError(name)
 checks.append({'test':name,'result':'PASS'});print('PASS',name)
with tempfile.TemporaryDirectory(prefix='sgc-git-roundtrip-') as td:
 root=Path(td)/'repository'
 shutil.copytree(ROOT,root,ignore=shutil.ignore_patterns('.git','__pycache__','*.pyc'))
 def git(*args):return subprocess.check_output(['git','-C',str(root),*args],stderr=subprocess.STDOUT,text=True)
 git('init','-q');git('config','core.autocrlf','true');git('config','user.email','test@example.invalid');git('config','user.name','Local packaging test')
 dummies={'never-upload.key':b'not a key','original.LBX':b'not a game','release/0.4.1/PRIVATE_BACKUP.zip':b'not a secret archive','release/0.4.1/unexpected.key':b'not a key'}
 for n,b in dummies.items():(root/n).write_bytes(b)
 git('add','.');tracked=set(git('ls-files').splitlines())
 check('dummy private/game inputs excluded from Git',not set(dummies)&tracked)
 release=ROOT/'release'/(ROOT/'VERSION').read_text().strip()
 release_paths=[p.relative_to(ROOT).as_posix() for p in release.iterdir() if p.is_file()]
 check('every prepared release artifact is tracked',all(n in tracked for n in release_paths))
 check('workflow and root metadata are tracked',all(n in tracked for n in ['.github/workflows/release.yml','.github/workflows/test.yml','VERSION','manifests/trust.json','.gitattributes','.gitignore']))
 expected=source_index(ROOT)
 check('every signed source input is tracked',set(expected)<=tracked)
 git('commit','-qm','local roundtrip, no remote')
 for n in dummies:(root/n).unlink()
 for n in tracked:(root/n).unlink()
 git('checkout-index','-a','-f')
 check('source fingerprints survive autocrlf checkout',source_index(root)==expected)
 check('signed release bytes survive autocrlf checkout',all((ROOT/n).read_bytes()==(root/n).read_bytes() for n in release_paths))
 check('no commercial game or private key tracked',not any(n.lower().endswith(('.lbx','.gam','.rac','.key','.pem','.p12','.pfx')) or 'PRIVATE' in Path(n).name for n in tracked))
 check('all public files fit ordinary Git size limit',all((root/n).stat().st_size<100*1024*1024 for n in tracked))
 check('no Git remote configured',git('remote').strip()=='')
 check('release signatures still verify after Git checkout',subprocess.run([sys.executable,str(root/'packaging/publish_prebuilt.py')],capture_output=True,text=True).returncode==0)
a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':'0.4.1','count':len(checks),'checks':checks,'remote_access':False,'core_autocrlf':True},indent=2)+'\n')
