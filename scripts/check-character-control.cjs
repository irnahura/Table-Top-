// Real API/browser authorization journey. Accounts and DB are disposable.
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const {spawn}=require('node:child_process');
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const folder=fs.mkdtempSync(path.join(os.tmpdir(),'astralis-control-'));
const base='http://127.0.0.1:18083',data=path.join(folder,'game.json');
let server,browser,host;const errors=[];
const wait=ms=>new Promise(resolve=>setTimeout(resolve,ms));
async function start(){server=spawn(path.resolve('output/astralis-test.exe'),[],{cwd:process.cwd(),env:{...process.env,PORT:'18083',ASTRALIS_DATA_FILE:data},windowsHide:true,stdio:'pipe'});for(let i=0;i<100;i++){try{if((await fetch(base+'/health')).ok)return;}catch{}await wait(100);}throw Error('Test server did not start');}
async function stop(){if(server&&!server.killed){const stopped=new Promise(resolve=>server.once('exit',resolve));server.kill();await stopped;}}
async function api(token,endpoint,body,status=200){const r=await fetch(base+'/api/v1'+endpoint,{method:body===undefined?'GET':'POST',headers:{Authorization:'Bearer '+token,'Content-Type':'application/json'},body:body===undefined?undefined:JSON.stringify(body)});assert.equal(r.status,status,await r.clone().text());return r.json();}
async function register(name,email){const credentials={display_name:name,email,password:'test-player-password'};await api('','/auth/register',credentials,201);return api('','/auth/login',credentials);}
async function enter(account,sid,cid){const context=await browser.newContext({viewport:{width:1440,height:950},reducedMotion:'reduce'}),page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));await page.goto(base);await page.evaluate(({account,sid,cid})=>{localStorage.setItem('astralis-token',account.token);localStorage.setItem('astralis-user',JSON.stringify(account.user));localStorage.setItem('astralis-session-'+cid,sid);},{account,sid,cid});await page.reload();await page.goto(base+'/#/table/'+cid);await page.locator('.map-world').waitFor();await page.waitForFunction(()=>document.querySelector('#live-status')?.textContent.includes('Live'));return page;}
async function tab(page,name){await page.locator(`[data-inspector-tab="${name}"]`).click();}
async function command(page,action){const response=page.waitForResponse(r=>r.url().endsWith('/commands')&&r.request().method()==='POST');await action();const r=await response;assert.equal(r.status(),200,await r.text());const result=await r.json();await page.waitForFunction(()=>document.querySelector('#table-status')?.textContent!=='Resolving action…');return result;}
async function assign(page,token,user){await tab(page,'characters');const form=page.locator('[data-tool-form="assign"]');await form.waitFor();await form.getByLabel('Party character to assign').selectOption(token);await form.getByLabel('Character controller').selectOption(user);return command(page,()=>form.getByRole('button',{name:'Save character control',exact:true}).click());}
async function waitMove(page,enabled){await page.waitForFunction(expected=>document.querySelector('[data-move="1,0"]')?.disabled===!expected,enabled);}
(async()=>{try{
 await start();browser=await chromium.launch({headless:true});
 const gm=await api('','/auth/demo',{},201),a=await register('Player A','a@example.test'),b=await register('Player B','b@example.test');
 const sid=gm.snapshot.id,cid=gm.campaign.id,endpoint='/sessions/'+sid+'/commands';
 const invite=await api(gm.token,'/campaigns/'+cid+'/invitations',{},201);await api(a.token,'/invitations/redeem',{token:invite.token});await api(b.token,'/invitations/redeem',{token:invite.token});
 host=await enter(gm,sid,cid);const first=await enter(a,sid,cid),second=await enter(b,sid,cid);
 await waitMove(first,false);await waitMove(second,false);assert.match(await first.locator('#panel-actions').textContent(),/Game master control/);
 await tab(first,'characters');assert.equal(await first.locator('[data-tool-form="assign"]').count(),0);await tab(first,'actions');
 await assign(host,'mayank',a.user.id);await waitMove(first,true);await waitMove(second,false);assert.match(await first.locator('#panel-actions').textContent(),/Controlled by you/);
 const moved=await command(first,()=>first.getByRole('button',{name:'Move east',exact:true}).click());assert.equal(moved.snapshot.state.actors.find(x=>x.id==='mayank').x,3);
 // GM override still works for an assigned character.
 await tab(host,'actions');await waitMove(host,true);await command(host,()=>host.getByRole('button',{name:'Move south',exact:true}).click());
 // Assigned players can attack and interact; results still come from Go rules.
 await api(gm.token,endpoint,{payload:{type:'token.place',token_id:'bandit-scout',x:5,y:3}});
 await first.locator('.map-token[data-actor-id="bandit-scout"]').click();await first.waitForFunction(()=>!document.querySelector('[data-action="attack"]')?.disabled);
 const attack=await command(first,()=>first.getByRole('button',{name:/Attack target/}).click());assert.equal(attack.event_type,'attack.resolved');
 await api(gm.token,endpoint,{payload:{type:'token.place',token_id:'mayank',x:8,y:5}});
 await first.waitForFunction(()=>!document.querySelector('[data-action="interact"]')?.disabled);await command(first,()=>first.getByRole('button',{name:'Investigate bridge',exact:true}).click());
 await command(first,()=>first.getByRole('button',{name:'End turn',exact:true}).click());await waitMove(first,false);
 await tab(host,'characters');await host.locator('#panel-characters [data-actor-id="lira"]').click();await tab(host,'actions');await command(host,()=>host.getByRole('button',{name:'End turn',exact:true}).click());
 // Transfer control: the old assignee loses controls, and forged requests fail.
 await assign(host,'mayank',b.user.id);await waitMove(first,false);await waitMove(second,true);
 await api(a.token,endpoint,{payload:{type:'character.move',character_id:'mayank',controller_id:a.user.id,user_id:gm.user.id,target:{x:9,y:5}}},403);
 // Clear and restore an assignment; the persisted choice survives a restart.
 await assign(host,'mayank','');await waitMove(second,false);
 await assign(host,'mayank',b.user.id);await waitMove(second,true);
 await host.locator('.battle-art').evaluate(img=>img.decode());await host.screenshot({path:'output/screenshots/astralis-character-control-desktop.png',animations:'disabled'});
 await host.setViewportSize({width:390,height:844});await host.getByRole('button',{name:'Open character and encounter panel'}).click();await tab(host,'characters');assert(await host.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));await assign(host,'mayank',b.user.id);await host.locator('[data-tool-form="assign"] button').scrollIntoViewIfNeeded();await host.screenshot({path:'output/screenshots/astralis-character-control-mobile.png',animations:'disabled'});
 await stop();await start();await second.reload();await second.locator('.map-world').waitFor();await waitMove(second,true);
 assert.match(await second.locator('#panel-actions').textContent(),/Controlled by you/);
 const snapshot=await api(gm.token,'/sessions/'+sid+'/snapshot');assert.equal(snapshot.state.actors.find(x=>x.id==='mayank').controller_id,b.user.id);
 assert.deepEqual(errors,[]);console.log('Character control checks passed: GM assignment/override, player movement/attack/interaction/end-turn, other-player restrictions, reassignment, clearing, forged request rejection, persistence and desktop/mobile layouts.');
 }catch(error){await host?.screenshot({path:'output/screenshots/character-control-failure.png'});throw error;}finally{await browser?.close();await stop();fs.rmSync(folder,{recursive:true,force:true});}})().catch(error=>{console.error(error);process.exitCode=1;});
