import test from 'node:test';
import assert from 'node:assert/strict';
import { DatabaseSync } from 'node:sqlite';
import fs from 'node:fs';
import * as account from './account-bundle.mjs';
import { buildBundle } from './build-bundle.mjs';
import enrol from './enroll.mjs';

class TestDB {
 constructor(){this.d=new DatabaseSync(':memory:');this.d.exec(fs.readFileSync(new URL('./schema.sql',import.meta.url),'utf8'))}
 prepare(sql){let stmt=this.d.prepare(sql);return {bind:(...args)=>({all:async()=>({results:stmt.all(...args)}),first:async()=>stmt.get(...args)||null,run:async()=>({meta:{changes:stmt.run(...args).changes}})})}}
 async batch(queries){const out=[];this.d.exec('BEGIN');try{for(let q of queries)out.push(await q.run());this.d.exec('COMMIT');return out}catch(e){this.d.exec('ROLLBACK');throw e}}
}
const owner={access:{getIdentity:async()=>({email:'owner@example.com',idp:{type:'github'}})}};
const root='https://account.sloth.marceli.dev';
const deviceUrl='https://enroll.sloth.marceli.dev';
function request(path,method='GET',data,origin=root){const headers=data?{'Content-Type':'application/json','Origin':origin}:{},body=data?JSON.stringify(data):undefined;return new Request(root+path,{method,headers,body})}
function devReq(path,method='POST',data,auth){return new Request(deviceUrl+path,{method,headers:{...(data?{'Content-Type':'application/json'}:{}),...(auth?{'Authorization':'Bearer '+auth}:{})},body:data?JSON.stringify(data):undefined})}
const parse=async r=>r.json();
test('requires Access, not arbitrary identity headers',async()=>{const db=new TestDB();assert.equal((await account.service.fetch(request('/v1/me'),{DB:db},{})).status,401)});
test('account sign-in provides user info and serves a dashboard',async()=>{const db=new TestDB();const me=await parse(await account.service.fetch(request('/v1/me'),{DB:db},owner));assert.equal(me.email,'owner@example.com');assert.equal(me.remote_sync,false);const r=await account.service.fetch(request('/'),{DB:db},owner);assert.equal(r.status,200);assert.match(await r.text(),/Everywhere you work/)});
test('link token is one-use, revocable, and scoped per account',async()=>{
 const DB=new TestDB(),env={DB};
 const invalid=await account.service.fetch(request('/v1/link-codes','POST',{},'https://attacker.example'),env,owner);assert.equal(invalid.status,403);
 const code=await parse(await account.service.fetch(request('/v1/link-codes','POST',{}),env,owner));assert.match(code.code,/^[A-Z2-7]{21}$/);
 const wrong=await enrol.fetch(devReq('/v1/claim','POST',{code:'A'.repeat(21),label:'Desktop'}),env);assert.equal(wrong.status,401);
 const claimed=await enrol.fetch(devReq('/v1/claim','POST',{code:code.code,label:'My laptop'}),env);assert.equal(claimed.status,201);const linked=await parse(claimed);assert.match(linked.token,/^[\w-]{43}$/);
 assert.equal((await enrol.fetch(devReq('/v1/claim','POST',{code:code.code,label:'Other'}),env)).status,401);
 const otherAccount={access:{getIdentity:async()=>({email:'other@example.com'})}};
 assert.equal((await parse(await account.service.fetch(request('/v1/devices'),env,otherAccount))).devices.length,0);
 let list=await parse(await account.service.fetch(request('/v1/devices'),env,owner));assert.equal(list.devices.length,1);assert.equal(list.devices[0].label,'My laptop');assert.equal(list.devices[0].token_hash,undefined);
 assert.equal((await enrol.fetch(devReq('/v1/heartbeat','POST',{},linked.token),env)).status,200);
 assert.equal((await account.service.fetch(request('/v1/devices/revoke','POST',{device_id:linked.device_id}),env,owner)).status,200);
 assert.equal((await enrol.fetch(devReq('/v1/heartbeat','POST',{},linked.token),env)).status,401);
});
test('public enrolment forbids browsers and arbitrary SQL fields',async()=>{const env={DB:new TestDB()};let r=await enrol.fetch(new Request(deviceUrl+'/v1/claim',{method:'POST',headers:{Origin:'https://attacker.example','Content-Type':'application/json'},body:JSON.stringify({code:'A'.repeat(21),label:'Test'})}),env);assert.equal(r.status,403);r=await enrol.fetch(devReq('/v1/claim','POST',{code:'evil',label:'x'}),env);assert.equal(r.status,400)});
test('new enrollment code invalidates previous code',async()=>{const env={DB:new TestDB()};const x=await parse(await account.service.fetch(request('/v1/link-codes','POST',{}),env,owner));const y=await parse(await account.service.fetch(request('/v1/link-codes','POST',{}),env,owner));assert.notEqual(x.code,y.code);assert.equal((await enrol.fetch(devReq('/v1/claim','POST',{code:x.code,label:'Desk'}),env)).status,401);assert.equal((await enrol.fetch(devReq('/v1/claim','POST',{code:y.code,label:'Desk'}),env)).status,201)});

test('account tasks are explicit, private to one Access identity, and not local notes',async()=>{
 const env={DB:new TestDB()};
 const otherAccount={access:{getIdentity:async()=>({email:'other@example.com',idp:{type:'google'}})}};
 assert.equal((await account.service.fetch(request('/v1/tasks'),env,{})).status,401);
 const me=await parse(await account.service.fetch(request('/v1/me'),env,owner));
 assert.equal(me.account_tasks,'opt_in');
 const input={action:'add',text:'  Renew my library card  '};
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',input,'https://evil.example'),env,owner)).status,403);
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'add',text:' ' }),env,owner)).status,400);
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'add',text:'A'.repeat(281)}),env,owner)).status,400);
 const created=await parse(await account.service.fetch(request('/v1/tasks','POST',input),env,owner));
 assert.equal(created.ok,true);
 assert.match(created.id,/^[a-f0-9-]{36}$/);
 const tasks=await parse(await account.service.fetch(request('/v1/tasks'),env,owner));
 assert.equal(tasks.tasks.length,1);
 assert.equal(tasks.tasks[0].text,'Renew my library card');
 assert.equal(tasks.tasks[0].done,false);
 assert.equal((await parse(await account.service.fetch(request('/v1/tasks'),env,otherAccount))).tasks.length,0);
 const cannotToggle=await account.service.fetch(request('/v1/tasks','POST',{action:'toggle',id:created.id,done:true}),env,otherAccount);
 assert.equal(cannotToggle.status,404);
 const cannotDelete=await account.service.fetch(request('/v1/tasks','POST',{action:'delete',id:created.id}),env,otherAccount);
 assert.equal(cannotDelete.status,404);
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'toggle',id:created.id,done:'yes'}),env,owner)).status,400);
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'toggle',id:created.id,done:true}),env,owner)).status,200);
 assert.equal((await parse(await account.service.fetch(request('/v1/tasks'),env,owner))).tasks[0].done,true);
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'delete',id:created.id}),env,owner)).status,200);
 assert.equal((await parse(await account.service.fetch(request('/v1/tasks'),env,owner))).tasks.length,0);
});

test('account task input bounds and per-account limit fail closed',async()=>{
 const env={DB:new TestDB()};
 for(const p of [
   {action:'add',text:'\u0000bad'},
   {action:'delete',id:'../../bad'},
   {action:'toggle',id:'123',done:true},
   {action:'unsupported'},
 ]) {
   assert.equal((await account.service.fetch(request('/v1/tasks','POST',p),env,owner)).status,400);
 }
 for(let i=0;i<200;i++){
  const response=await account.service.fetch(request('/v1/tasks','POST',{action:'add',text:'Synthetic account task '+i}),env,owner);
  assert.equal(response.status,201,'insert '+i);
 }
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'add',text:'Too many'}),env,owner)).status,409);
 assert.equal((await parse(await account.service.fetch(request('/v1/tasks'),env,owner))).tasks.length,200);
 const another={access:{getIdentity:async()=>({email:'independent@example.com'})}};
 assert.equal((await account.service.fetch(request('/v1/tasks','POST',{action:'add',text:'Separate account'}),env,another)).status,201);
 assert.equal((await parse(await account.service.fetch(request('/v1/tasks'),env,another))).tasks.length,1);
});

test('account browser assets show explicit opt-in storage and no automatic sync',async()=>{
 const env={DB:new TestDB()};
 const page=await (await account.service.fetch(request('/'),env,owner)).text();
 assert.match(page,/Save to my account/);
 assert.match(page,/do(es)? not automatically import or sync/);
 const js=await (await account.service.fetch(request('/app.js'),env,owner)).text();
 assert.match(js,/trackedTaskForm/);
 assert.match(js,/textContent=task.text/);
 assert.doesNotMatch(js,/innerHTML\s*=/);
});

test('committed Cloudflare account Worker bundle matches its source files',()=>{
 assert.equal(fs.readFileSync(new URL('./account-bundle.mjs',import.meta.url),'utf8'),buildBundle());
});
