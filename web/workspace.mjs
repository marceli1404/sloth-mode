// Unified dashboard layout. Keep existing DOM nodes and their event handlers intact.
// No note, account, connector, or model data is copied to the server by this module.
export const WORKSPACE_TABS=Object.freeze([
  {id:"today",label:"Today",symbol:"◫",title:"Your quiet desk.",detail:"Capture thoughts, sort a small list, and explicitly save account tasks when signed in.",nodes:["account-task-tracker","quiet-lab"]},
  {id:"gemma",label:"Browser Gemma",symbol:"✳",title:"Gemma, in this browser.",detail:"A local WebGPU assistant. Model downloads only when you ask.",nodes:["sm-browser-ai"]},
  {id:"companion",label:"Desktop companion",symbol:"⌘",title:"Your computer, your AI.",detail:"Pair the optional Windows or Linux connector for private notes, local Gemma and model settings.",nodes:["connector","install"]},
  {id:"integrations",label:"Integrations",symbol:"◈",title:"Connect on purpose.",detail:"Optional tools configured through your local companion. Keys stay on your computer.",nodes:["integrations"]},
  {id:"tour",label:"Explore",symbol:"☼",title:"See how it works.",detail:"Try illustrative sorting, learn the privacy principles and review suggestions before acting.",nodes:["demo","why"]}
]);
const HASH_ALIASES=Object.freeze({
  "account-task-tracker":"today","sm-tasks-heading":"today","quiet-lab":"today","ql-title":"today",
  "sm-browser-ai":"gemma","sm-browser-title":"gemma",
  "connector":"companion","install":"companion","bridge-title":"companion",
  "integrations":"integrations","sm-integrations-title":"integrations",
  "demo":"tour","why":"tour"
});
export function resolveWorkspaceTab(hash="",pathname="/dashboard"){
  const raw=decodeURIComponent(String(hash).replace(/^#/,""));
  const normalized=raw.startsWith("sm-panel-")?raw.slice("sm-panel-".length):raw;
  if(WORKSPACE_TABS.some(t=>t.id===normalized))return normalized;
  return HASH_ALIASES[normalized] || (pathname==="/integrations"?"integrations":"today");
}
export function initWorkspace(doc=typeof document==="undefined"?null:document,win=typeof window==="undefined"?null:window){
  if(!doc || !win || doc.documentElement.dataset.slothPage!=="dashboard")return false;
  if(doc.getElementById("sm-unified-window"))return true;
  const main=doc.querySelector(".site-shell > main");
  if(!main || !WORKSPACE_TABS.every(t=>t.nodes.every(id=>doc.getElementById(id))))return false;
  const shell=doc.createElement("div");
  shell.id="sm-unified-window";
  shell.className="sm-unified-window";
  shell.setAttribute("aria-label","Sloth Mode integrated workspace");
  const tabs=WORKSPACE_TABS.map(t=>'<button class="sm-window-tab" type="button" role="tab" id="sm-tab-'+t.id+'" aria-controls="sm-panel-'+t.id+'" aria-selected="false" tabindex="-1" data-sm-tab="'+t.id+'"><span class="sm-window-symbol" aria-hidden="true">'+t.symbol+'</span><span>'+t.label+'</span></button>').join("");
  const panes=WORKSPACE_TABS.map(t=>'<section class="sm-window-pane" id="sm-panel-'+t.id+'" role="tabpanel" aria-labelledby="sm-tab-'+t.id+'" tabindex="0" hidden><div class="sm-pane-intro"><span class="sm-pane-overline">SLOTH MODE / YOUR SPACE</span><h1>'+t.title+'</h1><p>'+t.detail+'</p></div><div class="sm-pane-items" id="sm-pane-items-'+t.id+'"></div></section>').join("");
  shell.innerHTML='<aside class="sm-window-sidebar" aria-label="Workspace navigation">'
    +'<a class="sm-window-logo" href="/" aria-label="Sloth Mode homepage"><span class="sm-window-logo-mark" aria-hidden="true">◕</span><span>sloth<span class="sm-window-logo-light">mode</span><i>.</i></span></a>'
    +'<div class="sm-window-sidebar-label">YOUR WORKSPACE</div><div class="sm-window-tabs" role="tablist" aria-label="Sloth Mode sections" aria-orientation="vertical">'+tabs+'</div>'
    +'<div class="sm-window-sidebar-bottom"><p class="sm-window-privacy"><span aria-hidden="true">●</span> Local-first by default</p><div id="sm-window-account"></div><a class="sm-window-home" href="/">← Home</a></div></aside>'
    +'<div class="sm-window-main"><header class="sm-window-toolbar"><div><div class="sm-window-crumb">YOUR PRIVATE WORKSPACE <span aria-hidden="true"> / </span><strong id="sm-window-location">TODAY</strong></div><span class="sm-window-hello">One place to make things easier.</span></div><div class="sm-window-indicators"><span class="sm-window-indicator" id="sm-window-guest">GUEST-FRIENDLY</span><span class="sm-window-indicator" id="sm-window-local">COMPANION OFFLINE</span></div></header>'
    +'<div class="sm-window-body" id="sm-unified-main">'+panes+'</div><footer class="sm-window-statusbar"><span>✳ All features, one window</span><span>Sloth Mode beta · Actions stay yours to approve</span></footer></div>';
  // Move the actual nodes; never clone forms/inputs or rebind their event handlers.
  for(const t of WORKSPACE_TABS){
    const dest=shell.querySelector("#sm-pane-items-"+t.id);
    for(const id of t.nodes)dest.appendChild(doc.getElementById(id));
  }
  const accountNav=doc.querySelector(".sloth-account-nav");
  if(accountNav)shell.querySelector("#sm-window-account").appendChild(accountNav);
  main.prepend(shell);
  doc.documentElement.classList.add("sm-workspace-ready");
  const buttons=Array.from(shell.querySelectorAll('[role="tab"]'));
  const panels=WORKSPACE_TABS.map(t=>shell.querySelector("#sm-panel-"+t.id));
  let active="";
  function show(id,replaceURL=false,focus=false){
    const selected=WORKSPACE_TABS.some(t=>t.id===id)?id:"today";
    for(const b of buttons){
      const isActive=b.dataset.smTab===selected;
      b.setAttribute("aria-selected",String(isActive));
      b.tabIndex=isActive?0:-1;
    }
    for(const p of panels)p.hidden=p.id!=="sm-panel-"+selected;
    active=selected;
    shell.querySelector("#sm-window-location").textContent=WORKSPACE_TABS.find(t=>t.id===selected).label.toUpperCase();
    if(replaceURL && win.history?.replaceState)win.history.replaceState(null,"","#sm-panel-"+selected);
    if(focus)buttons.find(b=>b.dataset.smTab===selected)?.focus();
    const scrollArea=shell.querySelector(".sm-window-main");
    if(scrollArea && scrollArea.scrollTop)scrollArea.scrollTop=0;
  }
  for(let index=0;index<buttons.length;index++){
    const button=buttons[index];
    button.addEventListener("click",()=>show(button.dataset.smTab,true));
    button.addEventListener("keydown",event=>{
      if(!["ArrowDown","ArrowUp","Home","End"].includes(event.key))return;
      event.preventDefault();
      const target=event.key==="Home"?0:event.key==="End"?buttons.length-1:(index+(event.key==="ArrowDown"?1:-1)+buttons.length)%buttons.length;
      show(buttons[target].dataset.smTab,true,true);
    });
  }
  function syncHash(){const next=resolveWorkspaceTab(win.location.hash,win.location.pathname);if(next!==active)show(next);}
  win.addEventListener("hashchange",syncHash);
  // Tabs also work for old deep links to #connector, #quiet-lab and #integrations.
  show(resolveWorkspaceTab(win.location.hash,win.location.pathname));
  function updateCompanion(){
    const connected=!!win.SlothConnector?.isConnected?.();
    const label=shell.querySelector("#sm-window-local");
    label.textContent=connected?"COMPANION CONNECTED":"COMPANION OFFLINE";
    label.dataset.connected=String(connected);
  }
  updateCompanion();
  win.addEventListener("sloth:connection",updateCompanion);
  win.addEventListener("focus",updateCompanion);
  // Sign-in is optional. Observe the existing account indicator without changing its API.
  const taskStatus=doc.getElementById("sm-account-status");
  const indicator=shell.querySelector("#sm-window-guest");
  const updateAccount=()=>{
    const signed=doc.getElementById("sm-task-signed");
    const isSigned=!!signed&&!signed.hidden;
    indicator.textContent=isSigned?"ACCOUNT SYNC ON":"GUEST-FRIENDLY";
    indicator.dataset.connected=String(isSigned);
  };
  updateAccount();
  if(typeof MutationObserver!=="undefined" && taskStatus){
    const observer=new MutationObserver(updateAccount);
    observer.observe(doc.getElementById("sm-task-signed"),{attributes:true,attributeFilter:["hidden"]});
  }
  return true;
}
if(typeof document!=="undefined"){
  if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",()=>initWorkspace(),{once:true});
  else initWorkspace();
}