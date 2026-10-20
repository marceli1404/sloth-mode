// First-party optional task tracker. Uses the existing hashed Sloth Mode account session.
async function trackedTaskRoute(req,env) {
 const headers={'Content-Type':'application/json; charset=utf-8','Cache-Control':'no-store','X-Content-Type-Options':'nosniff'};
 const json=(value,status=200)=>new Response(JSON.stringify(value),{status,headers});
 const now=Math.floor(Date.now()/1000),method=req.method;
 const token=authCookie(req,'sloth_session');
 if(!/^[a-f0-9]{64}$/.test(token))return json({error:'Sign in to save tasks'},401);
 if(!env.AUTH_DB)return json({error:'Task service unavailable'},503);
 const user=await env.AUTH_DB.prepare('SELECT account_id FROM browser_sessions WHERE token_hash=? AND revoked_at IS NULL AND expires_at>?').bind(await authHash(token),now).first();
 if(!user?.account_id)return json({error:'Session expired. Sign in again.'},401);
 const account=user.account_id;
 if(method==='GET'){
  const rows=await env.AUTH_DB.prepare('SELECT task_id AS id,task_text AS text,done,created_at FROM account_tasks WHERE account_id=? ORDER BY created_at DESC LIMIT 200').bind(account).all();
  return json({tasks:(rows.results||[]).map(x=>({...x,done:!!x.done}))});
 }
 if(method!=='POST')return json({error:'Unsupported method'},405);
 if(req.headers.get('Origin')!=='https://sloth.marceli.dev'||!req.headers.get('Content-Type')?.startsWith('application/json'))return json({error:'Same-origin JSON requests only'},403);
 let item;try{const raw=await req.text();if(raw.length>1200)return json({error:'Request too large'},413);item=JSON.parse(raw)}catch{return json({error:'Invalid JSON'},400)}
 if(!item||typeof item!=='object')return json({error:'Invalid task request'},400);
 if(item.action==='add'){
  const text=typeof item.text==='string'?item.text.trim():'';
  if(text.length<1||text.length>280)return json({error:'Task must be 1-280 characters'},400);
  const id=crypto.randomUUID();
  const inserted=await env.AUTH_DB.prepare('INSERT INTO account_tasks (account_id,task_id,task_text,done,created_at,updated_at) SELECT ?,?,?,0,?,? WHERE (SELECT COUNT(*) FROM account_tasks WHERE account_id=?)<200').bind(account,id,text,now,now,account).run();
  if((inserted.meta?.changes||0)!==1)return json({error:'Task limit reached (200)'},409);
  return json({ok:true,id},201);
 }
 if(!/^[a-f0-9-]{36}$/.test(item.id||''))return json({error:'Invalid task ID'},400);
 if(item.action==='toggle'){
  if(typeof item.done!=='boolean')return json({error:'Done must be a boolean'},400);
  await env.AUTH_DB.prepare('UPDATE account_tasks SET done=?,updated_at=? WHERE account_id=? AND task_id=?').bind(item.done?1:0,now,account,item.id).run();
  return json({ok:true});
 }
 if(item.action==='delete'){
  await env.AUTH_DB.prepare('DELETE FROM account_tasks WHERE account_id=? AND task_id=?').bind(account,item.id).run();
  return json({ok:true});
 }
 return json({error:'Unsupported task action'},400);
}
