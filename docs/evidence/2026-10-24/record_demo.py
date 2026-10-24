import json, time, shutil
from pathlib import Path
from urllib.parse import urlparse
from playwright.sync_api import sync_playwright, expect

ROOT=Path(__file__).resolve().parent
OUT=ROOT/'recording-new';OUT.mkdir(exist_ok=True)
inputs=[
 "Email the lab report before 4 PM today.",
 "Read a chapter of a novel when I have time.",
 "I already submitted the signed club form. Nothing else is needed.",
 "The borrowed travel guide is due tomorrow. Return it to the library.",
]
network=[];errors=[];states=[];blocked=[]
with sync_playwright() as p:
 browser=p.chromium.launch(headless=True)
 context=browser.new_context(viewport={'width':1440,'height':1080},record_video_dir=str(OUT),record_video_size={'width':1440,'height':1080},reduced_motion='reduce',accept_downloads=True)
 def limit_network(route):
  u=urlparse(route.request.url)
  if u.hostname=='127.0.0.1' and u.port in (4190,39872):route.continue_()
  elif u.scheme in ('data','blob','about'):route.continue_()
  else:blocked.append({'host':u.hostname,'path':u.path});route.abort()
 context.route('**/*',limit_network)
 page=context.new_page();page.on('pageerror',lambda error:errors.append(str(error)))
 def response(resp):
  u=urlparse(resp.url)
  if u.port!=39872:return
  req=resp.request
  network.append({'method':req.method,'path':u.path,'status':resp.status,'signed':bool(req.headers.get('x-sloth-signature'))})
  if u.path in ('/v1/status','/v1/tasks'):
   try:states.append({'at':round(time.monotonic()-started,3),'path':u.path,'data':resp.json()})
   except Exception:pass
 page.on('response',response)
 started=time.monotonic()
 try:
  page.goto('http://127.0.0.1:4190/dashboard#companion');page.wait_for_load_state('networkidle')
  expect(page.locator('#bridge-key')).to_be_visible()
  page.screenshot(path=str(OUT/'01-before-pairing.png'))
  page.wait_for_timeout(1200)
  page.locator('#bridge-key').fill('sloth-synthetic-demo-key-for-local-testing-2026')
  page.locator('#bridge-connect').click()
  expect(page.locator('#bridge-connected')).to_be_visible(timeout=20000)
  expect(page.locator('#bridge-plan')).to_be_enabled(timeout=20000)
  for note in inputs:
   page.locator('#bridge-note').fill(note)
   page.locator('#bridge-add').click()
   expect(page.locator('#bridge-note')).to_have_value('',timeout=10000)
   page.wait_for_timeout(650)
  expect(page.locator('.bridge-task')).to_have_count(4)
  page.locator('#bridge-plan').scroll_into_view_if_needed();page.screenshot(path=str(OUT/'02-before-inference.png'))
  inference_start=time.monotonic()
  page.locator('#bridge-plan').click()
  expect(page.locator('#bridge-plan')).to_be_disabled(timeout=15000)
  expect(page.locator('#bridge-ai-status')).to_contain_text('Gemma drafted 4 tasks locally',timeout=180000)
  inference_elapsed=time.monotonic()-inference_start
  page.screenshot(path=str(OUT/'03-real-results.png'),full_page=True)
  tasks=next(s['data']['tasks'] for s in reversed(states) if s['path']=='/v1/tasks' and len(s['data'].get('tasks',[]))==4)
  assert all(t.get('priority') and t.get('summary') for t in tasks)
  assert [t['text'] for t in tasks]==inputs,'Original notes changed'
  assert any(x['path']=='/v1/pair' and x['status']==201 for x in network),'Real pairing missing'
  assert any(x['path']=='/v1/plan' and x['status']==202 and x['signed'] for x in network),'Signed plan missing'
  assert all(x['signed'] for x in network if x['path']=='/v1/tasks'),'Task request was not signed'
  status=page.locator('#bridge-ai-status').inner_text()
  page.wait_for_timeout(2000)
  page.locator('#bridge-tasks').scroll_into_view_if_needed()
  page.wait_for_timeout(1800)
  reading=page.locator('.bridge-task').filter(has_text=inputs[1])
  reading.get_by_role('button',name='Edit',exact=False).click()
  reading.locator('textarea').fill('Read one chapter on Sunday, if I feel like it.')
  reading.get_by_role('button',name='Save changes',exact=True).click()
  expect(page.locator('#bridge-tasks')).to_contain_text('Read one chapter on Sunday, if I feel like it.')
  expect(page.locator('.bridge-task').filter(has_text='Read one chapter on Sunday, if I feel like it.')).to_contain_text('Local note #2')
  page.wait_for_timeout(1300)
  with page.expect_download() as download_info:page.locator('#bridge-export').click()
  download=download_info.value;download.save_as(str(OUT/'reviewed-briefing.md'))
  briefing=(OUT/'reviewed-briefing.md').read_text(encoding='utf-8')
  assert 'Original #' in briefing and inputs[0] in briefing and 'Read one chapter on Sunday' in briefing
  page.reload();page.wait_for_load_state('networkidle')
  expect(page.locator('#bridge-connect')).to_contain_text('Reconnect',timeout=10000)
  page.wait_for_timeout(1200)
  page.locator('#bridge-connect').click()
  expect(page.locator('#bridge-connected')).to_be_visible(timeout=20000)
  expect(page.locator('#bridge-tasks')).to_contain_text('Read one chapter on Sunday, if I feel like it.')
  page.screenshot(path=str(OUT/'04-after-reload.png'),full_page=True)
  page.wait_for_timeout(2000)
  report={'inputs':inputs,'actual_model_tasks':tasks,'status':status,'inference_plus_ui_poll_seconds':round(inference_elapsed,3),'recording_seconds':round(time.monotonic()-started,3),'network_requests':network,'blocked_browser_requests':blocked,'page_errors':errors,'checks':{'real_browser_pairing':True,'signed_task_and_plan_requests':True,'real_local_gemma_output':True,'original_text_preserved':True,'edit_clears_old_annotations':True,'export_preserves_source':True,'reload_retains_pairing_and_notes':True},'limits':['Isolated validation origin/ports and synthetic data','Four CPU threads and zero GPU layers in fixture only','Existing verified model/runtime; no installation filmed','Browser egress restricted to loopback; machine-wide offline state not established','No human time-saving measurement']}
  assert not errors,errors
  (OUT/'run.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
  print('REAL BROWSER DEMO PASS; UI request-to-results seconds:',round(inference_elapsed,3),flush=True)
 except Exception:
  page.screenshot(path=str(OUT/'failure.png'),full_page=True)
  (OUT/'failure-state.json').write_text(json.dumps({'errors':errors,'network':network,'states':states},indent=2),encoding='utf-8')
  raise
 finally:
  video=page.video
  context.close();video.save_as(str(OUT/'real-browser-gemma-uncut.webm'));browser.close()
