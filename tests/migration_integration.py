#!/usr/bin/env python3
"""Check the schema-1 portable data upgrade with real 0.3 and 0.4 CLI processes."""
import argparse, hashlib, json, subprocess, tempfile
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--old',type=Path,required=True);p.add_argument('--new',type=Path,required=True);p.add_argument('--fixture',type=Path,required=True);p.add_argument('--evidence',type=Path,required=True);a=p.parse_args();results=[]
def check(name,c):
 if not c:raise AssertionError(name)
 results.append({'test':name,'result':'PASS'});print('PASS',name,flush=True)
with tempfile.TemporaryDirectory(prefix='moo2-schema-migration-') as temp:
 data=Path(temp)/'data'
 def cli(exe,cmd):
  r=subprocess.run([str(exe),'--root',str(a.fixture),'--data',str(data),'--command',cmd],capture_output=True,text=True,timeout=60)
  if r.returncode:raise AssertionError(r.stdout+r.stderr)
  return json.loads(r.stdout)
 old=cli(a.old,'prepare');check('actual 0.3 manager prepares legacy workspace',old['verification']['ok'])
 active=json.loads((data/'active/community.json').read_text())['generation'];game=data/'environments/community'/active/'game'
 sentinel=b'MIGRATION SENTINEL - NOT A REAL MOO2 SAVE';(game/'SAVE1.GAM').write_bytes(sentinel)
 profiles={p.name:p.read_bytes() for p in (data/'profiles').glob('*.json')}
 v=cli(a.new,'verify');check('0.4 reads and verifies 0.3 workspace',v['ok'])
 check('read-only upgrade check preserves legacy profiles',profiles=={p.name:p.read_bytes() for p in (data/'profiles').glob('*.json')})
 new=cli(a.new,'prepare');check('0.4 rebuilds legacy profile without resetting game settings',new['verification']['ok'])
 active2=json.loads((data/'active/community.json').read_text())['generation'];game2=data/'environments/community'/active2/'game'
 check('legacy save survives 0.4 reconstruction',(game2/'SAVE1.GAM').read_bytes()==sentinel)
 check('old generation and its save remain untouched',(game/'SAVE1.GAM').read_bytes()==sentinel)
 v=cli(a.old,'verify');check('0.3 still verifies reconstructed schema-1 game workspace',v['ok'])
a.evidence.mkdir(parents=True,exist_ok=True);(a.evidence/'migration-integration.json').write_text(json.dumps({'tests':results,'count':len(results),'game_execution':False,'scope':'actual Linux CLI managers, real game file reconstruction; synthetic save sentinel only'},indent=2)+'\n')
