"""User-owned fixture integration plus real Chromium UI tests. Never claims emulator/game execution."""
from pathlib import Path
import subprocess,time,json,urllib.request,urllib.error,hashlib,os,sys
from playwright.sync_api import sync_playwright
ROOT=Path(os.environ.get('MOO2_PRIVATE_ROOT','/mnt/data/MOO2_Mod_Manager_0.3.0-alpha.1'))
DATA=Path(os.environ.get('MOO2_TEST_DATA','/mnt/data/dev/integration-data')); DATA.mkdir(parents=True,exist_ok=True); EVIDENCE=Path(__file__).parent/'evidence'
EXE=os.environ.get('MOO2_MANAGER_EXE','/mnt/data/dev/moo2-manager'); results=[]
def check(name,condition,detail=None):
 if not condition: raise AssertionError(name+': '+str(detail))
 results.append({'test':name,'result':'PASS','detail':detail});print('PASS',name,flush=True)
def digest(p):
 with open(p,'rb') as f:return hashlib.file_digest(f,'sha256').hexdigest()
base_before=digest(ROOT/'payloads/base.zip');patch_before=digest(ROOT/'payloads/patch-1.50.26.zip')
log=open(EVIDENCE/'server-integration.log','w')
proc=subprocess.Popen([EXE,'--root',str(ROOT),'--data',str(DATA),'--no-browser'],stdout=log,stderr=log)
try:
 for _ in range(100):
  if (DATA/'open-launcher.txt').exists():break
  if proc.poll() is not None:raise RuntimeError('server exited')
  time.sleep(.1)
 private=(DATA/'open-launcher.txt').read_text().strip();base,token=private.split('/#');base+='/'
 def api(route,body=None,authorized=True):
  b=None if body is None else json.dumps(body).encode();headers={'Content-Type':'application/json'}
  if authorized:headers['Authorization']='Bearer '+token
  req=urllib.request.Request(base+'api/'+route,b,headers)
  try:
   with urllib.request.urlopen(req,timeout=10) as r:return r.status,json.load(r)
  except urllib.error.HTTPError as e:return e.code,json.load(e)
 def job(action,**kw):
  status,r=api('action',dict(action=action,**kw));assert status==202,r
  for _ in range(600):
   _,state=api('state');j=state['job']
   if not j['busy']:return j
   time.sleep(.1)
  raise TimeoutError(action)
 _,state=api('state')
 if 'community' not in state['active']:
  initial=next(p for p in state['profiles'] if p['id']=='community');j=job('prepare',profile=initial);assert not j['error'],j
  _,state=api('state')
 check('actual HTTP API starts',state['version']==(Path(__file__).parents[1]/'VERSION').read_text().strip())
 check('distribution providers visible',state['distribution']['primary']=='github' and state['distribution']['secondary']=='cloudflare-r2')
 check('PRSL remains independently unavailable',not state['prsl_available'])
 status,_=api('state',authorized=False);check('actual unauthenticated HTTP request rejected',status==401)
 with sync_playwright() as pw:
  browser=pw.chromium.launch(executable_path=os.environ.get('MOO2_CHROMIUM','/usr/bin/chromium'),headless=True,args=['--no-sandbox'])
  page=browser.new_page(viewport={'width':1440,'height':1080},device_scale_factor=1)
  js_errors=[];page.on('pageerror',lambda e:js_errors.append(str(e)))
  # Browser policy blocks all URL navigation in this environment. Render offline DOM fixtures,
  # while actual HTTP/API behavior is tested separately with urllib above and below.
  p0=next(x for x in state['profiles'] if x['id']=='community').copy()
  p0.update(core='150',engine='1.50.26',mods=[],role='standalone')
  _,res0=api('resolve',p0)
  pm={**p0,'core':'150m'};_,resm=api('resolve',pm)
  po={**p0,'core':'','engine':'1.31'};_,reso=api('resolve',po)
  fixtures={'state':state,'res0':res0,'resm':resm,'reso':reso}
  page.set_content('<html><head></head><body></body></html>')
  page.evaluate("""f=>{
   const store={getItem:()=> 'testtoken',setItem:()=>{}};
   Object.defineProperty(window,'sessionStorage',{value:store,configurable:true});
   window.fetch=async (url,opt={})=>{
    let data;let ok=true;
    if(url==='/api/state')data=f.state;
    else if(url==='/api/addresses')data=['192.0.2.10'];
    else if(url==='/api/resolve'){
     const p=JSON.parse(opt.body);
     if(p.core==='150m'&&p.mods.includes('MAP5_GM1')){ok=false;data={error:'Map conflict: MAP1_150m and MAP5_GM1'};}
     else data=p.engine==='1.31'?f.reso:p.core==='150m'?f.resm:f.res0;
    }else{ok=false;data={error:'Offline browser fixture: action unavailable'};}
    return {ok,statusText:ok?'OK':'fixture rejection',json:async()=>structuredClone(data)};
   };
  }""",fixtures)
  html=(Path(__file__).parent/'web/index.html').read_text()
  import re
  html=re.sub(r'<link[^>]+>|<script[^>]+></script>','',html)
  page.set_content(html)
  page.add_style_tag(content=(Path(__file__).parent/'web/style.css').read_text())
  page.add_script_tag(content=(Path(__file__).parent/'web/app.js').read_text())
  page.wait_for_function("document.querySelector('#profiles').options.length===3")
  page.wait_for_function("document.querySelector('#resolution').textContent.includes('Selection fingerprint')")
  check('offline Chromium startup profile selector renders',page.locator('#profiles option').count()==3)
  check('PRSL and Chat visibly unavailable',page.locator('.feature input:disabled').count()==2)
  check('offline DOM fixture renders without a public URL',page.url=='about:blank')
  page.select_option('#core','150m');page.wait_for_function("document.querySelector('#resolution').textContent.includes('MIRROR_HW')")
  check('UI shows automatic multiplayer dependencies','MIRROR_HW' in page.locator('#resolution').inner_text())
  page.locator('summary').click()
  page.locator('#mods input[data-mod="MAP5_GM1"]').check();page.wait_for_function("document.querySelector('#resolution').textContent.includes('conflict')")
  check('UI exposes conflicting map selection','conflict' in page.locator('#resolution').inner_text())
  page.locator('#mods input[data-mod="MAP5_GM1"]').uncheck()
  page.select_option('#engine','1.31');page.wait_for_function("document.querySelector('#core').disabled")
  check('original engine disables community mod controls',page.locator('#core').is_disabled())
  page.select_option('#engine','1.50.26');page.wait_for_function("!document.querySelector('#core').disabled")
  page.select_option('#profiles','community');page.wait_for_timeout(200)
  page.locator('details').evaluate('(e)=>e.open=false')
  page.screenshot(path=str(EVIDENCE/'launcher-desktop.png'),full_page=True)
  page.set_viewport_size({'width':390,'height':844});page.screenshot(path=str(EVIDENCE/'launcher-narrow.png'),full_page=True)
  check('narrow UI has no horizontal overflow',page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'))
  check('Chromium reports no JavaScript exceptions',len(js_errors)==0,js_errors)
  browser.close()
 _,state=api('state');p=next(p for p in state['profiles'] if p['id']=='community')
 path=DATA/'environments'/'community'/state['active']['community']/'game'
 sentinel=b'USER SAVE PRESERVATION TEST -- NOT A REAL SAVE'
 (path/'SAVE1.GAM').write_bytes(sentinel)
 (path/'150/USER.CFG').write_text('# User preference survives repair\n',encoding='ascii')
 p['core']='150m'
 j=job('prepare',profile=p);check('actual mod-switch rebuild succeeds',not j['error'],j.get('result'))
 _,state=api('state');newgen=state['active']['community'];game=DATA/'environments'/'community'/newgen/'game'
 check('profile save survives mod-switch rebuild',(game/'SAVE1.GAM').read_bytes()==sentinel)
 check('user CFG survives mod-switch rebuild',(game/'150/USER.CFG').read_text()=='# User preference survives repair\n')
 check('selected configuration written without PRSL','enable 150m;' in (game/'150/ENABLE.CFG').read_text() and 'enable prsl' not in (game/'150/ENABLE.CFG').read_text().lower())
 j=job('verify',id='community');check('rebuilt workspace verifies',j['result']['ok'])
 engine=game/'ORION150.EXE';original=engine.read_bytes();changed=bytearray(original);changed[10]^=1;engine.write_bytes(changed)
 j=job('verify',id='community');check('engine corruption is detected',not j['result']['ok'])
 j=job('launch',profile=p);check('corrupt engine launch refused',bool(j['error']) and 'verification failed' in j['error'].lower())
 j=job('prepare',profile=p);check('repair restores pristine engine without reverse patching',not j['error'])
 _,state=api('state');repairgen=state['active']['community'];game=DATA/'environments'/'community'/repairgen/'game'
 check('repair retained saves',(game/'SAVE1.GAM').read_bytes()==sentinel)
 check('repair restored exact engine hash',digest(game/'ORION150.EXE')=='2db296e052419250d21866f7c23ac2978a33b3c05b451a9516f06599b91c3f5c')
 # A corrupt cached dependency must not replace a good active generation.
 cache=DATA/'cache'/'patch-1.50.26.zip';cache.write_bytes(b'bad package')
 j=job('prepare',profile=p);check('bad cached patch rejected',bool(j['error']) and 'checksum mismatch' in j['error'])
 _,state=api('state');check('failed install leaves active generation unchanged',state['active']['community']==repairgen);cache.unlink()
 j=job('diagnostics',id='community');check('diagnostic export succeeds',not j['error'])
 report=Path(j['result']['path']).read_text();check('diagnostics omit access token and save content',token not in report and sentinel.decode() not in report)
 j=job('remove',id='community');check('deactivation succeeds',not j['error'])
 check('deactivation does not delete saves',(game/'SAVE1.GAM').read_bytes()==sentinel)
 j=job('activate',id='community',generation=repairgen);check('verified retained generation reactivates',not j['error'])
 _,state=api('state');original_profile=next(x for x in state['profiles'] if x['id']=='original')
 j=job('prepare',profile=original_profile);check('original 1.31 environment installs without patch',not j['error'])
 _,state=api('state');stock=DATA/'environments'/'original'/state['active']['original']/'game'
 check('patch absent from original environment',not (stock/'ORION150.EXE').exists() and not (stock/'150').exists())
 check('original source saves not imported',not list(stock.glob('*.GAM')))
 # Verify process launch/config/log lifecycle with an explicitly labelled TEST DOUBLE only.
 fixture=DATA/'test-fixtures'/'dosbox';fixture.parent.mkdir(exist_ok=True);fixture.write_text('#!/bin/sh\necho "TEST DOUBLE - NOT DOSBox - arguments: $*"\nexit 0\n');fixture.chmod(0o700)
 j=job('runtime-select',runtime_path=str(fixture));check('explicit runtime selection persists',not j['error'])
 j=job('launch',profile=p);check('process lifecycle test double starts (NOT game execution)',not j['error'])
 time.sleep(.3);_,state=api('state');check('child process is reaped and status recorded',not state['running'] and 'exited normally' in state['last_exit'])
 (DATA/'settings.json').unlink();fixture.unlink()
 check('source base archive unchanged',digest(ROOT/'payloads/base.zip')==base_before)
 check('source patch archive unchanged',digest(ROOT/'payloads/patch-1.50.26.zip')==patch_before)
 status,r=api('action',{'action':'quit'});check('explicit launcher shutdown accepted',status==200)
 proc.wait(timeout=5);check('manager process lock removed on clean shutdown',not (DATA/'manager.lock').exists())
finally:
 if proc.poll() is None:proc.terminate();
 try:proc.wait(timeout=5)
 except subprocess.TimeoutExpired:proc.kill()
 log.close()
 (EVIDENCE/'integration-tests.json').write_text(json.dumps({'tests':results,'count':len(results),'game_execution':False,'runtime_network_download':False,'browser':'offline Chromium DOM fixture via Playwright; live browser-to-server navigation blocked by environment policy','os':'Linux x86_64'},indent=2)+'\n')
print('TOTAL',len(results))
