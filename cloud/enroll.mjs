// Public, deliberately limited device enrolment & heartbeat API.
// NO task reads/writes and no code execution. Registration requires 104-bit OTP.
const headers={'Cache-Control':'no-store','X-Content-Type-Options':'nosniff','Content-Type':'application/json; charset=utf-8'};
const json=(b,s=200)=>new Response(JSON.stringify(b),{status:s,headers});
const hash=async x=>[...new Uint8Array(await crypto.subtle.digest('SHA-256',new TextEncoder().encode(x)))].map(x=>x.toString(16).padStart(2,'0')).join('');
const token=()=>{let b=new Uint8Array(32);crypto.getRandomValues(b);return btoa(String.fromCharCode(...b)).replace(/=/g,'').replace(/\+/g,'-').replace(/\//g,'_')};
async function read(req){if(Number(req.headers.get('Content-Length')||0)>2048)throw Error('too large');let s=await req.text();if(s.length>2048)throw Error('too large');return JSON.parse(s)}
export default {async fetch(req,env){
 const path=new URL(req.url).pathname;
 if(req.method==='GET'&&path==='/health')return json({ok:true,service:'Sloth device link',notes_stored:false});
 if(req.method!=='POST'||(path!=='/v1/claim'&&path!=='/v1/heartbeat'))return json({error:'Not found'},404);
 if(!req.headers.get('Content-Type')?.startsWith('application/json'))return json({error:'JSON required'},415);
 if(req.headers.get('Origin'))return json({error:'Browser requests not permitted'},403);
 if(!env.DB)return json({error:'Registry unavailable'},503);
 if(path==='/v1/claim'){
  let p;try{p=await read(req)}catch{return json({error:'Invalid JSON'},400)}
  const code=String(p.code||'').replace(/[ -]/g,'').toUpperCase(), label=String(p.label||'').trim();
  if(!/^[A-Z2-7]{21}$/.test(code)||label.length<2||label.length>60||/[\x00-\x1f]/.test(label))return json({error:'Invalid code or computer name'},400);
  const digest=await hash(code),now=Math.floor(Date.now()/1000);
  const row=await env.DB.prepare('SELECT account_id FROM enrollment_codes WHERE code_hash=? AND consumed_at IS NULL AND expires_at>?').bind(digest,now).first();
  if(!row)return json({error:'Code invalid, expired or already used'},401);
  const used=await env.DB.prepare('UPDATE enrollment_codes SET consumed_at=? WHERE code_hash=? AND consumed_at IS NULL AND expires_at>?').bind(now,digest,now).run();
  if((used.meta?.changes||0)!==1)return json({error:'Code was already used'},409);
  const device_id=crypto.randomUUID(),credential=token();
  await env.DB.prepare('INSERT INTO device_links(account_id,device_id,label,token_hash,created_at) VALUES(?,?,?,?,?)').bind(row.account_id,device_id,label,await hash(credential),now).run();
  return json({device_id,token:credential,heartbeat_url:'https://enroll.sloth.marceli.dev/v1/heartbeat'},201);
 }
 const supplied=req.headers.get('Authorization')||'';
 if(!/^Bearer [\w-]{43}$/.test(supplied))return json({error:'Device token required'},401);
 const now=Math.floor(Date.now()/1000),digest=await hash(supplied.slice(7));
 const result=await env.DB.prepare('UPDATE device_links SET last_seen_at=? WHERE token_hash=? AND revoked_at IS NULL').bind(now,digest).run();
 if((result.meta?.changes||0)!==1)return json({error:'Device is no longer linked'},401);
 return json({ok:true,at:now});
}};
