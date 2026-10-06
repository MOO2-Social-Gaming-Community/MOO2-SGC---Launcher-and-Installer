#!/usr/bin/env python3
"""Browser DOM checks against a live API; optional explicit offline transport fixture.
The fixture avoids browser navigation in environments where it is blocked.
"""
from pathlib import Path
import argparse,json,tempfile,time,re
from lineage_acceptance import Harness
from playwright.sync_api import sync_playwright
ROOT=Path(__file__).resolve().parents[1]
def main():
 p=argparse.ArgumentParser();p.add_argument('--launcher',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--screenshot',type=Path);p.add_argument('--offline-dom',action='store_true');a=p.parse_args();checks=[];version=(ROOT/'VERSION').read_text().strip()
 def check(n,ok):
  if not ok:raise AssertionError(n)
  checks.append({'test':n,'result':'PASS'});print('PASS',n)
 with tempfile.TemporaryDirectory(prefix='sgc-ui-') as td:
  root=Path(td);(root/'moo2-sgc-portable.json').write_text('{"schema":1,"layout":"portable-baseline"}');h=Harness(a.launcher,root,root/'userdata')
  try:
   with sync_playwright() as pw:
    browser=pw.chromium.launch(executable_path='/usr/bin/chromium',headless=True,args=['--no-sandbox']);page=browser.new_page(viewport={'width':1440,'height':1080});errors=[];page.on('pageerror',lambda e:errors.append(str(e)))
    if a.offline_dom:
     # Explicit transport test fixture, not an end-to-end native browser request.
     page.set_content('<html><head></head><body></body></html>')
     def transport(route,body):
      try:return {'ok':True,'body':h.api(route,body)}
      except Exception as e:return {'ok':False,'body':{'error':str(e)}}
     page.expose_function('fixtureTransport',transport)
     page.evaluate('''() => {
       Object.defineProperty(window,'sessionStorage',{value:{getItem:()=> 'offline-fixture',setItem:()=>{}},configurable:true});
       window.fetch=async(url,opts={})=>{const r=await window.fixtureTransport(url.replace('/api/',''),opts.body?JSON.parse(opts.body):null);return {ok:r.ok,statusText:r.ok?'OK':'fixture rejection',json:async()=>r.body};};
     }''')
     html=(ROOT/'manager/web/index.html').read_text();html=re.sub(r'<link[^>]+>|<script[^>]+></script>','',html)
     page.set_content(html);page.add_style_tag(content=(ROOT/'manager/web/style.css').read_text());page.add_script_tag(content=(ROOT/'manager/web/app.js').read_text())
    else:page.goto(h.url+'/#'+h.token)
    page.wait_for_function("document.getElementById('version').textContent==='"+version+"'")
    page.wait_for_function("document.getElementById('resolution').textContent.includes('Selection fingerprint')")
    check('actual manager displays '+version,page.locator('#version').inner_text()==version)
    check('running executable and root identity visible',version in page.locator('#application-identity').inner_text() and str(a.launcher) in page.locator('#application-identity').inner_text() and str(root) in page.locator('#application-identity').inner_text())
    check('owned network kernel import is available',page.locator('#payload-kind option[value="kernel"]').count()==1)
    check('unprepared kernel is not shown as verified','Prepare this profile' in page.locator('#kernel-status').inner_text())
    check('future SGC service is explicitly disabled',page.locator('#network-service option[value="sgc"]').is_disabled())
    check('local profile retains a valid default transport choice',page.input_value('#network-service')=='direct')
    page.select_option('#role','host');page.select_option('#network-service','dopefish')
    check('relay selection hides direct-host fields',page.locator('#direct-network-fields').is_hidden() and page.input_value('#port')=='213')
    page.select_option('#network-service','direct')
    check('direct hosting restores nonprivileged default',page.locator('#direct-network-fields').is_visible() and page.input_value('#port')=='21300')
    page.select_option('#role','standalone')
    check('local play disables network service selection',page.locator('#network-service').is_disabled())
    check('portable default is baseline, not community',page.input_value('#profiles')=='baseline' and page.input_value('#engine')=='1.40b23')
    check('root baseline engine is locked',page.locator('#engine').is_disabled())
    values=page.locator('#engine option').evaluate_all('(nodes)=>nodes.map(n=>n.value)');check('four explicit engine choices',values==['1.40b23','1.50.26','1.2','1.31'])
    check('launch disabled before an environment/runtime exists',page.locator('#launch').is_disabled())
    for profile,engine in [('baseline','1.40b23'),('cd-original','1.2'),('original','1.31')]:
     page.select_option('#profiles',profile);page.wait_for_function(f"document.getElementById('resolution').textContent.includes('DOS {engine}')")
     check('profile '+profile+' resolves without 1.50 rules',page.input_value('#engine')==engine and page.locator('#core').is_disabled())
     check('profile '+profile+' cannot select community add-ons',page.locator('#mods input:not(:disabled)').count()==0)
    page.select_option('#profiles','community');page.wait_for_function("document.getElementById('core').disabled===false")
    check('community rulesets re-enable when returning to current',page.input_value('#core')=='150')
    check('official prerequisite local import available',page.locator('#payload-kind option[value="official131"]').count()==1)
    check('source status communicates effective baseline','1.40b23' in page.locator('#source-status').inner_text())
    page.select_option('#profiles','baseline');page.wait_for_function("document.getElementById('engine').value==='1.40b23'")
    if a.screenshot:a.screenshot.parent.mkdir(parents=True,exist_ok=True);page.screenshot(path=str(a.screenshot),full_page=True)
    page.set_viewport_size({'width':430,'height':900});check('narrow viewport has no horizontal document overflow',page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'))
    check('no JavaScript runtime errors',not errors);browser.close()
  finally:h.close()
 a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':version,'count':len(checks),'checks':checks,'real_local_api':True,'native_browser_navigation':not a.offline_dom,'offline_transport_fixture':a.offline_dom,'game_executed':False},indent=2)+'\n')
if __name__=='__main__':main()
