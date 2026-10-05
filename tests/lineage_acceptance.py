#!/usr/bin/env python3
"""Real owned-file and compiled-manager acceptance for 0.4.2.
No game binaries are published. MOO2/DOSBox execution is NOT represented by the
explicit process test double used for launch lifecycle checks.
"""
from pathlib import Path
import argparse,hashlib,json,os,subprocess,tempfile,time,urllib.request,urllib.error,zipfile
ROOT=Path(__file__).resolve().parents[1]
def sha(p):
 with Path(p).open('rb') as f:return hashlib.file_digest(f,'sha256').hexdigest()
class Harness:
 def __init__(self,exe,root,data):
  self.data=data;self.root=root;self.log=(root/'test-server.log').open('w')
  marker=data/'open-launcher.txt'
  if marker.exists():marker.unlink()
  self.proc=subprocess.Popen([str(exe),'--root',str(root),'--data',str(data),'--no-browser'],stdout=self.log,stderr=self.log)
  for _ in range(500):
   if marker.exists():break
   if self.proc.poll() is not None:raise RuntimeError((root/'test-server.log').read_text())
   time.sleep(.05)
  self.url,self.token=marker.read_text().strip().split('/#')
  self.client=urllib.request.build_opener(urllib.request.ProxyHandler({}))
 def api(self,route,body=None):
  headers={'Authorization':'Bearer '+self.token};b=None
  if body is not None:headers['Content-Type']='application/json';b=json.dumps(body).encode()
  with self.client.open(urllib.request.Request(self.url+'/api/'+route,b,headers),timeout=60) as r:return json.load(r)
 def job(self,action,**kw):
  self.api('action',dict(action=action,**kw))
  for _ in range(12000):
   j=self.api('state')['job']
   if not j['busy']:return j
   time.sleep(.05)
  raise TimeoutError(action)
 def close(self):
  if self.proc.poll() is None:
   try:self.api('action',{'action':'quit'});self.proc.wait(timeout=20)
   except Exception:self.proc.terminate();self.proc.wait(timeout=20)
  self.log.close()
 def game(self,id):return self.data/'environments'/id/self.api('state')['active'][id]/'game'

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--assets',type=Path,required=True);p.add_argument('--launcher',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--old-launcher',type=Path);a=p.parse_args()
 checks=[];versions={};source_names=['Master of Orion 2 - steam - v1_40b23.zip','Master of Orion 2 - v1_2(1).zip','_moo2v131_patch.zip','MOO2-1.50.26.zip','Master of Orion 2.zip'];archives={n:a.assets/n for n in source_names};before={n:sha(x) for n,x in archives.items()}
 def check(name,ok,detail=None):
  if not ok:raise AssertionError(name+': '+str(detail))
  checks.append({'test':name,'result':'PASS','detail':detail});print('PASS',name,flush=True)
 def good(h,action,**kw):
  j=h.job(action,**kw)
  if j['error']:raise RuntimeError(action+': '+j['error'])
  return j['result']
 with tempfile.TemporaryDirectory(prefix='sgc-lineage-') as td:
  t=Path(td);empty=t/'empty';empty.mkdir();data=t/'userdata';h=Harness(a.launcher,empty,data)
  fake=t/'dosbox';fake.write_text('#!/bin/sh\necho "TEST DOUBLE ONLY -- NO DOSBOX OR GAME EXECUTION"\nexit 0\n');fake.chmod(0o700)
  try:
   st=h.api('state');check('compiled 0.4.2 starts local API',st['version']=='0.4.2')
   profiles={p['id']:p for p in st['profiles']};check('default current and separate baseline/CD profiles',profiles['community']['engine']=='1.50.26' and profiles['baseline']['engine']=='1.40b23' and profiles['cd-original']['engine']=='1.2')
   good(h,'runtime-select',runtime_path=str(fake));good(h,'payload-import',payload_kind='patch',source_path=str(archives['MOO2-1.50.26.zip']))
   source=t/'Steam copy';source.mkdir();specs=json.loads((ROOT/'manager/assets/source-editions.json').read_text())['editions'];steam_spec=next(s for s in specs if s['id']=='steam-en-140b23')
   with zipfile.ZipFile(archives[source_names[0]]) as z:
    fs={Path(n).name.upper():n for n in z.namelist() if len(n.split('/'))==2 and not n.endswith('/')}
    for r in steam_spec['files']:(source/Path(fs[r['name']]).name).write_bytes(z.read(fs[r['name']]))
   (source/'README.TXT').write_text('Version 1.31 -- deliberately not used for DOS engine detection');(source/'SAVE1.GAM').write_bytes(b'OWNER SAVE NEVER IMPORTED');(source/'unrelated.exe').write_bytes(b'NOT A PROGRAM')
   source_before={p.name:sha(p) for p in source.iterdir()}
   result=good(h,'game-import',source_path=str(source));check('Steam folder recognized by DOS bytes as baseline 1.40b23',result['source_version']=='1.40b23' and result['files']==405,result)
   check('dashboard independently shows source baseline',h.api('state')['source']['version']=='1.40b23')
   receipt=(data/'cache/base-source.json').read_bytes();record=json.loads(receipt)
   with zipfile.ZipFile(data/'cache'/record['filename']) as z:check('source snapshot excludes saves and store executables',len(z.namelist())==405 and not any(n.upper().endswith('.GAM') or n.endswith('unrelated.exe') for n in z.namelist()))
   bad=source/'ANWINFIN.LBX';saved=bad.read_bytes();changed=bytearray(saved);changed[10]^=1;bad.write_bytes(changed)
   j=h.job('game-import',source_path=str(source));check('changed supported Steam asset rejected',bool(j['error']));check('failed import leaves previous known source active',(data/'cache/base-source.json').read_bytes()==receipt);bad.write_bytes(saved)
   result=good(h,'prepare-play',profile=profiles['baseline']);game=h.game('baseline');check('Steam baseline builds without requiring 1.31 package',not (data/'cache/official-1.31.zip').exists() and sha(game/'ORION2.EXE')=='7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5')
   check('baseline source ANWINFIN variant preserved',sha(game/'ANWINFIN.LBX')==sha(source/'ANWINFIN.LBX'))
   check('no source saved game copied into baseline',not (game/'SAVE1.GAM').exists())
   check('baseline native config uses ORION2', 'ORION2.EXE /skipintro' in (game.parent/'dosbox-manager.conf').read_text())
   result=good(h,'prepare-play',profile=profiles['community']);game=h.game('community');mf=json.loads((game.parent/'manifest.json').read_text())
   check('Steam directly upgrades baseline to current fan patch',[s['version'] for s in mf['lineage']]==['1.40b23','1.50.26'])
   check('correct current engine output',sha(game/'ORION150.EXE')=='2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c')
   check('preparation never auto-starts game',not result['game_started'] and not h.api('state')['running'])
   check('generated audio does not require MT-32 ROM',b'SBPRO2.MDI' in (game/'MDI.INI').read_bytes() and b'MT32' not in (game/'MDI.INI').read_bytes())
   check('CD pointer points to the isolated DOS C root',(game/'orioncd.ini').read_bytes()==b'C:\\\r\n')
   sentinel=b'GENERATED SYNTHETIC SAVE FOR PRESERVATION TEST';(game/'SAVE1.GAM').write_bytes(sentinel);(game/'DIG.INI').write_bytes(b'; User audio setting preserved\n')
   previous=game.parent;good(h,'prepare',profile=profiles['community']);game=h.game('community')
   check('same-engine rebuild preserves save and edited audio',(game/'SAVE1.GAM').read_bytes()==sentinel and (game/'DIG.INI').read_bytes()==b'; User audio setting preserved\n')
   check('previous generation remains intact',previous.exists() and (previous/'game/SAVE1.GAM').read_bytes()==sentinel)
   good(h,'verify',id='community');exe=game/'ORION150.EXE';saved=exe.read_bytes();b=bytearray(saved);b[20]^=1;exe.write_bytes(b)
   j=h.job('launch',profile=profiles['community']);check('modified engine cannot launch',bool(j['error']) and not h.api('state')['running']);good(h,'prepare',profile=profiles['community']);check('repair restores exact engine',sha(h.game('community')/'ORION150.EXE')=='2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c')
   launch=good(h,'launch',profile=profiles['community']);check('process lifecycle exercised only by named test double','ORION150.EXE' in launch['configuration'])
   for _ in range(200):
    if not h.api('state')['running']:break
    time.sleep(.05)
   check('launch process log proves test-double scope','TEST DOUBLE ONLY' in Path(launch['log']).read_text())
   check('original Steam folder unchanged in all operations',{p.name:sha(p) for p in source.iterdir()}==source_before)
   j=h.job('prepare',profile=profiles['cd-original']);check('Steam cannot be silently downgraded to CD',bool(j['error']) and 'cannot reconstruct' in j['error'])
   result=good(h,'payload-import',payload_kind='base',source_path=str(archives[source_names[0]]));check('Steam ZIP supported without special filename',result['source_version']=='1.40b23')
   result=good(h,'payload-import',payload_kind='base',source_path=str(archives[source_names[1]]));check('CD archive recognized as 1.2',result['source_version']=='1.2' and result['files']==397)
   result=good(h,'payload-import',payload_kind='official131',source_path=str(archives['_moo2v131_patch.zip']));check('official prerequisite imported by member hashes',result['files']==27)
   for engine,lineage,hash in [
    ('1.2',['1.2'],'558c2bb51354fa48ba3189021374d67873ea0e7b37bab77e32f44eda89e135bc'),
    ('1.31',['1.2','1.31'],'4e11be14217b4aafa1839f333bf5eba037f98b0c44e9e4752c96c464c260419f'),
    ('1.40b23',['1.2','1.31','1.40b23'],'7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5'),
    ('1.50.26',['1.2','1.31','1.40b23','1.50.26'],'2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c')]:
    profile=dict(profiles['community'],id='cd-'+engine.replace('.','-'),name='CD test '+engine,engine=engine,core='150' if engine=='1.50.26' else '',mods=[])
    result=good(h,'prepare-play',profile=profile);game=h.game(profile['id']);mf=json.loads((game.parent/'manifest.json').read_text());v=good(h,'verify',id=profile['id'])
    check('CD builds and verifies '+engine,v['ok'] and [s['version'] for s in mf['lineage']]==lineage,{'lineage':lineage,'checked':v['checked']})
    entry='ORION150.EXE' if engine=='1.50.26' else 'ORION2.EXE';check('CD '+engine+' exact executable identity',sha(game/entry)==hash)
    cfg=(game.parent/'dosbox-manager.conf').read_text();check('CD '+engine+' safe entry/config',entry in cfg and ('/skipintro' not in cfg if engine=='1.2' else True))
    versions[engine]={'lineage':mf['lineage'],'engine_sha256':sha(game/entry),'checked_files':v['checked']}
   result=good(h,'payload-import',payload_kind='base',source_path=str(archives['Master of Orion 2.zip']));check('legacy owned archive still imports',result['source_version']=='1.31')
   old=dict(profiles['community'],id='legacy-current',name='Legacy source current');good(h,'prepare-play',profile=old);game=h.game(old['id']);mf=json.loads((game.parent/'manifest.json').read_text());check('legacy source also constructs baseline before 1.50',[s['version'] for s in mf['lineage']]==['1.31','1.40b23','1.50.26'])
   check('independent PRSL and Chat stay disabled',not h.api('state')['prsl_available'] and not h.api('state')['chat_available'])
   check('all supplied source archives unchanged',{n:sha(x) for n,x in archives.items()}==before)
  finally:h.close()
  if a.old_launcher:
   olddata=t/'old-userdata';oldroot=t/'old-root';oldroot.mkdir();old=Harness(a.old_launcher,oldroot,olddata)
   try:
    good(old,'payload-import',payload_kind='base',source_path=str(archives['Master of Orion 2.zip']));good(old,'payload-import',payload_kind='patch',source_path=str(archives['MOO2-1.50.26.zip']));good(old,'runtime-select',runtime_path=str(fake));prof=next(x for x in old.api('state')['profiles'] if x['id']=='community');good(old,'prepare',profile=prof);oldgame=old.game('community');(oldgame/'SAVE1.GAM').write_bytes(b'MIGRATION SENTINEL NOT A REAL SAVE');oldgen=oldgame.parent
   finally:old.close()
   new=Harness(a.launcher,oldroot,olddata)
   try:
    v=good(new,'verify',id='community');check('0.4.2 verifies actual 0.4.1 existing environment',v['ok'])
    good(new,'prepare-play',profile=prof);newgame=new.game('community');check('old to new reconstruction retains save',(newgame/'SAVE1.GAM').read_bytes()==b'MIGRATION SENTINEL NOT A REAL SAVE');check('old generation retained after migration',oldgen.exists());check('migrated baseline is b23',sha(newgame/'ORION2.EXE')=='7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5')
   finally:new.close()
  check('operation locks released',not (data/'manager.lock').exists())
 report={'version':'0.4.2','platform':'linux-amd64','count':len(checks),'checks':checks,'cd_outputs':versions,'real_owned_files':True,'game_executed':False,'runtime':'explicit process test double, NOT DOSBox','windows_executed':False,'live_patch_download':False,'source_archives_modified':False}
 a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps(report,indent=2)+'\n')
if __name__=='__main__':main()
