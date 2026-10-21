import test from "node:test";
import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import {WORKSPACE_TABS,resolveWorkspaceTab,initWorkspace,accountLabel,companionLabel,browserLabel} from "./workspace.mjs";
const html=readFileSync(new URL("./index.html",import.meta.url),"utf8");
const css=readFileSync(new URL("./workspace.css",import.meta.url),"utf8");
const script=readFileSync(new URL("./workspace.mjs",import.meta.url),"utf8");

test("all five independent product features remain available",()=>{
 const nodes=WORKSPACE_TABS.flatMap(tab=>tab.nodes);
 assert.deepEqual(WORKSPACE_TABS.map(t=>t.id),["today","gemma","companion","integrations","tour"]);
 assert.equal(new Set(nodes).size,nodes.length);
 for(const id of nodes)assert.ok(html.includes('id="'+id+'"')||html.includes('id="'+id+'" '),id);
 assert.deepEqual(WORKSPACE_TABS[0].nodes,["quiet-lab","account-task-tracker"]);
 assert.ok(nodes.includes("sm-browser-ai"));
 assert.ok(nodes.includes("connector"));
 assert.ok(nodes.includes("demo"));
 assert.ok(nodes.includes("install"));
});
test("homepage unaffected, dashboard and integrations load application shell",()=>{
 assert.match(html,/location\.pathname==="\/integrations"\?"dashboard":"home"/);
 assert.match(html,/<link rel="stylesheet" href="\/workspace\.css">/);
 assert.match(html,/<script type="module" src="\/workspace\.mjs"><\/script>/);
 assert.match(script,/dataset\.slothPage!=="dashboard"/);
 assert.match(css,/html\.sm-workspace-ready\[data-sloth-page="dashboard"\]/);
});
test("semantic app navigation uses links and aria-current, not fake tabs",()=>{
 assert.match(script,/<nav class=\\"sm-window-tabs\\"/);
 assert.match(script,/aria-current/);
 assert.match(script,/history\.pushState/);
 assert.match(script,/addEventListener\("popstate"/);
 assert.doesNotMatch(script,/role="tablist"|role="tabpanel"|aria-selected|ArrowDown|ArrowUp/);
 assert.match(css,/\[aria-current="page"\]/);
 assert.match(css,/env\(safe-area-inset-bottom\)/);
 assert.match(css,/@media\(max-width:690px\)/);
});
test("existing deep links and malformed URL escapes resolve safely",()=>{
 for(const [hash,id] of [
  ["#quiet-lab","today"],["#ql-pile","today"],["#connector","companion"],
  ["#sm-browser-ai","gemma"],["#integrations","integrations"],
  ["#sm-panel-tour","tour"],["#explore","tour"]
 ])assert.equal(resolveWorkspaceTab(hash),id);
 assert.equal(resolveWorkspaceTab("","/integrations"),"integrations");
 assert.equal(resolveWorkspaceTab("#%zbad","/dashboard"),"today");
 assert.equal(resolveWorkspaceTab("#not-real"),"today");
 assert.equal(initWorkspace(null,null),false);
});
test("capability labels keep account, computer and browser AI separate",()=>{
 assert.equal(accountLabel(false),"Account · guest");
 assert.equal(accountLabel(true),"Account tasks · signed in");
 assert.equal(companionLabel(false),"Companion · not paired");
 assert.equal(companionLabel(true),"Companion · paired");
 assert.equal(browserLabel("ready"),"Browser Gemma · ready");
 assert.equal(browserLabel("error"),"Browser Gemma · check status");
 assert.equal(browserLabel("downloading"),"Browser Gemma · beta");
 assert.doesNotMatch(script,/ACCOUNT SYNC ON/);
 assert.match(script,/data-state/);
 assert.match(script,/MutationObserver/);
});
test("guest-first capture and native model manager preserve original DOM and handlers",()=>{
 assert.match(script,/dest\.appendChild\(doc\.getElementById\(id\)\)/);
 assert.match(script,/companion.*prepend\(modelCard\)/);
 assert.match(script,/getElementById\("ql-pile"\)/);
 assert.match(script,/sm-start-capture/);
 assert.doesNotMatch(script,/cloneNode|localStorage|sessionStorage|fetch\(/);
 assert.match(html,/<h2 id="sm-tasks-heading">Account tasks\.<\/h2>/);
 assert.match(css,/prefers-reduced-motion:reduce/);
});
