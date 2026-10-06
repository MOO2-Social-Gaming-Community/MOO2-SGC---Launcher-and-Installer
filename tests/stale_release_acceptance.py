#!/usr/bin/env python3
"""Real signed binaries + old release through a local HTTPS proxy fixture.
Reproduces setup 0.4.5 selecting launcher 0.4.2, then verifies the version floor
and pinned fallback in the current installer. No external service is contacted.
"""
from pathlib import Path
import argparse,datetime,json,os,socketserver,ssl,subprocess,tempfile,threading,time,urllib.request,zipfile
from cryptography import x509
from cryptography.x509.oid import NameOID
from cryptography.hazmat.primitives import hashes,serialization
from cryptography.hazmat.primitives.asymmetric import rsa
ROOT=Path(__file__).resolve().parents[1]

def main():
 p=argparse.ArgumentParser(description=__doc__)
 for n in ['old-repository','old-setup','output']:p.add_argument('--'+n,type=Path,required=True)
 a=p.parse_args();version=(ROOT/'VERSION').read_text().strip();release=ROOT/'release'/version
 setup=release/'MOO2-SGC-Setup-linux-amd64';setup.chmod(0o700);a.old_setup.chmod(0o700)
 checks=[]
 def check(name,ok,detail=None):
  if not ok:raise AssertionError(name+': '+str(detail))
  checks.append(dict(test=name,result='PASS',detail=detail));print('PASS',name,flush=True)
 with tempfile.TemporaryDirectory(prefix='sgc-stale-release-') as td:
  t=Path(td);old=t/'old-release';old.mkdir()
  with zipfile.ZipFile(a.old_repository) as z:
   for n in ['manifest.json','manifest.sig','launcher-0.4.2-linux-amd64.zip']:(old/n).write_bytes(z.read('release/0.4.2/'+n))
  key=rsa.generate_private_key(public_exponent=65537,key_size=2048)
  name=x509.Name([x509.NameAttribute(NameOID.COMMON_NAME,'SGC local regression CA')]);now=datetime.datetime.now(datetime.timezone.utc)
  cert=(x509.CertificateBuilder().subject_name(name).issuer_name(name).public_key(key.public_key()).serial_number(x509.random_serial_number()).not_valid_before(now-datetime.timedelta(minutes=1)).not_valid_after(now+datetime.timedelta(days=1)).add_extension(x509.BasicConstraints(ca=True,path_length=None),True).add_extension(x509.SubjectAlternativeName([x509.DNSName('github.com'),x509.DNSName('release-assets.githubusercontent.com')]),False).sign(key,hashes.SHA256()))
  cp=t/'ca.pem';kp=t/'tls.key';cp.write_bytes(cert.public_bytes(serialization.Encoding.PEM));kp.write_bytes(key.private_bytes(serialization.Encoding.PEM,serialization.PrivateFormat.PKCS8,serialization.NoEncryption()))
  tls=ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER);tls.load_cert_chain(cp,kp)
  behavior={'pinned_available':True};requests=[]
  class Handler(socketserver.StreamRequestHandler):
   def handle(self):
    try:
     first=self.rfile.readline().decode().strip()
     while self.rfile.readline().strip():pass
     if not first.startswith('CONNECT '):return
     self.wfile.write(b'HTTP/1.1 200 Connection established\r\n\r\n');self.wfile.flush()
     with tls.wrap_socket(self.connection,server_side=True) as sock:
      f=sock.makefile('rb');method,path,_=f.readline().decode().strip().split(' ',2)
      while f.readline().strip():pass
      requests.append(path);name=path.rsplit('/',1)[-1]
      directory=old if '/latest/' in path or '/v0.4.2/' in path else release
      if directory==release and not behavior['pinned_available']:body=None
      elif name in ['manifest.json','manifest.sig','launcher-0.4.2-linux-amd64.zip','launcher-'+version+'-linux-amd64.zip'] and (directory/name).is_file():body=(directory/name).read_bytes()
      else:body=None
      if body is None:sock.sendall(b'HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n');return
      sock.sendall(('HTTP/1.1 200 OK\r\nContent-Length: '+str(len(body))+'\r\nConnection: close\r\n\r\n').encode());sock.sendall(body)
    except (OSError,ValueError):pass
  class Server(socketserver.ThreadingMixIn,socketserver.TCPServer):allow_reuse_address=True;daemon_threads=True
  server=Server(('127.0.0.1',0),Handler);thread=threading.Thread(target=server.serve_forever,daemon=True);thread.start()
  env=dict(os.environ,HTTPS_PROXY='http://127.0.0.1:'+str(server.server_address[1]),HTTP_PROXY='',ALL_PROXY='',NO_PROXY='',SSL_CERT_FILE=str(cp),SSL_CERT_DIR=str(t/'empty'))
  def boot(exe,root,*args,ok=True):
   r=subprocess.run([str(exe),'--install-root',str(root),'--no-launch',*args],env=env,capture_output=True,text=True,timeout=90)
   if ok and r.returncode:raise RuntimeError(r.stdout+r.stderr)
   return r
  def selected(root):
   ptr=json.loads((root/'components/launcher/current.json').read_text());return ptr,json.loads((root/'components/launcher/versions'/ptr['current']/'receipt.json').read_text())
  try:
   root=t/'existing';r=boot(a.old_setup,root);ptr,rec=selected(root)
   check('reproduced unmodified 0.4.5 setup selecting signed 0.4.2',rec['package']['version']=='0.4.2',{'old_console':r.stdout.splitlines()})
   marker=root/'userdata/save-preservation-marker';marker.parent.mkdir(exist_ok=True);marker.write_text('keep me')
   oldptr=ptr['current'];requests.clear();r=boot(setup,root);ptr,rec=selected(root)
   check('new setup skips stale latest and installs '+version,rec['package']['version']==version)
   check('version-specific endpoint was actually requested',any('/v'+version+'/manifest.json' in x for x in requests))
   check('console states actual setup and launcher version',('Setup '+version) in r.stdout and ('launcher '+version) in r.stdout)
   check('application upgrade preserves user data',marker.read_text()=='keep me')
   check('old verified launcher retained for deliberate rollback',(root/'components/launcher/versions'/oldptr).is_dir())
   active=(root/'components/launcher/current.json').read_bytes();boot(setup,root)
   check('repeat installation keeps verified current generation',(root/'components/launcher/current.json').read_bytes()==active)
   # A stale feed without the pinned release must not be treated as success.
   behavior['pinned_available']=False;empty=t/'stale-only';r=boot(setup,empty,ok=False)
   check('stale-only publication fails clearly instead of starting 0.4.2',r.returncode!=0 and 'stale signed launcher' in r.stderr and not(empty/'components/launcher/current.json').exists())
   r=boot(setup,root,ok=False)
   check('failed online update keeps current installation and data',(root/'components/launcher/current.json').read_bytes()==active and marker.read_text()=='keep me' and r.returncode!=0)
   boot(setup,root,'--command','verify')
   check('installed verified software remains usable without live feed',True)
   # Explicit offline import cannot evade the local minimum either.
   r=boot(setup,t/'old-offline','--offline',str(old),ok=False)
   check('new setup refuses an older signed offline package',r.returncode!=0 and 'stale signed launcher' in r.stderr)
   behavior['pinned_available']=True;legacy=t/'legacy-other';boot(a.old_setup,legacy)
   r=boot(setup,legacy,'--command','launch-installed',ok=False)
   check('new helper refuses direct launch-installed of old generation',r.returncode!=0 and 'older than required' in r.stderr)
   # Start and read the actual installed local UI after a true old->new update.
   log=(t/'launched.log').open('w');proc=subprocess.Popen([str(setup),'--install-root',str(root),'--command','launch-installed','--no-browser'],stdout=log,stderr=log,env=env)
   try:
    loc=root/'userdata/open-launcher.txt'
    for _ in range(500):
     if loc.exists():break
     if proc.poll() is not None:raise RuntimeError((t/'launched.log').read_text())
     time.sleep(.05)
    url,token=loc.read_text().strip().split('/#');client=urllib.request.build_opener(urllib.request.ProxyHandler({}))
    req=urllib.request.Request(url+'/api/state',headers={'Authorization':'Bearer '+token})
    with client.open(req,timeout=10) as response:state=json.load(response)
    check('actual post-upgrade launcher API reports '+version,state['version']==version and state['application']['version']==version)
    check('actual app/data identity is unambiguous',state['application']['root']==str(root) and state['application']['data']==str(root/'userdata'))
    req=urllib.request.Request(url+'/api/action',json.dumps({'action':'quit'}).encode(),headers={'Authorization':'Bearer '+token,'Content-Type':'application/json'})
    with client.open(req,timeout=10) as response:response.read()
    proc.wait(timeout=20);check('upgraded launcher and helper exit cleanly',proc.returncode==0 and not(root/'installation.lock').exists())
   finally:
    if proc.poll() is None:proc.terminate();proc.wait(timeout=15)
    log.close()
  finally:server.shutdown();server.server_close()
 a.output.parent.mkdir(parents=True,exist_ok=True);a.output.write_text(json.dumps(dict(version=version,count=len(checks),checks=checks,real_signed_release_bytes=True,transport='local HTTPS proxy fixture, not live GitHub',actual_linux_processes=True,windows_executed=False,game_executed=False),indent=2)+'\n')
if __name__=='__main__':main()
