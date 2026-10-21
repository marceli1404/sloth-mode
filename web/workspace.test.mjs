import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { WORKSPACE_TABS, resolveWorkspaceTab, initWorkspace } from "./workspace.mjs";

const html=readFileSync(new URL("./index.html",import.meta.url),"utf8");
const css=readFileSync(new URL("./workspace.css",import.meta.url),"utf8");
const script=readFileSync(new URL("./workspace.mjs",import.meta.url),"utf8");

test("all features are present in exactly one integrated window",()=>{
  const nodes=WORKSPACE_TABS.flatMap(tab=>tab.nodes);
  assert.equal(WORKSPACE_TABS.length,5);
  assert.equal(new Set(nodes).size,nodes.length);
  for(const id of nodes)assert.ok(html.includes('id="'+id+'"') || html.includes('id="'+id+'" '),id);
  assert.ok(nodes.includes("sm-browser-ai"));
  assert.ok(nodes.includes("account-task-tracker"));
  assert.ok(nodes.includes("quiet-lab"));
  assert.ok(nodes.includes("connector"));
  assert.ok(nodes.includes("integrations"));
  assert.ok(nodes.includes("demo"));
  assert.ok(nodes.includes("install"));
});
test("dashboard and /integrations both boot the window",()=>{
  assert.match(html,/location\.pathname==="\/integrations"\?"dashboard":"home"/);
  assert.match(html,/<link rel="stylesheet" href="\/workspace\.css">/);
  assert.match(html,/<script type="module" src="\/workspace\.mjs"><\/script>/);
});
test("deep links and guest fallback resolve without server state",()=>{
  assert.equal(resolveWorkspaceTab("#quiet-lab"),"today");
  assert.equal(resolveWorkspaceTab("#connector"),"companion");
  assert.equal(resolveWorkspaceTab("#sm-browser-ai"),"gemma");
  assert.equal(resolveWorkspaceTab("#integrations"),"integrations");
  assert.equal(resolveWorkspaceTab("#sm-panel-tour"),"tour");
  assert.equal(resolveWorkspaceTab("","/integrations"),"integrations");
  assert.equal(resolveWorkspaceTab("#not-real","/dashboard"),"today");
  assert.equal(initWorkspace(null,null),false);
});
test("does not clone existing feature forms, persist or upload notes",()=>{
  assert.match(script,/dest\.appendChild\(doc\.getElementById\(id\)\)/);
  assert.doesNotMatch(script,/cloneNode|localStorage|sessionStorage|fetch\(/);
  assert.match(script,/window-sidebar/);
  assert.match(script,/role="tablist"/);
  assert.match(script,/aria-selected/);
  assert.match(script,/ArrowDown/);
  assert.match(css,/@media\(max-width:690px\)/);
  assert.match(css, /prefers-reduced-motion:reduce/);
});
test("the homepage remains independent of the integrated dashboard",()=>{
  assert.match(script,/dataset\.slothPage!=="dashboard"/);
  assert.match(css, /html\.sm-workspace-ready\[data-sloth-page="dashboard"\]/);
});
