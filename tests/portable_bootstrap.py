#!/usr/bin/env python3
"""Native Linux signed portable bootstrap + old-root update compatibility.
No game, DOSBox, Windows GUI or live GitHub access is exercised here.
"""
from pathlib import Path
import argparse,json,os,subprocess,tempfile,time,urllib.request
ROOT=Path(__file__).resolve().parents[1]
def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--output',required=True,type=Path);p.add_argument('--old-release',type=Path);a=p.parse_args()
 version=(ROOT/'VERSION').read_text().strip();release=ROOT/'release'/version;setup=release/'MOO2-SGC-Setup-linux-amd64';setup.chmod(0o755);checks=[]
 def check(n,ok):
  if not ok:raise AssertionError(n)
  checks.append({'test':n,'result':'PASS'});print('PASS',n,flush=True)
 with tempfile.TemporaryDirectory(prefix='sgc-portable-bootstrap-') as td:
  t=Path(td);root=t/'Games'/'MOO2-SGC'
  def run(exe,root,*args,ok=True):
   r=subprocess.run([str(exe),'--install-root',str(root),*args],capture_output=True,text=True,timeout=90)
   if ok and r.returncode:raise RuntimeError(r.stdout+r.stderr)
   return r
  run(setup,root,'--portable','--command','ensure-installed','--offline',str(release),'--no-launch')
  check('signed portable bootstrap installs launcher', (root/'components/launcher/current.json').is_file())
  before=(root/'components/launcher/current.json').read_bytes()
  run(setup,root,'--command','ensure-installed','--offline',str(t/'not-needed'),'--no-launch')
  check('repeat ensure reuses current verified generation without offline package',before==(root/'components/launcher/current.json').read_bytes())
  check('portable marker set explicitly',json.loads((root/'moo2-sgc-portable.json').read_text())=={'schema':1,'layout':'portable-baseline'})
  check('application installation does not create or modify game',not(root/'game').exists())
  run(setup,root,'--command','verify');check('installed release verifies',True)
  log=(t/'process.log').open('w');proc=subprocess.Popen([str(setup),'--install-root',str(root),'--portable','--command','launch-installed','--no-browser'],stdout=log,stderr=log)
  try:
   marker=root/'userdata/open-launcher.txt'
   for _ in range(500):
    if marker.is_file():break
    if proc.poll() is not None:raise RuntimeError((t/'process.log').read_text())
    time.sleep(.05)
   url,token=marker.read_text().strip().split('/#');client=urllib.request.build_opener(urllib.request.ProxyHandler({}))
   def api(n,body=None):
    h={'Authorization':'Bearer '+token};data=None
    if body is not None:h['Content-Type']='application/json';data=json.dumps(body).encode()
    with client.open(urllib.request.Request(url+'/api/'+n,data,h),timeout=20) as r:return json.load(r)
   s=api('state');check('signed launcher runs portable layout',s['version']==version and s['portable'] and s['default_profile']=='baseline')
   check('actual userdata remains inside selected root',Path(s['data_path'])==root/'userdata')
   check('production signature feed is configured',s['distribution']['production_feed_configured'])
   r=run(setup,root,'--command','verify',ok=False);check('running launcher excludes concurrent setup mutations',r.returncode!=0)
   api('action',{'action':'quit'});proc.wait(timeout=20);check('launcher exits cleanly',proc.returncode==0)
  finally:
   if proc.poll() is None:proc.terminate();proc.wait(timeout=20)
   log.close()
  r=run(setup,root,'--command','launch-installed','--launcher-command','recover-portable','--no-browser');check('setup forwards native recovery command','no interrupted transaction' in r.stdout)
  r=run(setup,root,'--command','launch-installed','--launcher-command','verify',ok=False);check('verification cannot pretend an absent game is prepared',r.returncode!=0)
  check('locks released after failing CLI command',not(root/'installation.lock').exists() and not(root/'userdata/manager.lock').exists())
  if a.old_release:
   old=a.old_release/'MOO2-SGC-Setup-linux-amd64';old.chmod(0o755);legacy=t/'appdata-old'
   run(old,legacy,'--offline',str(a.old_release),'--no-launch')
   sent=legacy/'userdata/owner-note.txt';sent.parent.mkdir(exist_ok=True);sent.write_text('retain existing user data')
   old_pointer=(legacy/'components/launcher/current.json').read_bytes()
   # The unmodified earlier verifier must accept the new same-key release.
   run(setup,legacy,'--command','ensure-installed','--offline',str(release),'--no-launch')
   check('ensure-installed upgrades an older launcher',old_pointer!=(legacy/'components/launcher/current.json').read_bytes())
   run(old,legacy,'--offline',str(release),'--no-launch')
   check('unchanged earlier bootstrap trusts '+version+' with same signing identity',old_pointer!=(legacy/'components/launcher/current.json').read_bytes())
   check('old-root upgrade keeps portable opt-in boundary',not(legacy/'moo2-sgc-portable.json').exists())
   check('old-root upgrade preserves user data',sent.read_text()=='retain existing user data')
   run(setup,legacy,'--command','verify');check('new bootstrap verifies old-root upgraded installation',True)
 a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':version,'count':len(checks),'checks':checks,'platform':'linux-amd64','game_executed':False,'windows_executed':False,'live_network':False},indent=2)+'\n')
if __name__=='__main__':main()
