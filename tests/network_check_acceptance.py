#!/usr/bin/env python3
"""Actual manager API/process tests. Uses an explicitly labelled DOSBox test double.
No public UDP connection, DOSBox, or MOO2 engine is executed by this test.
"""
from pathlib import Path
import argparse,json,tempfile,time,urllib.error,subprocess
from lineage_acceptance import Harness
ROOT=Path(__file__).resolve().parents[1]
def main():
 p=argparse.ArgumentParser();p.add_argument('--launcher',type=Path,required=True);p.add_argument('--output',type=Path,required=True);a=p.parse_args();checks=[]
 def check(n,ok):
  if not ok:raise AssertionError(n)
  checks.append({'test':n,'result':'PASS'});print('PASS',n,flush=True)
 def good(h,act,**kw):
  j=h.job(act,**kw)
  if j['error']:raise RuntimeError(j['error'])
  return j['result']
 with tempfile.TemporaryDirectory(prefix='sgc-047-diagnostic-') as td:
  root=Path(td);data=root/'userdata';h=Harness(a.launcher,root,data)
  try:
   check('fresh selection remains baseline',h.api('state')['selected_profile_id']=='baseline')
   settings=b'{"runtime_path":"custom prior runtime"}\n';(data/'settings.json').write_bytes(settings)
   h.api('selection',{'id':'community'})
   check('selection remembered without saving game settings',h.api('state')['selected_profile_id']=='community')
   check('selection preserves runtime preferences',(data/'settings.json').read_bytes()==settings)
   before={p.name:p.read_bytes() for p in (data/'profiles').glob('*.json')}
   check('preference selection does not build an environment',not list((data/'environments').iterdir()))
   h.close();h=Harness(a.launcher,root,data)
   check('selection survives real launcher restart/new local port',h.api('state')['selected_profile_id']=='community')
   check('restart preserves all original profile files',before=={p.name:p.read_bytes() for p in (data/'profiles').glob('*.json')})
   fake=root/'dosbox';fake.write_text('#!/bin/sh\necho "TEST DOUBLE ONLY - no DOSBox, no MOO2, no network"\nprintf "CWD=%s\\n" "$PWD"\nprintf "ARG=%s\\n" "$@"\nsleep 1\n');fake.chmod(0o700)
   good(h,'runtime-select',runtime_path=str(fake));check('runtime selection does not reset remembered profile',h.api('state')['selected_profile_id']=='community')
   out=good(h,'check-dopefish');cfg=Path(out['config_path']).read_text();st=h.api('state')
   check('diagnostic process identified separately from game',st['running_kind']=='network-check')
   check('diagnostic runs without owned source or prepared game',not list((data/'environments').iterdir()) and not list((data/'cache').iterdir()))
   check('diagnostic uses dedicated directory outside game',Path(out['working_directory']).parent==data/'network-checks')
   check('diagnostic uses independent pinned config arguments',out['arguments']==['--noprimaryconf','--conf',out['config_path']])
   check('diagnostic connects and checks actual shell status commands','IPXNET CONNECT moo2.thedopefish.com 213' in cfg and 'IPXNET STATUS' in cfg)
   check('diagnostic neither mounts nor runs game',not any(l.strip().upper().startswith(('MOUNT ','IMGMOUNT ','ORION')) for l in cfg.splitlines()))
   check('diagnostic remains open for manual observation',not any(l.strip().upper()=='EXIT' for l in cfg.splitlines()))
   check('no connectivity success fabricated',out['connection_verified'] is False and out['game_started'] is False)
   try:h.api('action',{'action':'check-dopefish'});blocked=False
   except urllib.error.HTTPError as e:blocked=e.code==409
   check('second diagnostic cannot compete for session',blocked)
   try:h.api('action',{'action':'quit'});blocked=False
   except urllib.error.HTTPError as e:blocked=e.code==409
   check('launcher remains locked while diagnostic is open',blocked)
   for _ in range(200):
    if not h.api('state')['running']:break
    time.sleep(.05)
   check('diagnostic releases process gate',not h.api('state')['running'] and h.api('state')['running_kind']=='')
   check('exit is not reported as connection success','exit code is not proof of connection' in h.api('state')['last_exit'])
   check('host process log proves test-double scope','TEST DOUBLE ONLY' in Path(out['log']).read_text())
   report=good(h,'diagnostics',id='community')['report'];check('export includes independent diagnostic record',report['last_network_check']['config_path']==out['config_path'])
   check('diagnostic preserved source/profile selection',h.api('state')['selected_profile_id']=='community' and before=={p.name:p.read_bytes() for p in (data/'profiles').glob('*.json')})
  finally:h.close()
  run=subprocess.run([str(a.launcher),'--root',str(root),'--data',str(data),'--command','check-dopefish','--no-browser'],capture_output=True,text=True,timeout=30)
  check('CLI diagnostic waits for process and returns without starting game',run.returncode==0 and '"game_started": false' in run.stdout)
  check('CLI diagnostic releases data and app locks',not(data/'manager.lock').exists() and not(data/'application/installation.lock').exists())
 a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':(ROOT/'VERSION').read_text().strip(),'count':len(checks),'checks':checks,'actual_manager':True,'process_test_double':True,'live_connection_tested':False,'game_executed':False},indent=2)+'\n')
if __name__=='__main__':main()
