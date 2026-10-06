#!/usr/bin/env python3
"""Private-fixture, real-file portable integration. Not a DOSBox/gameplay test.
Input archives stay local. No fixtures or generated game files enter release ZIPs.
"""
import argparse, hashlib, json, os, shutil, subprocess, sys, tempfile, time, zipfile
from pathlib import Path
from lineage_acceptance import Harness, sha
ROOT=Path(__file__).resolve().parents[1]
B23='7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5'
B131='4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f'
KFIX='18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97'
class PortableHarness(Harness):
 def game(self,id):return self.root/'game' if id=='baseline' else super().game(id)
def main():
 ap=argparse.ArgumentParser(description=__doc__)
 ap.add_argument('--launcher',required=True,type=Path);ap.add_argument('--baseline',required=True,type=Path);ap.add_argument('--patch',required=True,type=Path)
 ap.add_argument('--steam',type=Path);ap.add_argument('--cd',type=Path);ap.add_argument('--official',type=Path);ap.add_argument('--output',type=Path,required=True)
 a=ap.parse_args();checks=[];identity={};archives=[x for x in [a.baseline,a.patch,a.steam,a.cd,a.official] if x];before={x.name:sha(x) for x in archives}
 def check(n,ok,detail=None):
  if not ok:raise AssertionError(n+' '+str(detail))
  checks.append(dict(test=n,result='PASS',detail=detail));print('PASS',n,flush=True)
 def good(h,act,**kw):
  j=h.job(act,**kw)
  if j['error']:raise RuntimeError(act+': '+j['error'])
  return j['result']
 with tempfile.TemporaryDirectory(prefix='sgc-portable-') as td:
  t=Path(td);root=t/'Games'/'MOO2-SGC';root.mkdir(parents=True);data=root/'userdata'
  (root/'moo2-sgc-portable.json').write_text('{"schema":1,"layout":"portable-baseline"}\n')
  archive=root/'Master of Orion 2 - v1_40b23.zip';shutil.copy2(a.baseline,archive)
  game=root/'game';game.mkdir()
  # Simulate an existing user-owned working harness. Only a known engine and
  # synthetic save are needed to exercise adoption; staged target is fully built.
  with zipfile.ZipFile(a.baseline) as z:
   out=next(n for n in z.namelist() if Path(n).name.upper()=='ORION2V140.EXE')
   (game/'ORION2.EXE').write_bytes(z.read(out))
  (game/'SAVE1.GAM').write_bytes(b'EXISTING HARNESS SENTINEL NOT REAL GAME SAVE')
  (game/'custom-user-note.txt').write_text('Must be preserved in whole-directory backup')
  fake=t/'dosbox';fake.write_text('#!/bin/sh\necho "TEST DOUBLE ONLY: no emulator or MOO2 executed"\nprintf "working directory=%s\\n" "$PWD"\nprintf "argument=%s\\n" "$@"\nsleep 0.1\n');fake.chmod(0o700)
  h=PortableHarness(a.launcher,root,data)
  try:
   state=h.api('state');profiles={p['id']:p for p in state['profiles']}
   check('portable layout and baseline default reported',state['portable'] and state['default_profile']=='baseline' and state['version']=='0.4.5')
   good(h,'runtime-select',runtime_path=str(fake))
   r=good(h,'prepare-play',profile=profiles['baseline'],source_path='')
   mf=json.loads((root/'state/baseline.manifest.json').read_text())
   check('canonical root archive automatically recognized as manual b23',mf['source_base_kind']=='known-edition-v2:manual-en-140b23')
   check('root game is actual directory, not link',game.is_dir() and not game.is_symlink())
   check('only b23 baseline lineage is used',[x['version'] for x in mf['lineage']]==['1.40b23'])
   check('b23 is canonical executable',sha(game/'ORION2.EXE')==B23)
   check('official 1.31 retained separately',sha(game/'ORION131.EXE')==B131)
   check('manual patch output retained and matches active engine',sha(game/'ORION2V140.EXE')==B23)
   check('known kernel transformed to exact LAN-fixed hash',sha(game/'RKERNEL.COM')==KFIX)
   check('original archive byte-for-byte untouched',sha(archive)==before[a.baseline.name])
   check('existing harness save retained',(game/'SAVE1.GAM').read_bytes()==b'EXISTING HARNESS SENTINEL NOT REAL GAME SAVE')
   backups=list((root/'backups/baseline').glob('b*/game/custom-user-note.txt'))
   check('unregistered original directory fully backed up',len(backups)==1 and backups[0].read_text().startswith('Must be preserved'))
   for n in ['DIG.INI','MDI.INI','ORIONCD.INI']:
    check('exact supplied harness '+n,(game/n).read_bytes()==(ROOT/'manager/assets/harness'/n).read_bytes())
   check('baseline configuration is exact handoff config',(root/'config/moo2.conf').read_bytes()==(ROOT/'manager/assets/harness/moo2.conf').read_bytes())
   check('no 1.31 or community download required for baseline',not(data/'cache/official-1.31.zip').exists() and not(data/'cache/patch-1.50.26.zip').exists())
   check('preparation does not start process',not r['game_started'] and not h.api('state')['running'])
   v=good(h,'verify',id='baseline');check('full baseline verifies',v['ok'],{'checked':v['checked'],'fingerprint':v['fingerprint']});identity['manual_baseline']=v['fingerprint']
   receipt=json.loads((data/'cache/base-source.json').read_text())
   with zipfile.ZipFile(data/'cache'/receipt['filename']) as z:
    check('import excludes archived saves, wrappers and unselected assets',len(z.namelist())==408 and not any(n.upper().endswith('.GAM') or n.lower().endswith('.bat') for n in z.namelist()))
   # A later same-engine refresh preserves user data and retains previous game.
   saved=b'EDITED PERSONAL SAVE SENTINEL';(game/'SAVE1.GAM').write_bytes(saved);(game/'DIG.INI').write_bytes(b'; personal audio choice\n')
   prior=h.api('state')['active']['baseline'];good(h,'prepare',profile=profiles['baseline']);check('same-engine repair retains save and custom audio',(game/'SAVE1.GAM').read_bytes()==saved and (game/'DIG.INI').read_bytes()==b'; personal audio choice\n')
   hist=h.api('state')['history']['baseline'];backup=next(x['generation'] for x in hist if x['generation'].startswith('b'))
   check('prepared baseline snapshots visible for rollback',bool(backup))
   (game/'SAVE1.GAM').write_bytes(b'NEWER SAVE NOT TO MERGE ON ROLLBACK')
   good(h,'activate',id='baseline',generation=backup)
   check('rollback copies previous generation, does not merge later saves',(game/'SAVE1.GAM').read_bytes()==saved)
   check('rollback retains snapshot rather than moving it away',(root/'backups/baseline'/backup/'game').is_dir())
   # Corrupt and recover known kernel, not any retail source.
   k=game/'RKERNEL.COM';b=k.read_bytes();k.write_bytes(b[:-1]+bytes([b[-1]^1]))
   j=h.job('launch',profile=profiles['baseline']);check('kernel corruption blocks launch',bool(j['error']) and not h.api('state')['running'])
   good(h,'prepare',profile=profiles['baseline']);check('repair restores kernel and preserves save',sha(k)==KFIX and (game/'SAVE1.GAM').read_bytes()==saved)
   (game/'dosbox.conf').write_text('[autoexec]\nmalicious input\n')
   j=h.job('launch',profile=profiles['baseline']);check('unexpected game-local emulator config refused',bool(j['error']))
   (game/'dosbox.conf').unlink()
   r=good(h,'launch',profile=profiles['baseline']);expected=['--noprimaryconf','--conf',str(root/'config/moo2.conf'),str(game/'ORION2.EXE')]
   check('real launcher builds exact direct-program argv',r['arguments']==expected,r['arguments'])
   for _ in range(200):
    if not h.api('state')['running']:break
    time.sleep(.05)
   log=Path(r['log']).read_text();check('test-double scope and game working directory captured','TEST DOUBLE ONLY' in log and 'working directory='+str(game) in log)
   check('launcher process lifecycle completes without stale lock',not h.api('state')['running'])
   fp=dict(profiles['baseline'],fullscreen=True);r=good(h,'launch',profile=fp);check('fullscreen is optional launch override',r['arguments'][-2:] == ['--fullscreen',str(game/'ORION2.EXE')])
   for _ in range(200):
    if not h.api('state')['running']:break
    time.sleep(.05)
   # Community environment uses same source but never overwrites baseline game.
   base_hashes={p.name:sha(p) for p in game.iterdir() if p.is_file()}
   good(h,'payload-import',payload_kind='patch',source_path=str(a.patch));good(h,'prepare-play',profile=profiles['community'])
   cg=h.game('community');check('community game has own generation outside baseline',cg!=game and cg.is_relative_to(data/'environments/community'))
   check('community exact 1.50 executable',sha(cg/'ORION150.EXE')=='2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c')
   check('community carries exact LAN-fixed RKERNEL',sha(cg/'RKERNEL.COM')==KFIX)
   relay=dict(profiles['community'],role='host',network_service='dopefish',port=213,host='')
   r=good(h,'launch',profile=relay)
   for _ in range(200):
    if not h.api('state')['running']:break
    time.sleep(.05)
   relay_cfg=(cg.parent/'dosbox-manager.conf').read_text()
   check('dopefish relay config connects instead of starting local server','IPXNET CONNECT moo2.thedopefish.com 213' in relay_cfg and 'IPXNET STARTSERVER' not in relay_cfg)
   check('optional patch does not change baseline files',base_hashes=={p.name:sha(p) for p in game.iterdir() if p.is_file()})
   check('optional community profile does not inherit baseline save',not(cg/'SAVE1.GAM').exists())
   j=h.job('prepare',profile=dict(profiles['baseline'],engine='1.50.26',core='150'))
   check('root baseline cannot be silently switched to another engine',bool(j['error']) and sha(game/'ORION2.EXE')==B23)
   if a.steam:
    good(h,'payload-import',payload_kind='base',source_path=str(a.steam));sp=dict(profiles['baseline'],id='steam-comparison');good(h,'prepare-play',profile=sp)
    v=good(h,'verify',id=sp['id']);identity['steam_baseline']=v['fingerprint']
    check('Steam matching engine does not imply identical data fingerprint',v['ok'] and v['fingerprint']!=identity['manual_baseline'])
   if a.cd and a.official:
    good(h,'payload-import',payload_kind='base',source_path=str(a.cd));good(h,'payload-import',payload_kind='official131',source_path=str(a.official))
    for engine in ['1.2','1.31','1.40b23','1.50.26']:
     cp=dict(profiles['baseline'],id='cd-'+engine.replace('.','-'),engine=engine,core='150' if engine=='1.50.26' else '')
     good(h,'prepare-play',profile=cp);g=h.game(cp['id']);v=good(h,'verify',id=cp['id']);check('historical CD lineage retained: '+engine,v['ok'],{'checked':v['checked']})
     if engine in ['1.40b23','1.50.26']:check('historical chain normalizes kernel: '+engine,sha(g/'RKERNEL.COM')==KFIX and sha(g/'ORION131.EXE')==B131)
   # All further repairs should use the chosen canonical source, explicit selection.
   good(h,'payload-import',payload_kind='base',source_path=str(archive))
   check('PRSL and chat remain unavailable',not h.api('state')['prsl_available'] and not h.api('state')['chat_available'])
   good(h,'remove',id='baseline');check('deactivation retains actual game and saves',game.is_dir() and (game/'SAVE1.GAM').read_bytes()==saved and 'baseline' not in h.api('state')['active'])
   good(h,'prepare-play',profile=profiles['baseline']);check('deactivated root can be reconstructed without losing save',(game/'SAVE1.GAM').read_bytes()==saved)
   check('supplied fixture archives unchanged',{p.name:sha(p) for p in archives}==before)
  finally:h.close()
  check('application and user-data locks released',not(data/'manager.lock').exists())
  cli=subprocess.run([str(a.launcher),'--root',str(root),'--data',str(data),'--command','verify'],capture_output=True,text=True,timeout=60)
  check('CLI defaults to portable baseline',cli.returncode==0 and json.loads(cli.stdout)['ok'])
 a.output.parent.mkdir(parents=True,exist_ok=True)
 a.output.write_text(json.dumps(dict(version='0.4.5',platform='linux-amd64',checks=checks,count=len(checks),fingerprints=identity,source_hashes=before,real_owned_files=True,game_executed=False,windows_executed=False,process_test='explicit test double; not DOSBox'),indent=2)+'\n')
if __name__=='__main__':main()
