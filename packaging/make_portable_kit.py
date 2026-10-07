#!/usr/bin/env python3
"""Create project-only portable overlay. Does not read/copy any owned game,
DOSBox runtime, fonts, source ZIP, existing private working root or signing key.
"""
from pathlib import Path
import argparse,hashlib,json,subprocess,sys,zipfile
ROOT=Path(__file__).resolve().parents[1]
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--out',required=True,type=Path);a=p.parse_args();out=a.out.resolve()
 if out.exists():raise SystemExit('Refusing to overwrite existing kit')
 subprocess.run([sys.executable,str(ROOT/'packaging/publish_prebuilt.py')],check=True,cwd=ROOT)
 version=(ROOT/'VERSION').read_text().strip();rel=ROOT/'release'/version
 package='launcher-'+version+'-windows-amd64.zip'
 entries={'MOO2-SGC-Setup.exe':rel/'MOO2-SGC-Setup.exe','START-HERE-PORTABLE.md':ROOT/'START-HERE.md'}
 for n in ['manifest.json','manifest.sig',package]:entries['distribution/offline-'+version+'/'+n]=rel/n
 for f in (ROOT/'packaging/portable').iterdir():
  if f.is_file():entries[f.name]=f
 for f in (ROOT/'support').rglob('*'):
  if f.is_file():entries[f.relative_to(ROOT).as_posix()]=f
 for n in ['PORTABLE-HARNESS.md','TEST-CHECKLIST.md','TEST-REPORT.md','NETWORK-AND-PROFILES-0.4.7.md']:
  entries['sgc-docs/'+n]=ROOT/'docs'/n
 entries['sgc-docs/THIRD-PARTY-NOTICES.md']=ROOT/'THIRD-PARTY-NOTICES.md'
 sums=[]
 out.parent.mkdir(parents=True,exist_ok=True)
 with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
  for n,f in sorted(entries.items()):
   if n.lower().endswith(('.lbx','.gam','.key','.ttf','.otf','.woff','.dll')) or n.startswith(('game/','runtime/')):raise SystemExit('Forbidden kit payload '+n)
   b=f.read_bytes();i=zipfile.ZipInfo(n,(2026,10,5,0,0,0));i.compress_type=zipfile.ZIP_DEFLATED;i.create_system=3;i.external_attr=0o100644<<16;z.writestr(i,b)
   sums.append(hashlib.sha256(b).hexdigest()+'  '+n)
  z.writestr('KIT-SHA256SUMS.txt','\n'.join(sums)+'\n')
 print(json.dumps({'version':version,'kit':str(out),'members':len(entries)+1,'game_data':False,'runtime_binaries':False,'signing_keys':False},indent=2))
if __name__=='__main__':main()
