#!/usr/bin/env python3
"""Actual 0.4.1 launcher + supplied owned data. Local integration, no game execution.
The source archives must be provided locally; they never belong in this repository.
"""
from pathlib import Path
import argparse,hashlib,json,os,subprocess,tempfile,time,urllib.request,zipfile
ROOT=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--base',required=True,type=Path);p.add_argument('--patch',required=True,type=Path);p.add_argument('--launcher',required=True,type=Path);p.add_argument('--output',required=True,type=Path);a=p.parse_args()
checks=[]
def sha(p):
 with p.open('rb') as f:return hashlib.file_digest(f,'sha256').hexdigest()
def check(name,ok,detail=None):
 if not ok:raise AssertionError(name+': '+str(detail))
 checks.append({'test':name,'result':'PASS','detail':detail});print('PASS',name,flush=True)
original=[sha(a.base),sha(a.patch)]
with tempfile.TemporaryDirectory(prefix='sgc-owned-acceptance-') as td:
 tmp=Path(td);data=tmp/'data';folder=tmp/'owned';folder.mkdir();empty=tmp/'empty';empty.mkdir()
 # Export only already-fingerprinted members from the user's known archive.
 index=json.loads((ROOT/'manager/assets/base-files.json').read_text())['files']
 with zipfile.ZipFile(a.base) as z:
  entries={Path(n).name.upper():n for n in z.namelist() if not n.endswith('/')}
  for r in index:
   b=z.read(entries[r['name']]);check_hash=hashlib.sha256(b).hexdigest()
   if check_hash!=r['sha256']:raise ValueError('provided archive differs: '+r['name'])
   (folder/r['name']).write_bytes(b)
 (folder/'SAVE1.GAM').write_bytes(b'EXISTING OWNER SAVE: not real game data')
 (folder/'unrelated.txt').write_text('Do not copy unrelated content')
 check('408 real fingerprinted source files extracted for test',len(index)==408)
 source_before={p.name:sha(p) for p in folder.iterdir()}
 log=(tmp/'launcher.log').open('w');proc=subprocess.Popen([str(a.launcher),'--root',str(empty),'--data',str(data),'--no-browser'],stdout=log,stderr=log)
 try:
  for _ in range(300):
   if (data/'open-launcher.txt').exists():break
   if proc.poll() is not None:raise RuntimeError('launcher failed: '+(tmp/'launcher.log').read_text())
   time.sleep(.05)
  url,token=(data/'open-launcher.txt').read_text().strip().split('/#')
  def api(route,body=None):
   headers={'Authorization':'Bearer '+token};b=None
   if body is not None:headers['Content-Type']='application/json';b=json.dumps(body).encode()
   with urllib.request.build_opener(urllib.request.ProxyHandler({})).open(urllib.request.Request(url+'/api/'+route,b,headers),timeout=20) as r:return json.load(r)
  def job(action,**kw):
   api('action',dict(action=action,**kw))
   for _ in range(2000):
    j=api('state')['job']
    if not j['busy']:return j
    time.sleep(.05)
   raise TimeoutError(action)
  state=api('state');profile=next(p for p in state['profiles'] if p['id']=='community')
  check('actual 0.4.1 process accepts local API',state['version']=='0.4.1')
  j=job('game-import',source_path=str(folder));check('recognized folder imports through real API',not j['error'],j['error'])
  pointer=(data/'cache/base-source.json').read_bytes()
  record=json.loads(pointer);snapshot=data/'cache'/record['filename']
  check('imported private archive verified by content hash',sha(snapshot)==record['sha256'])
  with zipfile.ZipFile(snapshot) as z:
   names=z.namelist();check('snapshot excludes saved games and unrelated files',len(names)==408 and not any(n.upper().endswith('.GAM') or n.endswith('unrelated.txt') for n in names))
  # A modified source must not destroy the previous recognized import.
  engine=folder/'ORION2.EXE';saved=engine.read_bytes();bad=bytearray(saved);bad[10]^=1;engine.write_bytes(bad)
  j=job('game-import',source_path=str(folder));check('modified source rejected',bool(j['error']))
  check('failed import preserves previous source pointer',(data/'cache/base-source.json').read_bytes()==pointer)
  engine.write_bytes(saved)
  j=job('payload-import',payload_kind='patch',source_path=str(a.patch));check('actual local community patch imports',not j['error'],j['error'])
  # Explicit dummy runtime avoids pretending DOSBox was available.
  fixture=tmp/'dosbox';fixture.write_text('#!/bin/sh\necho "ACCEPTANCE TEST DOUBLE -- NOT DOSBox OR MOO2"\nexit 0\n');fixture.chmod(0o700)
  j=job('runtime-select',runtime_path=str(fixture));check('test-double runtime explicitly selected',not j['error'])
  j=job('prepare-play',profile=profile);check('one preparation action builds verified environment',not j['error'],j['error'])
  check('preparation does not launch game',j['result']['ready_for_launch'] and not j['result']['game_started'] and not api('state')['running'])
  state=api('state');game=data/'environments/community'/state['active']['community']/'game'
  check('exact community engine present',sha(game/'ORION150.EXE')=='2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c')
  check('base game files unchanged in patched workspace',all(sha(game/r['name'])==r['sha256'] for r in index))
  check('archived host drive replaced by emulated C root',(game/'orioncd.ini').read_bytes()==b'C:\\\r\n')
  configs=list((game.parent).rglob('*.conf'));text='\n'.join(p.read_text() for p in configs)
  check('DOSBox IRQ matches imported sound setup','irq=5' in text and '[sblaster]' in text, len(configs))
  check('no previous saved game silently imported',not (game/'SAVE1.GAM').exists())
  sentinel=b'PRESERVE TEST SAVE -- NOT LOADED BY GAME';(game/'SAVE1.GAM').write_bytes(sentinel)
  profile['core']='150m';j=job('prepare-play',profile=profile);check('one-action ruleset switch builds independent generation',not j['error'])
  state=api('state');game=data/'environments/community'/state['active']['community']/'game'
  check('test save preserved through ruleset switch',(game/'SAVE1.GAM').read_bytes()==sentinel)
  # Switch explicit source archive back; the fingerprinted folder remains untouched.
  j=job('payload-import',payload_kind='base',source_path=str(a.base));check('original archive import remains supported',not j['error'])
  check('explicit archive supersedes folder-source record',not (data/'cache/base-source.json').exists())
  check('source folder unchanged after all tests',{p.name:sha(p) for p in folder.iterdir()}==source_before)
  check('original supplied archives unchanged',[sha(a.base),sha(a.patch)]==original)
  api('action',{'action':'quit'});proc.wait(timeout=15);check('manager closes cleanly',proc.returncode==0 and not (data/'manager.lock').exists())
 finally:
  if proc.poll() is None:proc.terminate();proc.wait(timeout=15)
  log.close()
a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':'0.4.1','checks':checks,'count':len(checks),'platform':'linux-amd64','real_owned_game_files':True,'source_archives_modified':False,'runtime':'explicit process test double; not DOSBox','game_executed':False,'live_network_download':False},indent=2)+'\n')
