import test from "node:test";
import assert from "node:assert/strict";
import {readFileSync} from "node:fs";

const file=(name)=>readFileSync(new URL(name,import.meta.url));
const svg=file("./favicon.svg").toString("utf8").trim();
const logo=String.raw`<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 36 36">`;
const homepage=file("./index.html").toString("utf8");
const workspace=file("./workspace.mjs").toString("utf8");
const css=file("./workspace.css").toString("utf8");
const account=readFileSync(new URL("../cloud/account.html",import.meta.url),"utf8");
const icon=file("../connector/sloth.ico");

test("vector assets are identical to canonical homepage mark",()=>{
 assert.ok(svg.startsWith(logo));
 for(const p of ["../connector/icon.svg","../linux/sloth-mode.svg"]){
  assert.equal(file(p).toString("utf8").trim(),svg,p);
 }
 const outline='M8 24C4 17 9 9 17 8c7-1 13 4 12 12-1 7-8 12-15 9';
 const smile='M14 20c3 3 7 4 10 2';
 for(const source of [svg,homepage,workspace,account]){
  assert.ok(source.includes(outline),"missing homepage sloth outline");
  assert.ok(source.includes(smile),"missing sloth smile");
 }
 assert.ok(!workspace.includes('aria-hidden="true">◕</span>'),"dashboard must not use a placeholder logo");
});
test("colors and consistent shape are used across the website",()=>{
 assert.match(svg,/fill="#173b35"/);
 assert.match(svg,/stroke="#d3ebb5"/);
 assert.match(css,/\.sm-window-logo-mark\{background:#173b35!important;color:#d3ebb5!important/);
 assert.match(css,/\.sm-window-logo-mark svg\{width:28px/);
 assert.match(css,/\.site-header \.brand-mark\{border-radius:10px\}/);
});
test("Windows ICO embeds PNG images for tray and higher resolutions",()=>{
 assert.equal(icon.readUInt16LE(0),0);
 assert.equal(icon.readUInt16LE(2),1);
 const n=icon.readUInt16LE(4);assert.equal(n,3);
 const sizes=[];
 for(let i=0;i<n;i++){
  const p=6+i*16,width=icon[p]||256,height=icon[p+1]||256;
  assert.equal(width,height);
  const len=icon.readUInt32LE(p+8),off=icon.readUInt32LE(p+12);
  assert.ok(off+len<=icon.length);
  assert.equal(icon.subarray(off,off+8).toString("hex"),"89504e470d0a1a0a");
  sizes.push(width);
 }
 assert.deepEqual(sizes,[16,32,256]);
});
