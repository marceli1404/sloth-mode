// Research-led dashboard shell: semantic navigation, guest-first tasks, explicit data boundaries.
// Existing feature nodes are moved, never cloned. This module sends no notes or prompts to a server.
export const WORKSPACE_TABS = Object.freeze([
  {id:"today",label:"Today",symbol:"◫",title:"Make room for what matters.",detail:"Capture thoughts and find a gentle next step. You can start without an account.",nodes:["quiet-lab","account-task-tracker"]},
  {id:"gemma",label:"Browser Gemma",symbol:"✳",title:"A little AI, right here.",detail:"Optional Gemma 4 beta in this browser. Download and loading require your approval.",nodes:["sm-browser-ai"]},
  {id:"companion",label:"This computer",symbol:"⌘",title:"Your computer, your AI.",detail:"Pair Windows or Linux to manage your real local inbox and native Gemma.",nodes:["connector","install"]},
  {id:"integrations",label:"Integrations",symbol:"◈",title:"Your tools, on your terms.",detail:"External connections are optional and managed through your paired desktop companion.",nodes:["integrations"]},
  {id:"tour",label:"Explore",symbol:"☼",title:"A quieter way to work.",detail:"See an illustrative walkthrough, then decide which features you want to use.",nodes:["demo","why"]}
]);
const HASH_ALIASES=Object.freeze({
  "account-task-tracker":"today","sm-tasks-heading":"today","quiet-lab":"today","ql-title":"today",
  "sm-browser-ai":"gemma","sm-browser-title":"gemma",
  "connector":"companion","install":"companion","bridge-title":"companion",
  "integrations":"integrations","sm-integrations-title":"integrations",
  "demo":"tour","why":"tour","explore":"tour"
});
export function resolveWorkspaceTab(hash="",pathname="/dashboard"){
  let raw="";
  try{raw=decodeURIComponent(String(hash).replace(/^#/,""));}catch{raw=String(hash).replace(/^#/,"");}
  const normalized=raw.startsWith("sm-panel-")?raw.slice("sm-panel-".length):raw;
  if(WORKSPACE_TABS.some(t=>t.id===normalized))return normalized;
  return HASH_ALIASES[normalized] || (pathname==="/integrations"?"integrations":"today");
}
export function capabilityLabels({account="guest",companion="offline",browserAI="optional"}={}){
  const accounts={"guest":"Account tasks · guest","signed-in":"Account tasks · signed in","loading":"Account tasks · checking","expired":"Account tasks · session expired","error":"Account tasks · unavailable"};
  const companions={"offline":"Companion · not connected","paired":"Companion · connected","pairing":"Companion · pairing"};
  const models={"optional":"Browser Gemma · optional","loading":"Browser Gemma · loading","ready":"Browser Gemma · ready","error":"Browser Gemma · needs attention","unsupported":"Browser Gemma · unsupported"};
  return {account:accounts[account]||accounts.guest,companion:companions[companion]||companions.offline,browserAI:models[browserAI]||models.optional};
}
export function initWorkspace(doc=typeof document==="undefined"?null:document,win=typeof window==="undefined"?null:window){
  if(!doc||!win||doc.documentElement.dataset.slothPage!=="dashboard")return false;
  if(doc.getElementById("sm-unified-window"))return true;
  const main=doc.querySelector(".site-shell > main");
  if(!main||!WORKSPACE_TABS.every(t=>t.nodes.every(id=>doc.getElementById(id))))return false;
  const shell=doc.createElement("div");
  shell.id="sm-unified-window";
  shell.className="sm-unified-window";
  shell.setAttribute("aria-label","Sloth Mode workspace");
  const mobileLabels={today:"Today",gemma:"Gemma",companion:"Computer",integrations:"Tools",tour:"Explore"};
  const links=WORKSPACE_TABS.map(t=>'<a class="sm-window-tab" href="#'+t.id+'" id="sm-nav-'+t.id+'" aria-label="'+t.label+'" data-sm-tab="'+t.id+'"><span class="sm-window-symbol" aria-hidden="true">'+t.symbol+'</span><span class="sm-nav-label-desktop">'+t.label+'</span><span class="sm-nav-label-mobile" aria-hidden="true">'+mobileLabels[t.id]+'</span></a>').join("");
  const panes=WORKSPACE_TABS.map(t=>'<section class="sm-window-pane" id="sm-panel-'+t.id+'" aria-labelledby="sm-pane-title-'+t.id+'" hidden><div class="sm-pane-intro"><span class="sm-pane-overline">SLOTH MODE / '+(t.id==="today"?"YOUR DAY":"YOUR WORKSPACE")+'</span><h1 id="sm-pane-title-'+t.id+'" tabindex="-1">'+t.title+'</h1><p>'+t.detail+'</p></div><div class="sm-pane-items" id="sm-pane-items-'+t.id+'"></div></section>').join("");
  shell.innerHTML='<aside class="sm-window-sidebar">'
    +'<a class="sm-window-logo" href="/" aria-label="Sloth Mode homepage"><span class="sm-window-logo-mark" aria-hidden="true">◕</span><span>sloth<span class="sm-window-logo-light">mode</span><i>.</i></span></a>'
    +'<div class="sm-window-sidebar-label">YOUR SPACE</div><nav class="sm-window-tabs" aria-label="Workspace sections">'+links+'</nav>'
    +'<div class="sm-window-sidebar-bottom"><p class="sm-window-privacy"><span aria-hidden="true">●</span> Your pace. Your choice.</p><div id="sm-window-account"></div><a class="sm-window-home" href="/">← Back to the website</a></div></aside>'
    +'<div class="sm-window-main"><header class="sm-window-toolbar"><div><span class="sm-window-crumb">YOUR WORKSPACE <span aria-hidden="true"> / </span><strong id="sm-window-location">TODAY</strong></span><span class="sm-window-hello">Slow down. Get things done.</span></div>'
    +'<div class="sm-window-indicators" aria-label="Feature status"><a class="sm-window-indicator" id="sm-window-guest" href="#today">Account tasks · checking</a><a class="sm-window-indicator" id="sm-window-local" href="#companion">Companion · not connected</a><a class="sm-window-indicator" id="sm-window-ai" href="#gemma">Browser Gemma · optional</a></div></header>'
    +'<div class="sm-window-body" id="sm-unified-main" tabindex="-1">'+panes+'</div>'
    +'<footer class="sm-window-statusbar"><span>✳ Nothing happens without your say-so.</span><span>Browser · Computer · Account · External — clearly separated</span></footer></div>';
  // Move existing DOM nodes, preserving all feature handlers, pairing controls and task state.
  for(const t of WORKSPACE_TABS){
    const dest=shell.querySelector("#sm-pane-items-"+t.id);
    for(const id of t.nodes)dest.appendChild(doc.getElementById(id));
  }
  // Keep native model hardware/quantisation settings beside companion pairing, not cloud tasks.
  const nativeModel=shell.querySelector("#account-task-tracker .sm-model-card");
  if(nativeModel)shell.querySelector("#sm-pane-items-companion").prepend(nativeModel);
  const today=shell.querySelector("#sm-panel-today .sm-pane-items");
  const quick=doc.createElement("div");
  quick.className="sm-start-here";
  quick.innerHTML='<div class="sm-start-copy"><span class="sm-overline">FIRST THINGS FIRST</span><h2>Your thoughts have a place here.</h2><p>Use the local, no-account Quiet Lab to turn a scattered list into something manageable. Nothing is uploaded by Quiet Lab.</p><a class="sm-start-action" href="#quiet-lab" id="sm-start-capture">Start with a thought <span aria-hidden="true">↗</span></a></div>'
    +'<div class="sm-start-side"><span class="sm-side-eyebrow">ALWAYS OPTIONAL</span><a href="#gemma" class="sm-mini-action"><span aria-hidden="true">✳</span><span><strong>Try Browser Gemma</strong><small>Runs on your GPU after a model download</small></span><span aria-hidden="true">↗</span></a>'
    +'<a href="#companion" class="sm-mini-action"><span aria-hidden="true">⌘</span><span><strong>Connect your computer</strong><small>Local inbox, native AI and integrations</small></span><span aria-hidden="true">↗</span></a></div>';
  today.prepend(quick);
  const capture=quick.querySelector("#sm-start-capture");
  capture.addEventListener("click",event=>{
    event.preventDefault();
    const pile=doc.getElementById("ql-pile");
    pile?.focus();
    const reduceMotion=!!win.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches;
    pile?.scrollIntoView?.({behavior:reduceMotion?"auto":"smooth",block:"center"});
  });
  const accountNav=doc.querySelector(".sloth-account-nav");
  if(accountNav)shell.querySelector("#sm-window-account").appendChild(accountNav);
  const skip=doc.querySelector(".skip-link");
  if(skip){skip.href="#sm-unified-main";skip.textContent="Skip to workspace";}
  main.prepend(shell);
  doc.documentElement.classList.add("sm-workspace-ready");
  const navLinks=Array.from(shell.querySelectorAll(".sm-window-tab"));
  const panels=WORKSPACE_TABS.map(t=>shell.querySelector("#sm-panel-"+t.id));
  let active="";
  const labels=()=>({account:shell.querySelector("#sm-window-guest"),companion:shell.querySelector("#sm-window-local"),browserAI:shell.querySelector("#sm-window-ai")});
  function show(id,focus=false){
    const selected=WORKSPACE_TABS.some(t=>t.id===id)?id:"today";
    const changed=selected!==active;
    for(const link of navLinks){
      if(link.dataset.smTab===selected)link.setAttribute("aria-current","page");
      else link.removeAttribute("aria-current");
    }
    for(const p of panels)p.hidden=p.id!=="sm-panel-"+selected;
    active=selected;
    shell.querySelector("#sm-window-location").textContent=WORKSPACE_TABS.find(t=>t.id===selected).label.toUpperCase();
    if(changed&&focus)shell.querySelector("#sm-pane-title-"+selected)?.focus({preventScroll:true});
    if(changed){const scrolling=shell.querySelector(".sm-window-main");if(scrolling)scrolling.scrollTop=0;}
  }
  function syncHash(){show(resolveWorkspaceTab(win.location.hash,win.location.pathname),false);}
  navLinks.forEach(link=>link.addEventListener("click",()=>show(link.dataset.smTab,true)));
  shell.querySelectorAll(".sm-window-indicator,.sm-mini-action").forEach(a=>a.addEventListener("click",()=>show(resolveWorkspaceTab(a.hash,win.location.pathname),true)));
  win.addEventListener("hashchange",syncHash);
  win.addEventListener("popstate",syncHash);
  show(resolveWorkspaceTab(win.location.hash,win.location.pathname));
  // Account and connector have independent trust boundaries. Do not claim they sync local notes.
  const updateAccount=()=>{
    const status=doc.getElementById("sm-account-status")?.textContent||"";
    const signed=doc.getElementById("sm-task-signed");
    const current=!!signed&&!signed.hidden?"signed-in":/unavailable|error/i.test(status)?"error":/checking/i.test(status)?"loading":"guest";
    labels().account.textContent=capabilityLabels({account:current}).account;
    labels().account.dataset.connected=String(current==="signed-in");
  };
  const updateCompanion=()=>{
    const connected=!!win.SlothConnector?.isConnected?.();
    labels().companion.textContent=capabilityLabels({companion:connected?"paired":"offline"}).companion;
    labels().companion.dataset.connected=String(connected);
  };
  const updateModel=()=>{
    const node=doc.getElementById("sm-browser-status");
    const msg=node?.textContent||"";
    const state=node?.dataset.state||"";
    const browserAI=state==="error"?/unavailable|not supported/i.test(msg)?"unsupported":"error":state==="ready"?"ready":/loading|downloading|preparing/i.test(msg)?"loading":"optional";
    labels().browserAI.textContent=capabilityLabels({browserAI}).browserAI;
    labels().browserAI.dataset.connected=String(browserAI==="ready");
  };
  updateAccount();updateCompanion();updateModel();
  win.addEventListener("sloth:connection",updateCompanion);
  win.addEventListener("focus",()=>{updateCompanion();updateAccount();updateModel();});
  if(typeof MutationObserver!=="undefined"){
    const signed=doc.getElementById("sm-task-signed");
    const taskState=doc.getElementById("sm-account-status");
    const model=doc.getElementById("sm-browser-status");
    if(signed)new MutationObserver(updateAccount).observe(signed,{attributes:true,attributeFilter:["hidden"]});
    if(taskState)new MutationObserver(updateAccount).observe(taskState,{childList:true,subtree:true});
    if(model)new MutationObserver(updateModel).observe(model,{childList:true,subtree:true,attributes:true,attributeFilter:["data-state"]});
  }
  return true;
}
if(typeof document!=="undefined"){
  if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",()=>initWorkspace(),{once:true});
  else initWorkspace();
}