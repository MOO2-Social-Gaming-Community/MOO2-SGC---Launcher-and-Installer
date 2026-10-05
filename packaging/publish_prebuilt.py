#!/usr/bin/env python3
"""Verify prepared public artifacts, then publish to the configured repository.
No signing key is needed to publish already signed bytes. No release is ever
silently overwritten. Repeat runs verify the existing release and are no-ops.
"""
from __future__ import annotations
import argparse, hashlib, json, os, pathlib, re, subprocess, tempfile, zipfile
from source_fingerprint import source_index
ROOT=pathlib.Path(__file__).resolve().parents[1]
TARGETS=('windows-amd64','linux-amd64','darwin-amd64','darwin-arm64')

def sha(p:pathlib.Path)->str:
    with p.open('rb') as f:return hashlib.file_digest(f,'sha256').hexdigest()

def setup_name(target:str)->str:
    return 'MOO2-SGC-Setup.exe' if target=='windows-amd64' else 'MOO2-SGC-Setup-'+target

def verify_local(root: pathlib.Path=ROOT) -> tuple[pathlib.Path,dict,list[pathlib.Path]]:
    version=(root/'VERSION').read_text().strip()
    if not re.fullmatch(r'(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)',version):
        raise ValueError('VERSION must have exactly three numeric components')
    project=json.loads((root/'project.json').read_text())
    release=root/'release'/version
    trust=json.loads((release/'trust-public.json').read_text())
    if trust!=json.loads((root/'manifests/trust.json').read_text()):raise ValueError('release trust differs from source trust')
    if trust.get('development') or not trust.get('endpoints'):raise ValueError('online trust is not configured')
    subprocess.run(['go','run','./cmd/releasectl','--command','verify','--trust',str(release/'trust-public.json'),'--dir',str(release)],cwd=root/'manager',check=True)
    manifest=json.loads((release/'manifest.json').read_text())
    if manifest['release']!=version:raise ValueError('source and signed manifest versions disagree')
    expected_source=source_index(root)
    if {p['platform'] for p in manifest['packages']}!=set(TARGETS) or len(manifest['packages'])!=4:raise ValueError('expected four platform packages')
    files=[]
    for p in manifest['packages']:
        if p['kind']!='launcher' or p['id']!='launcher' or p['version']!=version:raise ValueError('unexpected package')
        expected_url=f'https://github.com/{project["repository"]}/releases/download/v{version}/{p["filename"]}'
        if not any(m['provider']=='github' and m['url']==expected_url for m in p['mirrors']):raise ValueError('wrong repository asset URL')
        archive=release/p['filename'];files.append(archive)
        with zipfile.ZipFile(archive) as z:
            signed_source=json.loads(z.read('SOURCE-FINGERPRINT.json'))
        if signed_source!=expected_source:raise ValueError('prepared binary build inputs differ from current source; rebuild and sign a new version')
        helper='MOO2-SGC-Setup.exe' if p['platform']=='windows-amd64' else 'MOO2-SGC-Setup'
        record=p['files'].get(helper)
        file=release/setup_name(p['platform'])
        if not record or sha(file)!=record['sha256'] or file.stat().st_size!=record['size']:raise ValueError('standalone setup differs from signed package')
        files.append(file)
    for n in ('manifest.json','manifest.sig','trust-public.json','SHA256SUMS.txt','RELEASE-NOTES.md','BUILD-RESULTS.json'):
        files.append(release/n)
    # The checksum list is a convenience; package/manifest signatures carry trust.
    for line in (release/'SHA256SUMS.txt').read_text().splitlines():
        h,n=line.split('  ',1)
        if '/' in n or '\\' in n or not re.fullmatch('[a-f0-9]{64}',h):raise ValueError('unsafe checksum entry')
        if sha(release/n)!=h:raise ValueError('checksum differs: '+n)
    allowed={p.name for p in files}|{'BUILD-RESULTS.json'}
    for p in release.iterdir():
        if not p.is_file() or p.name not in allowed:raise ValueError('unexpected release input: '+p.name)
    return release,project,files

def execute_publish(repo:str,version:str,files:list[pathlib.Path],commit:str,run=subprocess.run)->str:
    tag='v'+version
    def call(*args,check=True):return run(['gh',*args],capture_output=True,text=True,check=check)
    view=call('api',f'repos/{repo}/releases/tags/{tag}',check=False)
    if view.returncode:
        if '404' not in view.stderr and '404' not in view.stdout:raise RuntimeError('Could not check existing release: '+view.stderr)
        args=['release','create',tag,'--repo',repo,'--draft','--title',version,'--notes-file',str(next(p for p in files if p.name=='RELEASE-NOTES.md'))]
        if commit:args+=['--target',commit]
        call(*args)
        info={'draft':True,'assets':[]}
    else:info=json.loads(view.stdout)
    existing={x['name'] for x in info.get('assets',[])}
    expected={p.name for p in files}
    if existing-expected:raise RuntimeError('Existing release has unexpected assets; inspect before changing it')
    if not info['draft'] and existing!=expected:raise RuntimeError('Published release asset set differs; create a new version instead')
    with tempfile.TemporaryDirectory(prefix='sgc-release-readback-') as tmp:
        target=pathlib.Path(tmp)
        # Before resuming a draft, verify every previously uploaded asset.
        if existing:
            call('release','download',tag,'--repo',repo,'--dir',str(target))
            for p in files:
                if p.name in existing and sha(target/p.name)!=sha(p):raise RuntimeError('Existing remote asset differs; refusing overwrite: '+p.name)
        missing=[p for p in files if p.name not in existing]
        if missing:
            call('release','upload',tag,'--repo',repo,*map(str,missing))
            for p in missing:call('release','download',tag,'--repo',repo,'--pattern',p.name,'--dir',str(target))
        for p in files:
            if sha(target/p.name)!=sha(p):raise RuntimeError('Readback differs; leaving release draft: '+p.name)
    if info['draft']:
        call('release','edit',tag,'--repo',repo,'--draft=false','--prerelease=false','--latest')
        return 'Published verified release '+tag
    return 'Existing published release is byte-identical; no changes made'

def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('--execute',action='store_true');p.add_argument('--commit',default='');a=p.parse_args()
    release,project,files=verify_local()
    version=release.name
    plan={'version':version,'repository':project['repository'],'tag':'v'+version,'assets':[x.name for x in files],'remote_writes':a.execute}
    print(json.dumps(plan,indent=2))
    if a.execute:
        if os.environ.get('GITHUB_REPOSITORY',project['repository'])!=project['repository']:raise SystemExit('Refusing publication from a different repository/fork')
        print(execute_publish(project['repository'],version,files,a.commit))
if __name__=='__main__':main()
