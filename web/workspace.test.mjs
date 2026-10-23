import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { WORKSPACE_TABS, resolveWorkspaceTab, initWorkspace, capabilityLabels } from "./workspace.mjs";

const html=readFileSync(new URL("./index.html",import.meta.url),"utf8");
const css=readFileSync(new URL("./workspace.css",import.meta.url),"utf8");
const script=readFileSync(new URL("./workspace.mjs",import.meta.url),"utf8");

test("all original features appear once in integrated workspace",()=>{
 const nodes=WORKSPACE_TABS.flatMap(t=>t.nodes);
 assert.equal(WORKSPACE_TABS.length,5);
 assert.equal(new Set(nodes).size,nodes.length);
 for(const id of nodes)assert.ok(html.includes('id="'+id+'"'),id);
 assert.equal(WORKSPACE_TABS[0].nodes[0],"quiet-lab");
 assert.ok(nodes.includes("sm-browser-ai"));
 assert.ok(nodes.includes("connector"));
 assert.ok(nodes.includes("integrations"));
 assert.ok(nodes.includes("demo"));
});
test("both dashboard routes load module and preserve homepage",()=>{
 assert.match(html,/location\.pathname==="\/integrations"\?"dashboard":"home"/);
 assert.match(html,/<link rel="stylesheet" href="\/workspace\.css\?edition=botanical-20261023">/);
 assert.match(html,/<script type="module" src="\/workspace\.mjs"><\/script>/);
 assert.match(script,/dataset\.slothPage!=="dashboard"/);
 assert.match(css,/html\.sm-workspace-ready\[data-sloth-page="dashboard"\]/);
});
test("legacy hashes, malformed hashes and guest defaults resolve safely",()=>{
 assert.equal(resolveWorkspaceTab("#quiet-lab"),"today");
 assert.equal(resolveWorkspaceTab("#connector"),"companion");
 assert.equal(resolveWorkspaceTab("#sm-browser-ai"),"gemma");
 assert.equal(resolveWorkspaceTab("#integrations"),"integrations");
 assert.equal(resolveWorkspaceTab("#sm-panel-tour"),"tour");
 assert.equal(resolveWorkspaceTab("","/integrations"),"integrations");
 assert.equal(resolveWorkspaceTab("#not-real","/dashboard"),"today");
 assert.equal(resolveWorkspaceTab("#%E0%A4%A","/dashboard"),"today");
 assert.equal(initWorkspace(null,null),false);
});
test("primary navigation is semantic links, not misoriented ARIA tabs",()=>{
 assert.match(script,/<nav class="sm-window-tabs" aria-label="Workspace sections">/);
 assert.match(script,/href="#'\+t\.id\+'/);
 assert.match(script,/aria-current","page"/);
 assert.doesNotMatch(script,/role="tablist"|role="tabpanel"|aria-orientation="vertical"|role="tab"/);
 assert.doesNotMatch(script,/ArrowDown|ArrowUp|aria-selected/);
 assert.match(script,/win\.addEventListener\("hashchange",syncHash\)/);
 assert.match(script,/win\.addEventListener\("popstate",syncHash\)/);
 assert.match(script,/skip\.href="#sm-unified-main"/);
});
test("statuses separate cloud tasks, connector and WebGPU without implying local sync",()=>{
 const labels=capabilityLabels({account:"signed-in",companion:"offline",browserAI:"optional"});
 assert.equal(labels.account,"Account tasks · signed in");
 assert.equal(labels.companion,"Companion · not connected");
 assert.equal(labels.browserAI,"Browser Gemma · optional");
 assert.match(capabilityLabels({browserAI:"error"}).browserAI,/needs attention/);
 assert.doesNotMatch(script,/ACCOUNT SYNC ON|YOUR PRIVATE WORKSPACE · SYNCED/);
});
test("guest-first capture and explicit action boundaries",()=>{
 assert.match(script,/id="sm-start-capture"/);
 assert.match(script,/getElementById\("ql-pile"\)/);
 assert.match(script,/dest\.appendChild\(doc\.getElementById\(id\)\)/);
 assert.doesNotMatch(script,/cloneNode|localStorage|sessionStorage|fetch\(/);
 assert.match(script,/Nothing is uploaded by Quiet Lab/);
 assert.match(script,/Runs on your GPU after a model download/);
 assert.match(script,/win\.SlothConnector\?\.isConnected\?\.\(\)/);
 assert.match(html,/<h2 id="sm-tasks-heading">/);
});
test("responsive layout, accessible controls and reduced motion",()=>{
 assert.match(css,/@media\(max-width:690px\)/);
 assert.match(css,/env\(safe-area-inset-bottom\)/);
 assert.match(css,/\.sm-window-tab\[aria-current="page"\]/);
 assert.match(css,/.sm-window-tab:focus-visible/);
 assert.match(css,/prefers-reduced-motion:reduce/);
 assert.match(css,/\.sm-start-here\{display:grid/);
});

test("mobile navigation remains visible and fully labelled at narrow widths",()=>{
 assert.match(script,/aria-label="'\+t\.label\+'/);
 assert.match(script,/sm-nav-label-mobile/);
 assert.match(css,/\.sm-window-tab\{flex:1 1 0!important/);
 assert.match(css,/\.sm-nav-label-desktop\{display:none\}/);
 assert.match(css,/@media\(max-width:360px\)/);
});
test("model hardware controls live with computer while DOM handlers remain intact",()=>{
 assert.match(script,/shell\.querySelector\("#account-task-tracker \.sm-model-card"\)/);
 assert.match(script,/querySelector\("#sm-pane-items-companion"\)\.prepend\(nativeModel\)/);
 assert.match(script,/dest\.appendChild\(doc\.getElementById\(id\)\)/);
 assert.match(css,/#sm-pane-items-companion > \.sm-model-card/);
});
test("capture shortcut respects user's reduced-motion preference",()=>{
 assert.match(script,/prefers-reduced-motion: reduce/);
 assert.match(script,/reduceMotion\?"auto":"smooth"/);
});

test("detached-node lookup never searches document before shell mounts",()=>{
 const move=script.indexOf('dest.appendChild(doc.getElementById(id))');
 const select=script.indexOf('shell.querySelector("#account-task-tracker .sm-model-card")');
 const mount=script.indexOf('main.prepend(shell)');
 assert.ok(move>=0&&select>move&&mount>select,"Look up moved card in the detached shell before mount");
 assert.doesNotMatch(script,/doc\.querySelector\("#account-task-tracker \.sm-model-card"\)/);
});
