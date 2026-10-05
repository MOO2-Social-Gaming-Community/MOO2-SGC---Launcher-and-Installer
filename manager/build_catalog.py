"""Build immutable metadata from the user's exact upstream patch, without executing it."""
import zipfile,re,json,posixpath,hashlib,os
from pathlib import Path
z=zipfile.ZipFile(os.environ.get('MOO2_PATCH_ZIP','/mnt/data/MOO2-1.50.26.zip')); pref='MOO2-1.50.26/patch/'
files={n[len(pref):]:z.read(n).decode('cp1252') for n in z.namelist() if n.startswith(pref) and n.lower().endswith('.cfg')}
def clean(s):
 return re.sub(r'(?m)#.*$', '', s)
def field(s,k,default=''):
 m=re.search(r'(?m)^\s*'+re.escape(k)+r'\s*=\s*(?:"([^"]*)"|([^;]+));',clean(s))
 return (m.group(1) if m and m.group(1) is not None else m.group(2).strip() if m else default)
def deps(path,visited=None):
 visited=set() if visited is None else visited
 if path in visited:return []
 visited.add(path);s=clean(files.get(path,''));out=re.findall(r'\benable\s+([\w-]+)\s*;',s)
 for p in re.findall(r'\binclude\s+([^;\s]+)\s*;',s):
  p=p.lstrip('?').replace('\\','/');sub=posixpath.normpath(posixpath.join(posixpath.dirname(path),p))
  if sub in files:out+=deps(sub,visited)
 return sorted(set(out))
mods=[]
for path,s in files.items():
 if not path.startswith('150/mods/') or len(path.split('/'))!=4:continue
 mid=field(s,'mod_id')
 if not mid:continue
 mods.append(dict(id=mid,name=field(s,'mod_name',mid),group=field(s,'mod_class'),order=int(field(s,'mod_order','0')),description=field(s,'mod_desc'),path=path,version='Bundled with 1.50.26',dependencies=deps(path),source='MOO2 community / supplied 1.50.26 distribution'))
mods.sort(key=lambda m:(m['order'],m['name']))
(Path(__file__).parent/'assets/catalog.json').write_text(json.dumps(mods,indent=2)+'\n')
print(len(mods),'mods')
for m in mods:print(m['id'],m['group'],m['dependencies'])
