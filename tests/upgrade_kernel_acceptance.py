#!/usr/bin/env python3
"""Private real-file regression: old 0.4.2 workspace -> current launcher/kernel.
Executes actual Linux managers. The DOSBox executable is a clearly labelled test
substitute: no claim about rendering, engine execution or Internet connectivity.
"""
from pathlib import Path
import argparse,json,subprocess,tempfile,time,zipfile
from lineage_acceptance import Harness,sha
ROOT=Path(__file__).resolve().parents[1]
KFIX='18e8781f8ce64516e60b2947b487e8999f7d940b25af971359c7e7dea0fe9d97'

def main():
 p=argparse.ArgumentParser(description=__doc__)
 for n in ('launcher','old-launcher','steam','kernel','patch','output'):p.add_argument('--'+n,type=Path,required=True)
 a=p.parse_args();version=(ROOT/'VERSION').read_text().strip();checks=[]
 def check(name,ok,detail=None):
  if not ok:raise AssertionError(name+': '+str(detail))
  checks.append(dict(test=name,result='PASS',detail=detail));print('PASS',name,flush=True)
 def good(h,action,**kw):
  j=h.job(action,**kw)
  if j['error']:raise RuntimeError(action+': '+j['error'])
  return j['result']
 before={str(x):sha(x) for x in [a.steam,a.kernel,a.patch]}
 with tempfile.TemporaryDirectory(prefix='sgc-upgrade-kernel-') as td:
  root=Path(td)/'old AppData application';root.mkdir();data=root/'userdata'
  fake=root/'test-double'/'dosbox';fake.parent.mkdir();fake.write_text('#!/bin/sh\necho "TEST DOUBLE: not DOSBox, no MOO2 executed"\nprintf "CWD=%s\\n" "$PWD"\nprintf "ARG=%s\\n" "$@"\nsleep 0.1\n');fake.chmod(0o700)
  old=Harness(a.old_launcher,root,data)
  try:
   check('unmodified old manager actually reports 0.4.2',old.api('state')['version']=='0.4.2')
   good(old,'payload-import',payload_kind='base',source_path=str(a.steam))
   good(old,'payload-import',payload_kind='patch',source_path=str(a.patch))
   good(old,'runtime-select',runtime_path=str(fake))
   profile=next(x for x in old.api('state')['profiles'] if x['id']=='community')
   good(old,'prepare',profile=profile);old_game=old.game('community')
   check('reproduced 0.4.2 environment omitting RKERNEL.COM',not (old_game/'RKERNEL.COM').exists())
   check('old manager incorrectly verified missing-kernel environment',good(old,'verify',id='community')['ok'])
   good(old,'launch',profile=profile)
   for _ in range(200):
    if not old.api('state')['running']:break
    time.sleep(.05)
   check('old launch gate permitted missing kernel (test double only)',not old.api('state')['running'])
   save=b'SYNTHETIC SAVE PRESERVATION TEST, NOT A MOO2 SAVE';(old_game/'SAVE1.GAM').write_bytes(save)
   original_pointer=(data/'active/community.json').read_bytes()
  finally:old.close()
  h=Harness(a.launcher,root,data)
  try:
   st=h.api('state');check('new actual executable identity visible',st['version']==version and st['application']['version']==version and st['application']['executable']==str(a.launcher))
   check('upgrade did not move or replace game automatically',(data/'active/community.json').read_bytes()==original_pointer)
   check('UI network preflight shows exact missing-kernel path',not st['network_preflight']['community']['ok'] and st['network_preflight']['community']['kernel_path']==str(old_game/'RKERNEL.COM'))
   j=h.job('launch',profile=profile);check('new manager blocks old environment before process start',bool(j['error']) and str(old_game/'RKERNEL.COM') in j['error'] and not h.api('state')['running'])
   check('failed preflight saved for diagnosis',(data/'logs/last-network-preflight.json').exists())
   imported=good(h,'payload-import',payload_kind='kernel',source_path=str(a.kernel))
   check('archived pre-fix kernel normalized into private verified cache',imported['sha256']==KFIX and sha(Path(imported['cached']))==KFIX and not imported['source_modified'])
   check('kernel import did not alter active old game',not(old_game/'RKERNEL.COM').exists())
   good(h,'prepare',profile=profile);game=h.game('community')
   check('repair committed a separate new generation',game!=old_game and old_game.is_dir())
   check('repaired community contains exact canonical network kernel',sha(game/'RKERNEL.COM')==KFIX)
   check('full repaired workspace verifies',good(h,'verify',id='community')['ok'])
   check('same-engine saved bytes preserved',(game/'SAVE1.GAM').read_bytes()==save and (old_game/'SAVE1.GAM').read_bytes()==save)
   check('UI now reports network kernel verified',h.api('state')['network_preflight']['community']['ok'])
   with zipfile.ZipFile(a.steam) as z:
    f=next(n for n in z.namelist() if Path(n).name.upper()=='RKERNEL.COM')
    check('normalized archived driver equals supplied Steam driver byte-for-byte',(game/'RKERNEL.COM').read_bytes()==z.read(f))
   for service in ['direct','dopefish']:
    for role in ['host','join']:
     p=dict(profile,role=role,network_service=service,host='192.168.1.25',port=21300)
     result=good(h,'launch',profile=p)
     for _ in range(200):
      if not h.api('state')['running']:break
      time.sleep(.05)
     cfg=result['configuration'];last=json.loads((data/'logs/last-launch.json').read_text())
     check(service+' '+role+' uses same directory for runtime/kernel/game',last['working_directory']==str(game) and last['kernel']['kernel_path']==str(game/'RKERNEL.COM') and Path(last['game_executable']).parent==game and last['launcher_version']==version)
     expected='IPXNET CONNECT moo2.thedopefish.com 213' if service=='dopefish' else ('IPXNET STARTSERVER 21300' if role=='host' else 'IPXNET CONNECT 192.168.1.25 21300')
     check(service+' '+role+' produces intended IPX command',expected in cfg and ('STARTSERVER' not in cfg if service=='dopefish' else True))
     check(service+' '+role+' executed only the labelled test double','TEST DOUBLE: not DOSBox' in Path(result['log']).read_text())
   k=game/'RKERNEL.COM';k.unlink();j=h.job('launch',profile=profile)
   check('deleted kernel refused even for standalone role',bool(j['error']) and not h.api('state')['running'])
   good(h,'prepare',profile=profile);game=h.game('community');k=game/'RKERNEL.COM';b=k.read_bytes();k.write_bytes(b[:-1]+bytes([b[-1]^1]))
   j=h.job('launch',profile=profile);check('modified kernel refused before launch',bool(j['error']) and not h.api('state')['running'])
   bad=root/'RKERNEL.COM';bad.write_bytes(b'not a kernel');cache_before=sha(Path(imported['cached']))
   j=h.job('payload-import',payload_kind='kernel',source_path=str(bad));check('unknown manual kernel rejected without damaging verified cache',bool(j['error']) and sha(Path(imported['cached']))==cache_before)
   good(h,'prepare',profile=profile);game=h.game('community')
   check('second repair restores driver without losing save',sha(game/'RKERNEL.COM')==KFIX and (game/'SAVE1.GAM').read_bytes()==save)
   check('PRSL and future Chat remain disabled',not h.api('state')['prsl_available'] and not h.api('state')['chat_available'])
   check('future SGC online cannot be selected',bool(h.job('launch',profile=dict(profile,role='host',network_service='sgc'))['error']))
   check('all supplied archive bytes unchanged',{str(x):sha(x) for x in [a.steam,a.kernel,a.patch]}==before)
  finally:h.close()
 a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps(dict(version=version,count=len(checks),checks=checks,real_old_launcher='0.4.2',real_owned_files=True,real_game_executed=False,windows_executed=False,scope='actual Linux managers; DOSBox test double'),indent=2)+'\n')
if __name__=='__main__':main()
