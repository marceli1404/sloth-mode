import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';

const html=readFileSync(new URL('./index.html',import.meta.url),'utf8');
const assignment=html.split('\n').find(line=>line.includes('window.SlothConnector=Object.freeze('));
function workspaceAPI({hidden=false,device={id:'synthetic'},key='',connectionActive=true}={}){
 const calls=[];
 const context={window:{},device,key,connectionActive,connected:{hidden},request:async(...args)=>{calls.push(args);return {ok:true};}};
 vm.runInNewContext(assignment,context);
 return {api:context.window.SlothConnector,calls};
}

test('paired workspace can read hardware and save model preference',async()=>{
 const {api,calls}=workspaceAPI();
 await api.request('/v1/hardware');
 await api.request('/v1/model/preference',{quant:'Q4_K_M'});
 await api.request('/v1/integrations');
 assert.equal(calls.length,3);
 assert.equal(calls[0][0],'/v1/hardware');
 assert.equal(calls[0][1],undefined);
 assert.equal(calls[1][1].quant,'Q4_K_M');
 assert.ok(calls.every(call=>call[2]===true),'all calls retain authentication');
});

test('workspace API keeps unrelated paths and unintended methods private',async()=>{
 const {api,calls}=workspaceAPI();
 for(const [path,body] of [['/v1/tasks',undefined],['/v1/setup',{}],['/v1/hardware',{}],['/v1/model/preference',undefined]]){
  await assert.rejects(api.request(path,body),/not available/);
 }
 assert.equal(calls.length,0);
});

test('saved browser credentials alone do not mean an active connection',()=>{
 assert.equal(workspaceAPI({hidden:true}).api.isConnected(),false);
 assert.equal(workspaceAPI({hidden:false}).api.isConnected(),true);
 assert.equal(workspaceAPI({hidden:false,device:null}).api.isConnected(),false);
 assert.equal(workspaceAPI({hidden:false,connectionActive:false}).api.isConnected(),false);
});
