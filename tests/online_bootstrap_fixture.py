#!/usr/bin/env python3
"""Exercise the final binary's ONLINE path through a LOCAL TLS proxy fixture.
The production URLs and signature keys stay embedded and unchanged. Only the
child process trusts a temporary test CA and uses the loopback HTTPS proxy.
This is not a live GitHub/R2 connectivity test. Linux only; no game execution.
"""
from pathlib import Path
import argparse,datetime,hashlib,json,os,socketserver,ssl,subprocess,tempfile,threading,time,urllib.request
from cryptography import x509
from cryptography.x509.oid import NameOID
from cryptography.hazmat.primitives import hashes,serialization
from cryptography.hazmat.primitives.asymmetric import rsa
ROOT=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--output',type=Path,required=True);a=p.parse_args()
v=(ROOT/'VERSION').read_text().strip();release=ROOT/'release'/v;setup=release/'MOO2-SGC-Setup-linux-amd64';setup.chmod(0o755)
manifest=json.loads((release/'manifest.json').read_text());package=next(x for x in manifest['packages'] if x['platform']=='linux-amd64');blob=(release/package['filename']).read_bytes()
checks=[]
def check(name,ok,detail=None):
 if not ok:raise AssertionError(name+': '+str(detail))
 checks.append({'test':name,'result':'PASS','detail':detail});print('PASS',name,flush=True)
with tempfile.TemporaryDirectory(prefix='sgc-https-fixture-') as td:
 temp=Path(td);key=rsa.generate_private_key(public_exponent=65537,key_size=2048);name=x509.Name([x509.NameAttribute(NameOID.COMMON_NAME,'MOO2 local acceptance CA')]);now=datetime.datetime.now(datetime.timezone.utc)
 cert=(x509.CertificateBuilder().subject_name(name).issuer_name(name).public_key(key.public_key()).serial_number(x509.random_serial_number()).not_valid_before(now-datetime.timedelta(minutes=1)).not_valid_after(now+datetime.timedelta(days=1)).add_extension(x509.BasicConstraints(ca=True,path_length=None),True).add_extension(x509.SubjectAlternativeName([x509.DNSName('github.com'),x509.DNSName('release-assets.githubusercontent.com')]),False).sign(key,hashes.SHA256()))
 cp=temp/'ca.pem';kp=temp/'tls.key';cp.write_bytes(cert.public_bytes(serialization.Encoding.PEM));kp.write_bytes(key.private_bytes(serialization.Encoding.PEM,serialization.PrivateFormat.PKCS8,serialization.NoEncryption()))
 ctx=ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER);ctx.load_cert_chain(cp,kp)
 behavior={'interrupt':2,'bad_manifest':False,'bad_package':False};requests=[]
 class Handler(socketserver.StreamRequestHandler):
  def handle(self):
   try:
    line=self.rfile.readline().decode().strip()
    while self.rfile.readline().strip():pass
    if not line.startswith('CONNECT '):return
    self.wfile.write(b'HTTP/1.1 200 Connection established\r\n\r\n');self.wfile.flush()
    with ctx.wrap_socket(self.connection,server_side=True) as s:
     f=s.makefile('rb');first=f.readline().decode();method,path,_=first.strip().split(' ',2);headers={}
     while True:
      line=f.readline().decode().strip()
      if not line:break
      k,val=line.split(':',1);headers[k.lower()]=val.strip()
     requests.append({'path':path,'range':headers.get('range'),'host':headers.get('host')})
     if path.endswith('/manifest.json'):
      body=(release/'manifest.json').read_bytes()+(b' ' if behavior['bad_manifest'] else b'')
     elif path.endswith('/manifest.sig'):body=(release/'manifest.sig').read_bytes()
     elif path.endswith('/'+package['filename']) and not path.startswith('/fixture/'):
      s.sendall(('HTTP/1.1 302 Found\r\nLocation: https://release-assets.githubusercontent.com/fixture/'+package['filename']+'\r\nContent-Length: 0\r\nConnection: close\r\n\r\n').encode());return
     elif path.startswith('/fixture/'):
      body=blob
      if behavior['bad_package']:body=body[:-1]+bytes([body[-1]^1])
     else:s.sendall(b'HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n');return
     start=0;status='200 OK';extra=''
     if headers.get('range'):
      start=int(headers['range'].split('=')[1].split('-')[0]);status='206 Partial Content';extra=f'Content-Range: bytes {start}-{len(body)-1}/{len(body)}\r\n'
     payload=body[start:]
     s.sendall((f'HTTP/1.1 {status}\r\nContent-Length: {len(payload)}\r\n{extra}Connection: close\r\n\r\n').encode())
     if path.startswith('/fixture/') and behavior['interrupt']:
      behavior['interrupt']-=1;s.sendall(payload[:len(payload)//3]);return
     s.sendall(payload)
   except (OSError,ValueError):pass
 class Server(socketserver.ThreadingMixIn,socketserver.TCPServer):daemon_threads=True;allow_reuse_address=True
 server=Server(('127.0.0.1',0),Handler);thread=threading.Thread(target=server.serve_forever,daemon=True);thread.start()
 env=dict(os.environ,HTTPS_PROXY='http://127.0.0.1:'+str(server.server_address[1]),HTTP_PROXY='',ALL_PROXY='',NO_PROXY='',SSL_CERT_FILE=str(cp),SSL_CERT_DIR=str(temp/'empty'))
 root=temp/'installed'
 def boot(path=root,ok=True,*extra):
  r=subprocess.run([str(setup),'--install-root',str(path),'--no-launch',*extra],capture_output=True,text=True,env=env,timeout=90)
  if ok and r.returncode:raise RuntimeError(r.stdout+r.stderr)
  return r
 try:
  r=boot(root,False);check('interrupted HTTPS download refuses activation',r.returncode!=0 and not (root/'components/launcher/current.json').exists())
  check('partial package retained for resume',any((root/'cache/packages').glob('*.part')))
  boot();check('actual compiled bootstrap installs over HTTPS fixture',True)
  check('redirect to approved GitHub asset host followed',any(x['host']=='release-assets.githubusercontent.com' for x in requests))
  check('second attempt sends a nonzero Range',any(x['range'] and x['range']!='bytes=0-' for x in requests))
  boot(root,True,'--command','verify');check('downloaded real launcher passed health and installed-file verification',True)
  prior=(root/'components/launcher/current.json').read_bytes();boot();check('repeat online setup preserves active generation',prior==(root/'components/launcher/current.json').read_bytes())
  behavior['bad_manifest']=True;r=boot(temp/'bad-metadata',False);check('tampered remote metadata never installs',r.returncode!=0 and 'signature mismatch' in r.stderr)
  behavior['bad_manifest']=False;behavior['bad_package']=True;r=boot(temp/'bad-package',False);check('tampered remote launcher package never installs',r.returncode!=0 and not (temp/'bad-package/components/launcher/current.json').exists())
  behavior['bad_package']=False
  # Genuine installed launcher → coordinator check/stage → signed helper restart.
  log=(temp/'launcher.log').open('w');proc=subprocess.Popen([str(setup),'--install-root',str(root),'--command','launch-installed','--no-browser'],env=env,stdout=log,stderr=log)
  try:
   marker=root/'userdata/open-launcher.txt'
   for _ in range(400):
    if marker.exists():break
    time.sleep(.05)
   old_url=marker.read_text().strip();local,token=old_url.split('/#')
   def api(route,body=None,where=None):
    u,t=(local,token) if where is None else where;headers={'Authorization':'Bearer '+t};data=None
    if body is not None:headers['Content-Type']='application/json';data=json.dumps(body).encode()
    # Loopback UI goes direct; only the child application uses the HTTPS proxy.
    with urllib.request.build_opener(urllib.request.ProxyHandler({})).open(urllib.request.Request(u+'/api/'+route,data,headers),timeout=15) as r:return json.load(r)
   def job(action):
    api('action',{'action':action})
    for _ in range(600):
     j=api('state')['job']
     if not j['busy']:
      if j['error']:raise RuntimeError(j['error'])
      return j
     time.sleep(.05)
    raise TimeoutError(action)
   j=job('launcher-check');check('actual launcher checks online signed release',j['result']['release']==v)
   j=job('launcher-stage');check('actual launcher stages online update without activation',not j['result']['activated'])
   marker_save=root/'userdata/acceptance-marker.txt';marker_save.write_text('preserve user data')
   api('action',{'action':'launcher-apply'})
   proc.wait(timeout=20)
   new_url=None
   for _ in range(600):
    if marker.exists():
     candidate=marker.read_text().strip()
     if candidate!=old_url:new_url=candidate;break
    time.sleep(.05)
   check('signed update helper restarts a new real launcher session',new_url is not None)
   check('restart preserved user data',marker_save.read_text()=='preserve user data')
   new_where=tuple(new_url.split('/#'));check('restarted launcher reports '+v,api('state',where=new_where)['version']==v)
   api('action',{'action':'quit'},where=new_where)
   for _ in range(200):
    if not (root/'installation.lock').exists():break
    time.sleep(.05)
   check('restarted updater and launcher release locks',not (root/'installation.lock').exists())
  finally:
   if proc.poll() is None:proc.terminate();proc.wait(timeout=15)
   log.close()
 finally:server.shutdown();server.server_close()
a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps({'version':v,'count':len(checks),'checks':checks,'transport':'local TLS CONNECT proxy; not live GitHub','production_binaries_used':True,'custom_fixture_ca_in_test_process_only':True,'game_execution':False},indent=2)+'\n')
print('TOTAL',len(checks))
