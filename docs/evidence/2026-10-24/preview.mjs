import http from 'node:http';
import {readFile} from 'node:fs/promises';
const root=new URL('./web/',import.meta.url);
const assets=new Map([['/workspace.css','text/css'],['/workspace.mjs','text/javascript'],['/browser-ai.mjs','text/javascript'],['/favicon.svg','image/svg+xml']]);
http.createServer(async(req,res)=>{
 const path=new URL(req.url,'http://localhost').pathname;
 if(path==='/auth/status'){res.writeHead(200,{'Content-Type':'application/json'});return res.end('{"authenticated":false}');}
 if(['/','/index.html','/dashboard','/integrations'].includes(path)){res.writeHead(200,{'Content-Type':'text/html; charset=utf-8'});return res.end(await readFile(new URL('index.html',root)));}
 if(assets.has(path)){res.writeHead(200,{'Content-Type':assets.get(path)+'; charset=utf-8'});return res.end(await readFile(new URL(path.slice(1),root)));}
 res.writeHead(404);res.end('Not found');
}).listen(4190,'127.0.0.1',()=>console.log('Designer preview http://127.0.0.1:4190'));
