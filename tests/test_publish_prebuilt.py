import json, pathlib, subprocess, sys, tempfile, unittest
sys.path.insert(0,str(pathlib.Path(__file__).resolve().parents[1]/'packaging'))
from publish_prebuilt import execute_publish

class FakeGitHub:
    def __init__(self):self.exists=False;self.draft=True;self.assets={};self.calls=[];self.corrupt=False;self.latest=None;self.tag=None
    def __call__(self,args,**kwargs):
        self.calls.append(args);command=args[1:];out='';err='';rc=0
        if command[0]=='api' and command[1].endswith('/releases/latest'):
            if self.latest is None:rc=1;err='HTTP 404: Not Found'
            else:out=json.dumps({'tag_name':self.latest})
        elif command[0]=='api':
            if not self.exists:rc=1;err='HTTP 404: Not Found'
            else:out=json.dumps({'draft':self.draft,'assets':[{'name':n} for n in self.assets]})
        elif command[:2]==['release','create']:self.exists=True;self.tag=command[2]
        elif command[:2]==['release','upload']:
            i=command.index('--repo')+2
            for name in command[i:]:
                p=pathlib.Path(name);self.assets[p.name]=p.read_bytes()
        elif command[:2]==['release','download']:
            dest=pathlib.Path(command[command.index('--dir')+1]);names=self.assets
            if '--pattern' in command:names=[command[command.index('--pattern')+1]]
            for n in names:(dest/n).write_bytes(self.assets[n]+(b'x' if self.corrupt else b''))
        elif command[:2]==['release','edit']:
            self.draft=False
            if '--latest' in command:self.latest=command[2]
        else:raise AssertionError(args)
        if rc and kwargs.get('check'):raise subprocess.CalledProcessError(rc,args,out,err)
        return subprocess.CompletedProcess(args,rc,out,err)

class PublishSafeguards(unittest.TestCase):
    def setUp(self):
        self.temp=tempfile.TemporaryDirectory();self.addCleanup(self.temp.cleanup);r=pathlib.Path(self.temp.name)
        self.files=[r/'RELEASE-NOTES.md',r/'manifest.json',r/'package.zip']
        for p in self.files:p.write_bytes(p.name.encode())
        self.gh=FakeGitHub()
    def publish(self):return execute_publish('owner/repo','0.4.1',self.files,'a'*40,self.gh)
    def test_fresh_draft_verify_publish(self):
        self.assertIn('Published',self.publish());self.assertFalse(self.gh.draft)
        self.assertTrue(any('--target' in x for x in self.gh.calls));self.assertFalse(any('--clobber' in x for x in self.gh.calls))
    def test_repeat_is_noop(self):
        self.publish();self.gh.calls=[];self.assertIn('no changes',self.publish());self.assertFalse(any(x[1:3] in (['release','create'],['release','upload'],['release','edit']) for x in self.gh.calls))
    def test_published_missing_assets_refused(self):
        self.gh.exists=True;self.gh.draft=False
        with self.assertRaisesRegex(RuntimeError,'asset set differs'):self.publish()
    def test_existing_different_asset_refused(self):
        self.publish();self.gh.assets['package.zip']=b'altered'
        with self.assertRaisesRegex(RuntimeError,'refusing overwrite'):self.publish()
    def test_partial_draft_resumes(self):
        self.gh.exists=True;self.gh.assets={'manifest.json':b'manifest.json'};self.assertIn('Published',self.publish());self.assertEqual(len(self.gh.assets),3)
    def test_corrupt_download_never_published(self):
        self.gh.corrupt=True
        with self.assertRaisesRegex(RuntimeError,'Readback differs'):self.publish()
        self.assertTrue(self.gh.draft)
    def test_unexpected_remote_asset_refused(self):
        self.gh.exists=True;self.gh.assets={'private.key':b'no'}
        with self.assertRaisesRegex(RuntimeError,'unexpected'):self.publish()
    def test_stale_latest_repaired_without_upload(self):
        self.publish();self.gh.latest='v0.4.0';self.gh.calls=[]
        self.assertIn('pointer repaired',self.publish());self.assertEqual(self.gh.latest,'v0.4.1')
        self.assertFalse(any(x[1:3]==['release','upload'] for x in self.gh.calls))
    def test_newer_latest_never_downgraded(self):
        self.gh.latest='v0.4.9';self.publish();self.assertEqual(self.gh.latest,'v0.4.9')
        self.assertTrue(any('--latest=false' in x for x in self.gh.calls))
    def test_unknown_latest_refused(self):
        self.gh.latest='unrecognized'
        with self.assertRaisesRegex(RuntimeError,'Unexpected latest'):self.publish()
if __name__=='__main__':unittest.main()
