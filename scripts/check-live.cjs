// Live transport checks against a real server, using disposable accounts/storage.
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const {spawn}=require('node:child_process');
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const folder=fs.mkdtempSync(path.join(os.tmpdir(),'astralis-live-'));
const base='http://127.0.0.1:18082',data=path.join(folder,'game.json');
const wait=ms=>new Promise(resolve=>setTimeout(resolve,ms));
let server,browser;
const errors=[];
async function start(){server=spawn(path.resolve('output/astralis-test.exe'),[],{cwd:process.cwd(),env:{...process.env,PORT:'18082',ASTRALIS_DATA_FILE:data},windowsHide:true,stdio:'pipe'});for(let i=0;i<100;i++){try{if((await fetch(base+'/health')).ok)return;}catch{}await wait(100);}throw Error('Test server did not start');}
async function stop(){if(server&&!server.killed){const stopped=new Promise(resolve=>server.once('exit',resolve));server.kill();await stopped;}}
async function request(token,endpoint,body){const r=await fetch(base+'/api/v1'+endpoint,{method:body===undefined?'GET':'POST',headers:{Authorization:'Bearer '+token,'Content-Type':'application/json'},body:body===undefined?undefined:JSON.stringify(body)});assert(r.ok,await r.clone().text());return r.json();}
async function enter(context,account,sid,cid,live=true){const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));await page.goto(base);await page.evaluate(({account,sid,cid})=>{localStorage.setItem('astralis-token',account.token);localStorage.setItem('astralis-user',JSON.stringify(account.user));localStorage.setItem('astralis-session-'+cid,sid);},{account,sid,cid});await page.reload();await page.goto(base+'/#/table/'+cid);await page.locator('.map-world').waitFor();if(live)await page.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('Live'));await page.locator('[data-inspector-tab="tools"]').click();return page;}
(async()=>{try{
 await start();browser=await chromium.launch({headless:true});
 const gm=await request('','/auth/demo',{}),player=await request('','/auth/demo',{}),sid=gm.snapshot.id,cid=gm.campaign.id;
 const invitation=await request(gm.token,'/campaigns/'+cid+'/invitations',{});await request(player.token,'/invitations/redeem',{token:invitation.token});
 const gmContext=await browser.newContext({viewport:{width:1440,height:900},reducedMotion:'reduce'}),playerContext=await browser.newContext({viewport:{width:1100,height:850},reducedMotion:'reduce'});
 const host=await enter(gmContext,gm,sid,cid),guest=await enter(playerContext,player,sid,cid);
 await host.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('2 online'));
 // Live updates arrive while a form draft remains focused and intact.
 const notes=host.locator('[data-tool-form="notes"] textarea');await notes.fill('Unsaved notes stay here.');
 await guest.locator('[data-tool-form="chat"] textarea').fill('Arrived over WebSocket');
 const sent=guest.waitForResponse(r=>r.url().endsWith('/commands'));
 await guest.getByRole('button',{name:'Send message',exact:true}).click();assert.equal((await sent).status(),200);
 await wait(250);assert.equal(await notes.inputValue(),'Unsaved notes stay here.');assert(await notes.evaluate(el=>el===document.activeElement));
 await notes.blur();await host.waitForFunction(()=>document.querySelector('.table-chat')?.textContent.includes('Arrived over WebSocket'),null,{timeout:2500});assert.equal(await notes.inputValue(),'Unsaved notes stay here.');
 // Lose the first response after the server accepts a roll. The browser retries
 // the same command ID and receives the same outcome without another event.
 const before=await request(gm.token,'/sessions/'+sid+'/snapshot');let lost=false;const bodies=[];
 await host.route('**/api/v1/sessions/*/commands',async route=>{bodies.push(route.request().postDataJSON());if(!lost){lost=true;const response=await route.fetch();assert.equal(response.status(),200);await route.abort('failed');}else await route.continue();});
 const retry=host.waitForResponse(r=>r.url().endsWith('/commands'));
 await host.getByRole('button',{name:'Roll dice',exact:true}).click();const retried=await(await retry).json();assert.equal(retried.duplicate,true);assert.equal(retried.sequence_number,before.sequence+1);assert.equal(bodies.length,2);assert.equal(bodies[0].command_id,bodies[1].command_id);assert.equal(bodies[0].expected_state_version,before.state_version);
 await host.unroute('**/api/v1/sessions/*/commands');await host.waitForFunction(()=>!document.querySelector('#table-status')?.textContent.includes('Resolving'));
 // Leave both pages open during a server restart. Reconnect must restore live
 // updates and presence without a reload or another sign-in.
 await stop();await host.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('Reconnecting'));await start();
 await host.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('2 online'),null,{timeout:15000});await guest.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('Live'),null,{timeout:15000});
 await request(gm.token,'/sessions/'+sid+'/commands',{payload:{type:'chat.send',message:'Recovered after restart'}});
 await guest.waitForFunction(()=>document.querySelector('.table-chat')?.textContent.includes('Recovered after restart'),null,{timeout:2500});
 await host.locator('.battle-art').evaluate(img=>img.decode());
 await host.screenshot({path:'output/screenshots/astralis-live-desktop.png',animations:'disabled'});
 await guest.setViewportSize({width:390,height:844});assert(await guest.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
 // Leaving the table closes its connection and updates the other user's count.
 await guest.goto(base+'/#/library');await host.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('1 online'),null,{timeout:3000});
 // An unavailable WebSocket transport still receives real server state by polling.
 const fallbackContext=await browser.newContext();await fallbackContext.addInitScript(()=>{window.WebSocket=undefined;});
 const fallback=await enter(fallbackContext,player,sid,cid,false);
 await request(gm.token,'/sessions/'+sid+'/commands',{payload:{type:'chat.send',message:'Polling fallback still works'}});
 await fallback.waitForFunction(()=>document.querySelector('.table-chat')?.textContent.includes('Polling fallback still works'),null,{timeout:7500});
 assert.deepEqual(errors,[]);
 console.log('Live browser checks passed: two-member presence, live chat, draft preservation, lost-response retry without reroll, reconnect after server restart, mobile overflow, route cleanup and polling fallback.');
 }finally{await browser?.close();await stop();fs.rmSync(folder,{recursive:true,force:true});}})().catch(error=>{console.error(error);process.exitCode=1;});
