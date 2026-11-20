// Dependency free contract checks for the production frontend and Go API shapes.
const fs = require('node:fs');
const vm = require('node:vm');
const assert = require('node:assert/strict');
const listeners = {};
const storage = new Map();
const node = () => ({innerHTML:'',textContent:'',classList:{add(){},remove(){},toggle(){}},querySelectorAll(){return []},setAttribute(){}});
const nodes = new Map(['#app','#legal-dialog','#legal-content','#toast','#table-status'].map(key=>[key,node()]));
let calls = [], replies = [];
const context = {
  document:{querySelector:key=>nodes.get(key)||null,querySelectorAll:()=>[],addEventListener:(type,fn)=>listeners[type]=fn},
  window:{addEventListener(){},scrollTo(){}},location:{hash:'#/register'},
  localStorage:{getItem:key=>storage.get(key)||null,setItem:(key,value)=>storage.set(key,value),removeItem:key=>storage.delete(key)},
  matchMedia:()=>({matches:true}),setTimeout:()=>0,clearTimeout(){},setInterval:()=>0,
  FormData:class { constructor(form){ return Object.entries(form.values)[Symbol.iterator](); } },
  fetch:async(path,options)=>{calls.push({path,...options});const reply=replies.shift();if(!reply)throw Error('Unexpected request '+path);return {ok:reply.ok!==false,status:reply.status||200,json:async()=>reply.data};}
};
const source = fs.readFileSync('frontend/app.js','utf8');
const start = source.lastIndexOf("  window.addEventListener('hashchange'");
assert(start>0);
vm.runInNewContext(source.slice(0,start)+'globalThis.testAPI={state,createSession,resumeSession,token,command,landing,guidePage,canMove,eventDescription,renderTable,tacticalMap,previewCharacterCopy,authPage,campaignForm,campaignList};})();',context);
const api = context.testAPI;
const snapshot={id:'session-qa',sequence:0,state:{width:20,height:14,round:1,active_actor_id:'mayank',actors:[{id:'mayank',name:'Mayank',x:2,y:2,hp:82,max_hp:100,movement:6,attack_bonus:5,damage:8}],blocked:[{x:7,y:4}]}};
(async()=>{
  const error=node(),button=node();
  const form={id:'auth-form',values:{display_name:'QA',email:'qa@example.test',password:'password123'},reportValidity:()=>true,querySelector:selector=>selector==='#auth-error'?error:button};
  replies=[{data:{user:{id:'qa'}}},{data:{token:'qa-token',user:{id:'qa',display_name:'QA'}}}];
  await listeners.submit({target:form,preventDefault(){}});
  assert.equal(error.textContent,'');
  assert.equal(calls[0].path,'/api/v1/auth/register');
  assert.equal(calls[1].path,'/api/v1/auth/login');
  assert.equal(storage.get('astralis-token'),'qa-token');
  const campaign={id:'campaign-qa',owner_id:'qa',name:'QA world'};
  api.state.campaigns=[campaign];
  api.state.selectedCampaign=campaign;
  replies=[{data:{map:{id:'map-qa'}}},{data:{session_id:'session-qa',snapshot}}];
  await api.createSession(campaign);
  assert.equal(storage.get('astralis-session-campaign-qa'),'session-qa');
  assert.equal(api.state.snapshot.id,'session-qa');
  assert.equal(context.location.hash,'/table/campaign-qa');
  assert.match(api.token(snapshot.state.actors[0],true,false),/left:12.5%;top:17.857/);
  replies=[{data:snapshot}];
  await api.resumeSession(campaign);
  assert.equal(calls.at(-1).path,'/api/v1/sessions/session-qa/snapshot');
  const previous=api.state.snapshot;
  replies=[{ok:false,status:409,data:{error:'BLOCKED_PATH: target is blocked'}}];
  await api.command({type:'character.move',character_id:'mayank',target:{x:7,y:4}});
  assert.equal(api.state.snapshot,previous,'Rejected commands must preserve server snapshot');
  assert.match(nodes.get('#table-status').textContent,/BLOCKED_PATH/);
  assert.equal(api.state.commandPending,false);
  const hero=snapshot.state.actors[0];
  assert.equal(api.canMove(hero,1,0),true);
  assert.equal(api.canMove({...hero,x:0},-1,0),false,'Map edge must disable movement');
  assert.equal(api.canMove({...hero,x:6,y:4},1,0),false,'Blocked terrain must disable movement');
  assert.equal(api.canMove({...hero,id:'lira'},1,0),false,'Inactive characters must not move');
  assert.equal(api.canMove({...hero,hp:0},1,0),false,'Defeated characters must not move');
  api.state.user={id:'player'};
  assert.equal(api.canMove(hero,1,0),false,'Players cannot move unassigned tokens');
  assert.equal(api.canMove({...hero,controller_id:'other'},1,0),false,'Players cannot move other players tokens');
  assert.equal(api.canMove({...hero,controller_id:'player'},1,0),true,'Assigned players can move on their turn');
  api.state.user={id:'qa'};
  assert.match(api.tacticalMap(),/class="grid-lines"/,'Grid toggle must affect an actual SVG grid layer');
  const attack={event_type:'attack.resolved',payload:{attacker_id:'mayank',target_id:'bandit',roll:14,total:19,hit:true,damage:8,remaining_hp:22}};
  assert.match(api.eventDescription(attack),/Roll 14 \+ 5 = 19/);
  assert.match(api.eventDescription(attack),/Hit for 8 damage/);
  replies=[{data:{snapshot:{...snapshot,sequence:1,events:[{...attack,sequence:1}]},...attack}}];
  await api.command({type:'combat.attack',attacker_id:'mayank',target_id:'bandit'});
  assert.match(api.state.tableMessage,/Hit for 8 damage/);
  assert.match(nodes.get('#app').innerHTML,/Encounter log/);
  assert.match(nodes.get('#app').innerHTML,/Character inspector/);
  assert.match(api.landing(),/data-tagline/);
  assert.match(api.landing(),/Create your campaign/);
  assert.match(api.landing(),/virtual tabletop/);
  assert.match(api.guidePage(),/Game master \(GM\)/);
  assert.match(api.guidePage(),/Nonplayer character \(NPC\)/);
  assert.match(api.guidePage(),/roll party initiative on the server/);
  assert.match(api.guidePage(),/data-guide-section="guide-3"/);
  assert.match(api.landing(),/product-preview/);
  assert.match(api.landing(),/data-preview-mode="tactical"/);
  assert.doesNotMatch(api.landing(),/assets\/astralis-guide.png/);
  assert.doesNotMatch(api.landing(),/data-webgl-background/);
  assert.match(api.tacticalMap(),/class="battle-art" src="\/assets\/riverside-battlemap.png"/);
  assert.match(api.previewCharacterCopy('Mayank'),/82 \/ 100 HP/);
  assert.match(api.previewCharacterCopy('<script>'),/^Select a character/);
  assert.match(api.campaignForm(),/id="campaign-form"/);
  assert.match(api.authPage('register'),/id="auth-form"/);
  assert.match(api.campaignList([{id:'unsafe',name:'<script>alert(1)</script>'}]),/&lt;script&gt;/);
  for(const asset of ['frontend/assets/astralis-guide.png','frontend/assets/riverside-battlemap.png'])assert(fs.existsSync(asset));
  console.log('Frontend contracts passed: registration, login, session creation, resume, coordinates, rejected commands, landing.');
})().catch(error=>{console.error(error);process.exitCode=1;});
