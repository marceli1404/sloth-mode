'use strict';
const $=id=>document.getElementById(id);
const formatTime=ts=>ts?new Date(ts*1000).toLocaleString():null;
async function api(path,init={}){const r=await fetch(path,{credentials:'same-origin',...init,headers:{...(init.body?{'Content-Type':'application/json'}:{}),...(init.headers||{})}});let value;try{value=await r.json()}catch{throw Error('Server returned an unexpected response')}if(!r.ok)throw Error(value.error||'Request failed');return value}
function status(message,bad=false){$('message').textContent=message;$('message').classList.toggle('is-error',bad)}
function renderDevices(data){const root=$('devices');root.replaceChildren();$('deviceCount').textContent=String(data.devices.length);if(!data.devices.length){const card=document.createElement('div');card.className='empty';const strong=document.createElement('strong');strong.textContent='No linked computers yet';const p=document.createElement('p');p.textContent='Link your first Sloth Mode Connector to see it here.';card.append(strong,p);root.append(card);return}for(const d of data.devices){const card=document.createElement('article');card.className='device';const icon=document.createElement('div');icon.className='device-icon';icon.textContent='▣';const details=document.createElement('div');details.className='device-meta';const name=document.createElement('div');name.className='device-name';name.textContent=d.label;const info=document.createElement('div');info.className='device-detail';info.textContent=d.revoked_at?`Revoked ${formatTime(d.revoked_at)}`:d.last_seen_at?`Last seen ${formatTime(d.last_seen_at)}`:'Linked, waiting for first heartbeat';details.append(name,info);card.append(icon,details);if(!d.revoked_at){const button=document.createElement('button');button.textContent='Revoke';button.setAttribute('aria-label','Revoke '+d.label);button.onclick=async()=>{if(!confirm(`Revoke “${d.label}” from this account? Its local notes will remain on that PC.`))return;button.disabled=true;try{await api('/v1/devices/revoke',{method:'POST',body:JSON.stringify({device_id:d.device_id})});status('Device revoked. It will no longer authenticate to the account service.');await loadDevices()}catch(e){status(e.message,true)}finally{button.disabled=false}};card.append(button)}root.append(card)}}
async function loadDevices(){try{renderDevices(await api('/v1/devices'))}catch(e){status(e.message,true);$('devices').textContent='Unable to load devices.'}}
let validUntil=0;
async function createCode(){const btn=$('newCode');btn.disabled=true;btn.textContent='Generating…';try{const value=await api('/v1/link-codes',{method:'POST',body:'{}'});$('linkBox').hidden=false;$('enrollCode').textContent=value.code;validUntil=Date.now()+value.expires_in*1000;$('copyCode').disabled=false;status('One-time code ready. Only enter it in your own computer’s Sloth Mode Connector.');$('copyCode').focus()}catch(e){status(e.message,true)}finally{btn.disabled=false;btn.textContent='+ Generate device code'}}
function updateExpiry(){if(!validUntil)return;const seconds=Math.max(0,Math.ceil((validUntil-Date.now())/1000));$('expiry').textContent=seconds?`Expires in ${Math.ceil(seconds/60)} minute${Math.ceil(seconds/60)===1?'':'s'}`:'Expired — generate a new code';if(!seconds)$('copyCode').disabled=true}
async function copyCode(){const value=$('enrollCode').textContent;try{await navigator.clipboard.writeText(value);status('Code copied. It expires shortly and works only once.')}catch{status('Select the code above and copy it manually.',true)}}

function taskStatus(message,bad=false){$('taskMessage').textContent=message;$('taskMessage').classList.toggle('is-error',bad)}
function renderTrackedTasks(data){
 const root=$('trackedTasks');root.replaceChildren();
 const tasks=Array.isArray(data.tasks)?data.tasks:[];
 $('trackedTaskCount').textContent=String(tasks.length)+' / 200';
 if(!tasks.length){
  const empty=document.createElement('p');empty.className='loading';
  empty.textContent='No account tasks saved. Your local notes are unaffected.';
  root.append(empty);return;
 }
 for(const task of tasks){
  const row=document.createElement('div');row.className='tracked-task';
  const check=document.createElement('input');check.type='checkbox';
  check.checked=!!task.done;check.setAttribute('aria-label','Mark account task done: '+task.text);
  const label=document.createElement('span');label.textContent=task.text;
  if(task.done)label.classList.add('tracked-task-done');
  const remove=document.createElement('button');remove.type='button';remove.className='tracked-task-delete';
  remove.textContent='Delete';remove.setAttribute('aria-label','Delete account task: '+task.text);
  check.onchange=async()=>{
   check.disabled=true;remove.disabled=true;
   try{await api('/v1/tasks',{method:'POST',body:JSON.stringify({action:'toggle',id:task.id,done:check.checked})});
    taskStatus('Account task updated. Local notes were not changed.');await loadTrackedTasks();
   }catch(e){taskStatus(e.message,true);check.checked=!check.checked}
   finally{check.disabled=false;remove.disabled=false}
  };
  remove.onclick=async()=>{
   if(!confirm('Delete this account-saved task? Your local notes will not be affected.'))return;
   check.disabled=true;remove.disabled=true;
   try{await api('/v1/tasks',{method:'POST',body:JSON.stringify({action:'delete',id:task.id})});
    taskStatus('Task deleted from your account only.');await loadTrackedTasks();
   }catch(e){taskStatus(e.message,true);check.disabled=false;remove.disabled=false}
  };
  row.append(check,label,remove);root.append(row);
 }
}
async function loadTrackedTasks(){
 try{renderTrackedTasks(await api('/v1/tasks'))}
 catch(e){$('trackedTasks').textContent='Unable to load account tasks.';taskStatus(e.message,true)}
}
async function addTrackedTask(event){
 event.preventDefault();
 const content=$('trackedTaskText').value.trim(),button=$('addTrackedTask');
 if(!content||content.length>280){taskStatus('Enter a task of 1–280 characters.',true);return}
 button.disabled=true;
 try{
  await api('/v1/tasks',{method:'POST',body:JSON.stringify({action:'add',text:content})});
  $('trackedTaskText').value='';
  taskStatus('Saved to this account. No local notes were uploaded.');
  await loadTrackedTasks();
 }catch(e){taskStatus(e.message,true)}
 finally{button.disabled=false}
}

async function setup(){try{const me=await api('/v1/me');$('email').textContent=me.email;let prov=typeof me.provider==='string'?me.provider:'Cloudflare Access';$('provider').textContent='Verified via '+prov}catch(e){$('email').textContent='Unable to read account';status(e.message,true)}await loadDevices();await loadTrackedTasks();$('trackedTaskForm').onsubmit=addTrackedTask;$('newCode').onclick=createCode;$('refresh').onclick=loadDevices;$('copyCode').onclick=copyCode;setInterval(updateExpiry,10000);setInterval(loadDevices,45000)}
setup();
