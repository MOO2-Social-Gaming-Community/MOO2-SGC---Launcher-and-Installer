#!/usr/bin/env python3
"""Real-file regression for 0.4.5's explicit owned-game ZIP path.
Uses a process test double, never DOSBox or MOO2. The source ZIP stays unchanged.
"""
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
import argparse, hashlib, json, subprocess, tempfile, time, urllib.request
p=argparse.ArgumentParser();p.add_argument('--baseline',required=True,type=Path);p.add_argument('--launcher',required=True,type=Path);p.add_argument('--output',required=True,type=Path);a=p.parse_args()
checks=[]
def sha(x):
 with x.open('rb') as f:return hashlib.file_digest(f,'sha256').hexdigest()
def check(name,ok,detail=None):
 if not ok:raise AssertionError(name+': '+str(detail))
 checks.append({'test':name,'result':'PASS','detail':detail});print('PASS',name,flush=True)
before=sha(a.baseline)
with tempfile.TemporaryDirectory(prefix='sgc-zip-source-') as td:
 root=Path(td);data=root/'data';app=root/'app';app.mkdir()
 fake=root/'dosbox';fake.write_text('#!/bin/sh\necho TEST-DOUBLE-NOT-DOSBOX\n');fake.chmod(0o700)
 log=(root/'launcher.log').open('w')
 proc=subprocess.Popen([str(a.launcher),'--root',str(app),'--data',str(data),'--no-browser'],stdout=log,stderr=log)
 try:
  for _ in range(300):
   if (data/'open-launcher.txt').exists():break
   if proc.poll() is not None:raise RuntimeError((root/'launcher.log').read_text())
   time.sleep(.05)
  url,token=(data/'open-launcher.txt').read_text().strip().split('/#')
  def api(route,body=None):
   headers={'Authorization':'Bearer '+token};blob=None
   if body is not None:headers['Content-Type']='application/json';blob=json.dumps(body).encode()
   req=urllib.request.Request(url+'/api/'+route,blob,headers)
   with urllib.request.build_opener(urllib.request.ProxyHandler({})).open(req,timeout=30) as r:return json.load(r)
  def job(action,**kw):
   api('action',dict(action=action,**kw))
   for _ in range(3000):
    j=api('state')['job']
    if not j['busy']:return j
    time.sleep(.03)
   raise TimeoutError(action)
  st=api('state');check('launcher version',st['version']==(ROOT/'VERSION').read_text().strip(),st['version'])
  profile=next(x for x in st['profiles'] if x['id']=='baseline')
  j=job('runtime-select',runtime_path=str(fake));check('test runtime selected',not j['error'],j['error'])
  j=job('prepare-play',profile=profile,source_path=str(a.baseline))
  check('explicit recognized ZIP prepares without legacy payload importer',not j['error'],j['error'])
  st=api('state');check('baseline activated','baseline' in st['active'],st['active'])
  src=st['source'];check('manual b23 source identified',src.get('version')=='1.40b23',src)
  check('source archive untouched',sha(a.baseline)==before)
  # Generated launch config must explicitly fill the available viewport without integer-scaling padding.
  gens=list((data/'environments/baseline').glob('g*'))
  check('managed generation exists',len(gens)==1,len(gens))
  conf=(gens[0]/'dosbox-manager.conf').read_text()
  check('window fill directives generated','viewport = fit' in conf and 'integer_scaling = off' in conf,conf)
  j=job('verify',id='baseline');check('prepared baseline verifies',not j['error'] and j['result']['ok'],j)
  api('action',{'action':'quit'});proc.wait(timeout=15)
 finally:
  if proc.poll() is None:proc.terminate();proc.wait(timeout=15)
  log.close()
a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':(ROOT/'VERSION').read_text().strip(),'checks':checks,'count':len(checks),'source_unchanged':sha(a.baseline)==before,'game_executed':False,'runtime':'test double'},indent=2)+'\n')
