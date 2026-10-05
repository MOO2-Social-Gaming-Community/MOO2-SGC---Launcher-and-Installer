#!/usr/bin/env python3
"""Native setup + launcher smoke test; no game data, emulator or live network.
This is also run on GitHub's Windows/macOS runners after the owner's push.
"""
from pathlib import Path
import argparse,json,os,platform,subprocess,tempfile,time,urllib.request
ROOT=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--output',type=Path);a=p.parse_args()
v=(ROOT/'VERSION').read_text().strip();release=ROOT/'release'/v
system={'Windows':'windows','Darwin':'darwin','Linux':'linux'}[platform.system()]
arch={'x86_64':'amd64','AMD64':'amd64','arm64':'arm64','aarch64':'arm64'}[platform.machine()]
name='MOO2-SGC-Setup.exe' if system=='windows' else f'MOO2-SGC-Setup-{system}-{arch}'
setup=release/name
if system!='windows':setup.chmod(0o755)
checks=[]
def check(name,ok):
    if not ok:raise AssertionError(name)
    checks.append({'test':name,'result':'PASS'});print('PASS',name,flush=True)
with tempfile.TemporaryDirectory(prefix='moo2-sgc-native-') as tmp:
    root=Path(tmp)/'install';env=dict(os.environ)
    if system=='windows':env['APPDATA']=str(Path(tmp)/'appdata')
    def boot(*args,ok=True):
        r=subprocess.run([str(setup),'--install-root',str(root),*args],capture_output=True,text=True,timeout=90,env=env)
        if ok and r.returncode:raise RuntimeError(r.stdout+r.stderr)
        return r
    check('version has exactly three components',boot('--version').stdout.strip()==v)
    check('online endpoints embedded','development=false' in boot('--trust-status').stdout and 'endpoints=2' in boot('--trust-status').stdout)
    boot('--offline',str(release),'--no-launch');check('signed launcher installed natively',True)
    boot('--command','verify');check('installed files verified',True)
    pointer=root/'components/launcher/current.json';before=pointer.read_bytes()
    boot('--offline',str(release),'--no-launch');check('repeat install idempotent',before==pointer.read_bytes())
    current=json.loads(pointer.read_text())['current'];helper=root/'components/launcher/versions'/current/'files'/('MOO2-SGC-Setup.exe' if system=='windows' else 'MOO2-SGC-Setup')
    check('signed installed update helper executes',subprocess.check_output([str(helper),'--version'],text=True).strip()==v)
    logfile=(Path(tmp)/'launcher.log').open('w')
    proc=subprocess.Popen([str(setup),'--command','launch-installed','--install-root',str(root),'--no-browser'],stdout=logfile,stderr=logfile,env=env)
    try:
        marker=root/'userdata/open-launcher.txt'
        for _ in range(400):
            if marker.exists():break
            if proc.poll() is not None:raise RuntimeError('launcher stopped')
            time.sleep(.05)
        local,token=marker.read_text().strip().split('/#')
        def api(route,body=None):
            headers={'Authorization':'Bearer '+token};data=None
            if body is not None:headers['Content-Type']='application/json';data=json.dumps(body).encode()
            with urllib.request.urlopen(urllib.request.Request(local+'/api/'+route,data,headers),timeout=15) as r:return json.load(r)
        state=api('state');check('actual native launcher HTTP API',state['version']==v)
        check('PRSL and Chat stay independently disabled',not state['prsl_available'] and not state['chat_available'])
        check('signed online feed configured',state['distribution']['production_feed_configured'])
        api('action',{'action':'quit'});proc.wait(timeout=15);check('clean launcher shutdown',proc.returncode==0)
    finally:
        if proc.poll() is None:proc.terminate();proc.wait(timeout=15)
        logfile.close()
    check('all operation locks released',not (root/'installation.lock').exists() and not (root/'userdata/manager.lock').exists())
result={'platform':system+'-'+arch,'version':v,'checks':checks,'count':len(checks),'network_used':False,'game_executed':False}
if a.output:a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
