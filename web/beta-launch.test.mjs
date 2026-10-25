import test from "node:test";
import assert from "node:assert/strict";
import {readFileSync} from "node:fs";

const source=readFileSync(new URL("./index.html",import.meta.url),"utf8");
const workspace=readFileSync(new URL("./workspace.mjs",import.meta.url),"utf8");
const issue="https://github.com/marceli1404/sloth-mode/issues/30";
const windowsBuild="https://github.com/marceli1404/sloth-mode/actions/runs/40133971950";

test("beta announcement only displays on public marketing route",()=>{
 assert.match(source,/class="sm-beta-ribbon" aria-label="Sloth Mode public beta"/);
 assert.match(source,/html\[data-sloth-page="dashboard"\] \.sm-beta-ribbon\{display:none!important\}/);
 assert.match(source,/PUBLIC BETA IS OPEN/);
 assert.match(source,/href="\/dashboard">Try the web workspace/);
 assert.match(source,/href="\/dashboard">Try Sloth Mode beta/);
});
test("unsupported and security-uncleared native installer is labelled honestly",()=>{
 assert.ok(source.includes(windowsBuild));
 assert.match(source,/id="bridge-setup" class="bridge-button secondary">Set up local Gemma<\/button>/);
 assert.match(source,/The first model download needs your approval and uses several GB, depending on model quality\./);
 assert.match(source,/Approve a model download of several GB, depending on model quality/);
 assert.match(source,/Download Gemma 3 \(several GB, depending on model quality\)/);
 assert.doesNotMatch(source,/Gemma 3 \(~2\.5 GB\)|Approve the optional ~2\.5 GB model download|Install Gemma \(~2\.5 GB\)|Install Gemma \(several GB, depending on model quality\)/);
 assert.match(source,/Windows installer is unsigned and experimental/);
 assert.match(source,/Native security review and wider device testing are still open/);
 assert.match(source,/GitHub login may be required/);
 assert.doesNotMatch(source,/SIGN(?:ED|ED AND VERIFIED) STABLE/);
});
test("visible feedback routes to correct public issue without third-party analytics",()=>{
 assert.ok(source.includes(issue));
 assert.ok(workspace.includes(issue));
 assert.match(workspace,/sm-window-feedback/);
 assert.match(workspace,/rel="noopener noreferrer"/);
 assert.match(source,/target="_blank" rel="noopener noreferrer"/);
 assert.doesNotMatch(source,/analytics\.google\.com|googletagmanager\.com/);
});
test("existing app bootstrap and private account integrations are untouched",()=>{
 assert.match(source,/\/api\/tasks/);
 assert.match(source,/\/auth\/status/);
 assert.match(source,/<script type="module" src="\/browser-ai\.mjs"><\/script>/);
 assert.match(source,/<script type="module" src="\/workspace\.mjs"><\/script>/);
});
