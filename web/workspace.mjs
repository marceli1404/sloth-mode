// Calm application shell: semantic navigation, real capability boundaries and guest-first capture.
// Original feature nodes move into panels intact; no task, prompt, key or note is synced here.
export const WORKSPACE_TABS=Object.freeze([
 {id:"today",label:"Today",symbol:"◫",title:"Clear your head.",detail:"Start with your thoughts. Nothing to install, nothing to sign into.",boundary:"BROWSER ONLY · LOCAL RULES",nodes:["quiet-lab","account-task-tracker"]},
 {id:"gemma",label:"Browser Gemma",symbol:"✳",title:"Make room for thinking.",detail:"Run Gemma 4 on your browser's GPU, when your device supports it.",boundary:"BROWSER GPU · OPTIONAL DOWNLOAD",nodes:["sm-browser-ai"]},
 {id:"companion",label:"This computer",symbol:"⌘",title:"A quieter computer.",detail:"The optional Windows and Linux companion unlocks a private desktop inbox, native Gemma and hardware tools.",boundary:"THIS PC · EXPLICIT PAIRING",nodes:["connector","install"]},
 {id:"integrations",label:"Integrations",symbol:"◇",title:"Only what you connect.",detail:"Add tools only when you need them. The companion keeps credentials on your computer.",boundary:"THIS PC → EXTERNAL PROVIDER",nodes:["integrations"]},
 {id:"tour",label:"Explore",symbol:"☼",title:"Get to know Sloth Mode.",detail:"A quick walkthrough of how Sloth Mode helps you sort the noise.",boundary:"ILLUSTRATIVE EXAMPLE · NOT LIVE AI",nodes:["demo","why"]}
]);
const HASH_ALIASES=Object.freeze({
 "account-task-tracker":"today","sm-tasks-heading":"today","quiet-lab":"today","ql-title":"today","ql-pile":"today",
 "sm-browser-ai":"gemma","sm-browser-title":"gemma",
 "connector":"companion","install":"companion","bridge-title":"companion",
 "integrations":"integrations","sm-integrations-title":"integrations",
 "demo":"tour","why":"tour","explore":"tour"
});
const htmlEsc=value=>String(value).replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
export function resolveWorkspaceTab(hash="",pathname="/dashboard"){
 let raw="";
 try{raw=decodeURIComponent(String(hash).replace(/^#/,""));}catch{raw=String(hash).replace(/^#/,"");}
 const normalized=raw.startsWith("sm-panel-")?raw.slice("sm-panel-".length):raw;
 if(WORKSPACE_TABS.some(t=>t.id===normalized))return normalized;
 return HASH_ALIASES[normalized]||(pathname==="/integrations"?"integrations":"today");
}
export function accountLabel(isSignedIn){return isSignedIn?"Account tasks · signed in":"Account · guest";}
export function companionLabel(isPaired){return isPaired?"Companion · paired":"Companion · not paired";}
export function browserLabel(state){return state==="ready"?"Browser Gemma · ready":state==="error"?"Browser Gemma · check status":"Browser Gemma · beta";}
export function initWorkspace(doc=typeof document==="undefined"?null:document,win=typeof window==="undefined"?null:window){
 if(!doc||!win||doc.documentElement.dataset.slothPage!=="dashboard")return false;
 if(doc.getElementById("sm-unified-window"))return true;
 const main=doc.querySelector(".site-shell > main");
 if(!main||!WORKSPACE_TABS.every(t=>t.nodes.every(id=>doc.getElementById(id))))return false;
 const shell=doc.createElement("div");
 shell.id="sm-unified-window";shell.className="sm-unified-window";
 shell.setAttribute("aria-label","Sloth Mode workspace");
 const SHORT_NAV=Object.freeze({today:"Today",gemma:"Gemma",companion:"Computer",integrations:"Tools",tour:"Explore"});
 const nav=WORKSPACE_TABS.map(t=>'<a class="sm-window-tab" href="#'+t.id+'" aria-label="'+htmlEsc(t.label)+'" data-sm-route="'+t.id+'"><span class="sm-window-symbol" aria-hidden="true">'+t.symbol+'</span><span class="sm-window-nav-label">'+htmlEsc(t.label)+'</span><span class="sm-window-nav-mobile" aria-hidden="true">'+SHORT_NAV[t.id]+'</span></a>').join("");
 const panes=WORKSPACE_TABS.map(t=>'<section class="sm-window-pane" id="sm-panel-'+t.id+'" aria-labelledby="sm-heading-'+t.id+'" hidden>'
  +'<div class="sm-pane-intro"><div class="sm-pane-overline"><span class="sm-overline-dot" aria-hidden="true">✳</span> YOUR QUIET SPACE <span class="sm-overline-slash">/</span> '+htmlEsc(t.label.toUpperCase())+'</div>'
  +'<div class="sm-pane-heading"><div><h1 id="sm-heading-'+t.id+'">'+htmlEsc(t.title)+'</h1><p>'+htmlEsc(t.detail)+'</p></div>'
  +(t.id==="today"?'<button class="sm-start-capture" id="sm-start-capture" type="button">Start with a thought <span aria-hidden="true">↗</span></button>':'')
  +'</div><div class="sm-pane-boundary"><span class="sm-boundary-mark" aria-hidden="true">●</span> '+htmlEsc(t.boundary)+'</div></div>'
  +(t.id==="today"?'<div class="sm-today-guide" aria-label="Optional ways to extend your workspace"><div><span class="sm-guide-eyebrow">MAKE IT YOURS</span><strong>Keep it simple. Add power when you need it.</strong></div><div class="sm-today-guide-actions"><a href="#gemma" data-sm-route="gemma">Try browser AI <span aria-hidden="true">↗</span></a><a href="#companion" data-sm-route="companion">Connect this PC <span aria-hidden="true">↗</span></a></div></div>':'')
  +'<div class="sm-pane-items" id="sm-pane-items-'+t.id+'"></div></section>').join("");
 shell.innerHTML='<aside class="sm-window-sidebar" aria-label="Application shell">'
  +'<a class="sm-window-logo" href="/" aria-label="Sloth Mode home"><span class="sm-window-logo-mark" aria-hidden="true">✳</span><span>sloth<span class="sm-window-logo-light">mode</span><i>.</i></span></a>'
  +'<div class="sm-window-sidebar-label">YOUR WORKSPACE</div><nav class="sm-window-tabs" aria-label="Workspace sections">'+nav+'</nav>'
  +'<div class="sm-window-sidebar-bottom"><p class="sm-window-privacy"><span aria-hidden="true">●</span> Local first. Your choice.</p><div id="sm-window-account"></div><a class="sm-window-home" href="/">← Back to website</a></div></aside>'
  +'<div class="sm-window-main"><header class="sm-window-toolbar"><div><div class="sm-window-crumb">SLOTH MODE <span aria-hidden="true">/</span> <strong id="sm-window-location">TODAY</strong></div>'
  +'<span class="sm-window-hello">A calmer way through the day.</span></div><div class="sm-window-indicators" aria-label="Independent feature statuses">'
  +'<a class="sm-window-indicator" href="#today" data-sm-route="today" id="sm-window-guest">Account · guest</a>'
  +'<a class="sm-window-indicator" href="#companion" data-sm-route="companion" id="sm-window-local">Companion · not paired</a>'
  +'<a class="sm-window-indicator" href="#gemma" data-sm-route="gemma" id="sm-window-ai">Browser Gemma · beta</a>'
  +'</div></header>'
  +'<div class="sm-window-body" id="sm-unified-main" tabindex="-1">'+panes+'</div>'
  +'<footer class="sm-window-statusbar"><span>✳ Work at your own pace.</span><span>Private actions remain yours to approve.</span></footer></div>';
 // Preserve original control objects, form state and attached listeners; never clone.
 for(const t of WORKSPACE_TABS){
  const dest=shell.querySelector("#sm-pane-items-"+t.id);
  for(const id of t.nodes)dest.appendChild(doc.getElementById(id));
 }
 // Hardware/quantisation controls belong to the paired computer, not the cloud task list.
 const modelCard=doc.querySelector("#account-task-tracker .sm-model-card");
 if(modelCard)shell.querySelector("#sm-pane-items-companion").prepend(modelCard);
 const accountNav=doc.querySelector(".sloth-account-nav");
 if(accountNav)shell.querySelector("#sm-window-account").appendChild(accountNav);
 main.prepend(shell);
 const skip=doc.querySelector(".skip-link");
 if(skip){skip.href="#sm-unified-main";skip.textContent="Skip to workspace";}
 doc.documentElement.classList.add("sm-workspace-ready");
 const links=Array.from(shell.querySelectorAll("[data-sm-route]"));
 const panels=WORKSPACE_TABS.map(t=>shell.querySelector("#sm-panel-"+t.id));
 const heading=shell.querySelector("#sm-window-location");
 let active="";
 function show(id,{push=false,focus=false}={}){
  const selected=WORKSPACE_TABS.some(t=>t.id===id)?id:"today";
  for(const link of links){
   if(link.classList.contains("sm-window-tab")){
    if(link.dataset.smRoute===selected)link.setAttribute("aria-current","page");
    else link.removeAttribute("aria-current");
   }
  }
  for(const panel of panels)panel.hidden=panel.id!=="sm-panel-"+selected;
  heading.textContent=WORKSPACE_TABS.find(t=>t.id===selected).label.toUpperCase();
  active=selected;
  if(push&&win.history?.pushState&&win.location.hash!=="#"+selected)win.history.pushState(null,"","#"+selected);
  if(focus)shell.querySelector("#sm-heading-"+selected)?.focus?.();
  const scroller=shell.querySelector(".sm-window-main");
  if(scroller&&scroller.scrollTop)scroller.scrollTop=0;
 }
 for(const link of links)link.addEventListener("click",event=>{
  const to=link.dataset.smRoute;
  if(!to)return;
  event.preventDefault();
  show(to,{push:true});
 });
 const capture=shell.querySelector("#sm-start-capture");
 capture?.addEventListener("click",()=>{
  const input=doc.getElementById("ql-pile");
  input?.focus();
  input?.scrollIntoView?.({block:"center",behavior:win.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches?"auto":"smooth"});
 });
 function syncRoute(){const next=resolveWorkspaceTab(win.location.hash,win.location.pathname);if(next!==active)show(next);}
 win.addEventListener("popstate",syncRoute);
 win.addEventListener("hashchange",syncRoute);
 show(resolveWorkspaceTab(win.location.hash,win.location.pathname));
 const observer=typeof MutationObserver==="undefined"?null:MutationObserver;
 const account=shell.querySelector("#sm-window-guest");
 const signed=doc.getElementById("sm-task-signed");
 function updateAccount(){
  const isSigned=!!signed&&!signed.hidden;
  account.textContent=accountLabel(isSigned);
  account.dataset.state=isSigned?"ready":"neutral";
 }
 updateAccount();
 if(observer&&signed)new observer(updateAccount).observe(signed,{attributes:true,attributeFilter:["hidden"]});
 const companion=shell.querySelector("#sm-window-local");
 function updateCompanion(){
  let paired=false;
  try{paired=!!win.SlothConnector?.isConnected?.();}catch{}
  companion.textContent=companionLabel(paired);
  companion.dataset.state=paired?"ready":"neutral";
 }
 updateCompanion();
 win.addEventListener("sloth:connection",updateCompanion);
 win.addEventListener("focus",updateCompanion);
 const ai=shell.querySelector("#sm-window-ai");
 const status=doc.getElementById("sm-browser-status");
 function updateAI(){
  const state=status?.dataset?.state||"";
  ai.textContent=browserLabel(state);
  ai.dataset.state=state==="ready"?"ready":state==="error"?"error":"neutral";
 }
 updateAI();
 if(observer&&status)new observer(updateAI).observe(status,{attributes:true,attributeFilter:["data-state"],childList:true,characterData:true,subtree:true});
 return true;
}
if(typeof document!=="undefined"){
 if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",()=>initWorkspace(),{once:true});
 else initWorkspace();
}
