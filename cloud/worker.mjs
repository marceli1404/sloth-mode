// Sloth Mode Account Worker. Protect account.sloth.marceli.dev with Cloudflare Access.
// Other origins and anonymous callers cannot read/write account resources.
// The local note store, model prompts, and inference are never transferred.
// Only tasks explicitly saved in this signed-in account UI are stored in D1.
const APP_ORIGIN = 'https://account.sloth.marceli.dev';
const TTL_SECONDS = 600;
const safeHeaders = {'Cache-Control':'no-store','X-Content-Type-Options':'nosniff','Referrer-Policy':'no-referrer'};
function json(data, status=200){return new Response(JSON.stringify(data),{status,headers:{...safeHeaders,'Content-Type':'application/json; charset=utf-8'}})}
function randBytes(n){let data=new Uint8Array(n);crypto.getRandomValues(data);return data}
function base32(b){let chars='ABCDEFGHIJKLMNOPQRSTUVWXYZ234567',bits=0,buf=0,out='';for(let byte of b){buf=(buf<<8)|byte;bits+=8;while(bits>=5){bits-=5;out+=chars[(buf>>>bits)&31]}}if(bits>0)out+=chars[(buf<<(5-bits))&31];return out}
async function hash(s){const b=await crypto.subtle.digest('SHA-256',new TextEncoder().encode(s));return [...new Uint8Array(b)].map(x=>x.toString(16).padStart(2,'0')).join('')}
function authenticatedAccount(identity){if(!identity||typeof identity.email!=='string'||!identity.email.includes('@'))return null;return identity.email.trim().toLowerCase()}
async function parseBody(request){if(Number(request.headers.get('Content-Length')||0)>4096)throw new Error('Request too large');const b=await request.text();if(b.length>4096)throw new Error('Request too large');return JSON.parse(b)}
function requireWriteOrigin(req){return req.headers.get('Origin')===APP_ORIGIN && req.headers.get('Content-Type')?.startsWith('application/json')}
export const service = {
 async fetch(request,env,ctx){
  const path=new URL(request.url).pathname;
  if(!ctx.access)return json({error:'Cloudflare Access required'},401);
  let identity;try{identity=await ctx.access.getIdentity()}catch{return json({error:'Could not verify identity'},401)}
  const email=authenticatedAccount(identity);if(!email)return json({error:'Verified account required'},401);
  const account_id=await hash(email);
  if(request.method==='GET'&&path==='/')return new Response(PAGE,{headers:{...safeHeaders,'Content-Type':'text/html; charset=utf-8','Content-Security-Policy':"default-src 'none'; script-src 'self'; style-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"}});
  if(request.method==='GET'&&path==='/app.js')return new Response(APP_JS,{headers:{...safeHeaders,'Content-Type':'application/javascript; charset=utf-8'}});
  if(request.method==='GET'&&path==='/app.css')return new Response(APP_CSS,{headers:{...safeHeaders,'Content-Type':'text/css; charset=utf-8'}});
  if(request.method==='GET'&&path==='/v1/me')return json({email,provider:identity.idp?.type||identity.idp||'Cloudflare Access',remote_sync:false,remote_inference:false,account_tasks:'opt_in'});
  if(!env.DB)return json({error:'Device registry unavailable'},503);
  // Cloud-tracked tasks are separate from local notes. There is no automatic
  // import, read of localhost, model inference, or background synchronization.
  if(path==='/v1/tasks'){
   if(request.method==='GET'){
    const result=await env.DB.prepare('SELECT task_id AS id,task_text AS text,done,created_at,updated_at FROM account_tasks WHERE account_id=? ORDER BY created_at DESC,task_id DESC LIMIT 200').bind(account_id).all();
    return json({tasks:(result.results||[]).map(t=>({...t,done:!!t.done}))});
   }
   if(request.method!=='POST')return json({error:'Unsupported method'},405);
   if(!requireWriteOrigin(request))return json({error:'Same-origin JSON request required'},403);
   let input;try{input=await parseBody(request)}catch{return json({error:'Invalid JSON or request too large'},400)}
   if(!input||typeof input!=='object'||Array.isArray(input))return json({error:'Invalid task request'},400);
   const now=Math.floor(Date.now()/1000);
   if(input.action==='add'){
    const value=typeof input.text==='string'?input.text.trim():'';
    if(value.length<1||value.length>280||/[\\x00-\\x08\\x0B\\x0C\\x0E-\\x1F]/.test(value))
     return json({error:'Task must contain 1–280 printable characters'},400);
    const id=crypto.randomUUID();
    const result=await env.DB.prepare('INSERT INTO account_tasks(account_id,task_id,task_text,done,created_at,updated_at) SELECT ?,?,?,0,?,? WHERE (SELECT COUNT(*) FROM account_tasks WHERE account_id=?)<200').bind(account_id,id,value,now,now,account_id).run();
    if((result.meta?.changes||0)!==1)return json({error:'Account task limit reached (200)'},409);
    return json({ok:true,id},201);
   }
   if(!/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/.test(input.id||''))
    return json({error:'Invalid task ID'},400);
   if(input.action==='toggle'){
    if(typeof input.done!=='boolean')return json({error:'Done must be true or false'},400);
    const result=await env.DB.prepare('UPDATE account_tasks SET done=?,updated_at=? WHERE account_id=? AND task_id=?').bind(input.done?1:0,now,account_id,input.id).run();
    if((result.meta?.changes||0)!==1)return json({error:'Task not found'},404);
    return json({ok:true});
   }
   if(input.action==='delete'){
    const result=await env.DB.prepare('DELETE FROM account_tasks WHERE account_id=? AND task_id=?').bind(account_id,input.id).run();
    if((result.meta?.changes||0)!==1)return json({error:'Task not found'},404);
    return json({ok:true});
   }
   return json({error:'Unsupported task action'},400);
  }
  if(request.method==='GET'&&path==='/v1/devices'){
   const x=await env.DB.prepare('SELECT device_id,label,created_at,last_seen_at,revoked_at FROM device_links WHERE account_id=? ORDER BY created_at DESC LIMIT 50').bind(account_id).all();return json({devices:x.results||[]});
  }
  if(request.method==='POST'&&path==='/v1/link-codes'){
   if(!requireWriteOrigin(request))return json({error:'Same-origin JSON request required'},403);
   const now=Math.floor(Date.now()/1000);
   // 100-bit one-time code. Rate of guesses is infeasible even without IP profiling.
   const code=base32(randBytes(13));const digest=await hash(code);
   await env.DB.batch([
    env.DB.prepare('UPDATE enrollment_codes SET consumed_at=? WHERE account_id=? AND consumed_at IS NULL').bind(now,account_id),
    env.DB.prepare('INSERT INTO enrollment_codes (code_hash,account_id,expires_at,issued_at) VALUES (?,?,?,?)').bind(digest,account_id,now+TTL_SECONDS,now),
   ]);
   return json({code,expires_in:TTL_SECONDS},201);
  }
  if(request.method==='POST'&&path==='/v1/devices/revoke'){
   if(!requireWriteOrigin(request))return json({error:'Same-origin JSON request required'},403);
   let payload;try{payload=await parseBody(request)}catch{return json({error:'Invalid JSON'},400)}
   if(typeof payload.device_id!=='string'||!/^[-a-f0-9]{36}$/.test(payload.device_id))return json({error:'Invalid device ID'},400);
   await env.DB.prepare('UPDATE device_links SET revoked_at=? WHERE account_id=? AND device_id=? AND revoked_at IS NULL').bind(Math.floor(Date.now()/1000),account_id,payload.device_id).run();
   return json({ok:true});
  }
  return json({error:'Not found'},404);
 }
};
// Filled in by release build. Keep UI external to inline scripts so CSP can block injection.
const PAGE=__PAGE__;
const APP_CSS=__CSS__;
const APP_JS=__JS__;
export default service;
