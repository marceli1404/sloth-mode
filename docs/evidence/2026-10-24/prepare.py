"""Prepare an isolated Windows real-inference recording. Never downloads assets."""
import argparse,hashlib,json,os,shutil,subprocess
from pathlib import Path

HERE=Path(__file__).resolve().parent
parser=argparse.ArgumentParser()
parser.add_argument('--repo',type=Path,default=HERE.parents[2])
parser.add_argument('--assets',type=Path,required=True,help='Existing verified directory containing models/ and runtime-cpu/')
parser.add_argument('--output',type=Path,required=True,help='A NEW isolated fixture directory')
args=parser.parse_args()
if os.name!='nt':raise SystemExit('This recorded fixture is Windows-specific; no Linux hardware claim is made.')
repo,assets,out=args.repo.resolve(),args.assets.resolve(),args.output.resolve()
if out.exists():raise SystemExit('Output exists; preserve earlier evidence and choose a new directory.')
if out==assets or assets in out.parents:raise SystemExit('Keep fixture output separate from installed assets.')
if not shutil.which('go') or not shutil.which('node'):raise SystemExit('Go and Node must already be on PATH; this script installs nothing.')
model=assets/'models/google_gemma-3-4b-it-Q4_K_M.gguf'
with model.open('rb') as stream:model_hash=hashlib.file_digest(stream,'sha256').hexdigest()
if model_hash!='4996030242583a40aa151ff93f49ed787ac8c25e4120c3ae4588b2e2a7d1ae94':raise SystemExit('Existing Q4_K_M model hash differs.')
if not (assets/'runtime-cpu/llama-server.exe').is_file():raise SystemExit('Existing CPU llama.cpp runtime required.')
source=out/'candidate/connector';source.mkdir(parents=True)
for file in (repo/'connector').iterdir():
 if file.is_file() and file.suffix in ('.go','.mod','.svg','.ico'):shutil.copy2(file,source/file.name)
path=source/'model.go';text=path.read_text(encoding='utf-8')
needle='"--gpu-layers", "0"}'
if needle not in text:raise SystemExit('Runtime arguments changed; review substitutions before proceeding.')
text=text.replace(needle,'"--gpu-layers", "0", "--threads", "4", "--threads-batch", "4"}',1)
text=text.replace('args[len(args)-1] = "99"','args[len(args)-5] = "99"')
path.write_text(text,encoding='utf-8')
path=source/'main.go';text=path.read_text(encoding='utf-8')
if 'const bridgePort = "39871"' not in text or 'const bridgeOrigin = "https://sloth.marceli.dev"' not in text:raise SystemExit('Origin constants changed; review the fixture before proceeding.')
text=text.replace('const bridgePort = "39871"','const bridgePort = "39872"').replace('const bridgeOrigin = "https://sloth.marceli.dev"','const bridgeOrigin = "http://127.0.0.1:4190"')
path.write_text(text,encoding='utf-8')
shutil.copy2(HERE/'browser_harness_test.go.txt',source/'browser_harness_test.go')
shutil.copytree(HERE/'web',out/'web')
for name in ['record_demo.py','run_demo.py','preview.mjs']:shutil.copy2(HERE/name,out/name)
base=out/'demo-verified-data/SlothModeNative';base.mkdir(parents=True)
(base/'model-preference.json').write_text('{"quant":"Q4_K_M"}',encoding='ascii')
for name in ['models','runtime-cpu']:
 env=os.environ.copy();env['SLOTH_EVIDENCE_LINK']=str(base/name);env['SLOTH_EVIDENCE_TARGET']=str(assets/name)
 subprocess.run(['powershell.exe','-NoProfile','-NonInteractive','-Command','New-Item -ItemType Junction -Path $env:SLOTH_EVIDENCE_LINK -Target $env:SLOTH_EVIDENCE_TARGET | Out-Null'],env=env,check=True,creationflags=0x08000000)
subprocess.run(['go','test','-c','-o','browser-harness.exe'],cwd=source,check=True,creationflags=0x08000000)
print('Prepared',out)
print('Run with Playwright already installed: python',out/'run_demo.py')
