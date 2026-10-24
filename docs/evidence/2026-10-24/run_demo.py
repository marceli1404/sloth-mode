import os,sys,socket,subprocess,time
from pathlib import Path
root=Path(__file__).resolve().parent
base=root/'demo-verified-data'/'SlothModeNative'
env=os.environ.copy();env['SLOTH_BROWSER_HARNESS']='1';env['SLOTH_DEMO_BASE']=str(base)
processes=[];logs=[]
try:
 for command,port,name in [(['node','preview.mjs'],4190,'web'),([str(root/'candidate/connector/browser-harness.exe'),'-test.run=TestBrowserDemoHarness','-test.v'],39872,'bridge')]:
  with socket.socket() as probe:
   if probe.connect_ex(('127.0.0.1',port))==0:raise RuntimeError('Port already occupied: '+str(port))
  log=(root/(name+'-verified.log')).open('w',encoding='utf-8');logs.append(log)
  process=subprocess.Popen(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,creationflags=0x08000000)
  processes.append(process)
  deadline=time.monotonic()+20
  while time.monotonic()<deadline:
   if process.poll() is not None:raise RuntimeError(name+' server exited')
   with socket.socket() as probe:
    if probe.connect_ex(('127.0.0.1',port))==0:break
   time.sleep(.15)
  else:raise RuntimeError(name+' readiness timed out')
 result=subprocess.run([sys.executable,'record_demo.py'],cwd=root,env=env)
 if result.returncode:raise RuntimeError('Recording verification failed; artifacts retained')
finally:
 (base/'stop-harness').write_text('stop',encoding='ascii')
 for process in reversed(processes):
  try:process.wait(timeout=2)
  except subprocess.TimeoutExpired:
   process.terminate()
   try:process.wait(timeout=5)
   except subprocess.TimeoutExpired:process.kill();process.wait()
 for log in logs:log.close()
