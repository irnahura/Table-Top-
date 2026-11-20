// Isolated UI fixtures. Does not read personal browser profiles or change saved game data.
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const fs=require('node:fs');
const assert=require('node:assert/strict');
const base=process.env.ASTRALIS_URL||'http://localhost:8080';
const output='output/screenshots';
fs.mkdirSync(output,{recursive:true});
const user={id:'ui-user',display_name:'Mira',email:'mira@example.test'};
let campaigns=[],sequence=0;
const snapshot={id:'ui-session',sequence:0,state_version:1,state:{width:20,height:14,round:1,active_actor_id:'mayank',blocked:[{x:7,y:4},{x:7,y:5},{x:7,y:6},{x:8,y:6},{x:9,y:6}],actors:[{id:'mayank',name:'Mayank',x:2,y:2,hp:82,max_hp:100,movement:6,attack_bonus:5,damage:8},{id:'lira',name:'Lira',x:4,y:3,hp:100,max_hp:100,movement:6,attack_bonus:4,damage:6},{id:'bandit-scout',name:'Bandit Scout',x:10,y:5,hp:30,max_hp:30,hostile:true},{id:'bandit-archer',name:'Bandit Archer',x:13,y:4,hp:24,max_hp:24,hostile:true}]},events:[]};
const pageErrors=[];
async function attach(context){
  context.on('page',page=>page.on('pageerror',error=>pageErrors.push(error.message)));
  await context.route('**/api/v1/**',async route=>{
    const request=route.request(),path=new URL(request.url()).pathname.replace('/api/v1',''),method=request.method();
    const body=request.postDataJSON()||{};let response={},status=200;
    if(path==='/auth/register'){response={user};status=201;}
    else if(path==='/auth/login')response={token:'ui-token',user};
    else if(path==='/campaigns'&&method==='GET')response={campaigns};
    else if(path==='/campaigns'&&method==='POST'){const campaign={id:'ui-campaign',owner_id:user.id,...body};campaigns.push(campaign);response={campaign};status=201;}
    else if(path==='/campaigns/ui-campaign/members'&&method==='GET')response={members:[{id:user.id,display_name:user.display_name,role:'owner'}]};
    else if(path==='/maps'){response={map:{id:'ui-map'}};status=201;}
    else if(path==='/campaigns/ui-campaign/sessions'){response={session_id:snapshot.id,snapshot};status=201;}
    else if(path==='/sessions/ui-session/snapshot')response=snapshot;
    else if(path==='/sessions/ui-session/commands'){
      const cmd=body.payload,a=snapshot.state.actors.find(actor=>actor.id===(cmd.character_id||cmd.attacker_id));let type,payload;
      if(cmd.type==='character.move'){const from={x:a.x,y:a.y};a.x=cmd.target.x;a.y=cmd.target.y;type='character.moved';payload={character_id:a.id,from,to:cmd.target};}
      else if(cmd.type==='combat.attack'){const target=snapshot.state.actors.find(actor=>actor.id===cmd.target_id);target.hp=Math.max(0,target.hp-8);type='attack.resolved';payload={attacker_id:a.id,target_id:target.id,roll:14,total:19,hit:true,damage:8,remaining_hp:target.hp};}
      else if(cmd.type==='turn.end'){snapshot.state.round++;type='turn.advanced';payload={round:snapshot.state.round,active_actor_id:'mayank'};}
      else{type='object.interacted';payload={character_id:a.id,result:'The bridge bears fresh wagon marks.'};}
      snapshot.sequence=++sequence;snapshot.state_version++;snapshot.events.push({sequence,event_type:type,payload});response={snapshot,event_type:type,payload};
    }else{status=404;response={error:'Unexpected UI fixture request: '+method+' '+path};}
    await route.fulfill({status,contentType:'application/json',body:JSON.stringify(response)});
  });
}
async function noOverflow(page,label){assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),label+' has horizontal overflow');}
async function shot(page,name){await page.waitForFunction(()=>Array.from(document.images).filter(img=>{const r=img.getBoundingClientRect();return r.width&&r.height&&r.top<innerHeight&&r.bottom>0;}).every(img=>img.complete&&img.naturalWidth>0));await page.screenshot({path:output+'/'+name+'.png',animations:'disabled'});}
(async()=>{
  const browser=await chromium.launch({headless:true,args:['--enable-unsafe-swiftshader']});
  try{
    const desktop=await browser.newContext({viewport:{width:1440,height:900},reducedMotion:'reduce'});await attach(desktop);
    const page=await desktop.newPage();
    await page.goto(base,{waitUntil:'networkidle'});
    await page.locator('.hero-heading').waitFor();
    assert.equal(await page.locator('.final-section img').count(),0,'Final CTA must not contain the portrait');
    assert.equal(await page.locator('canvas').count(),0,'The product preview replaces the WebGL backdrop');
    await noOverflow(page,'Desktop landing');await shot(page,'astralis-home-desktop');
    await page.locator('main img').evaluateAll(images=>Promise.all(images.map(async img=>{img.loading='eager';await img.decode();})));await page.screenshot({path:output+'/astralis-home-full.png',fullPage:true,animations:'disabled'});
    await page.locator('.faq-question').first().click();assert(await page.locator('.faq-item').first().getAttribute('open')!==null);await page.locator('.faq-question').first().click();
    await page.locator('[data-legal="privacy"]').click();assert(await page.locator('#legal-dialog').isVisible());await page.locator('[data-close-dialog]').click();
    await page.locator('[data-preview-mode="tactical"]').click();
    assert(await page.locator('[data-preview-board]').evaluate(node=>node.classList.contains('preview-tactical')));
    await page.locator('.demo-token[data-preview-character="Mayank"]').click();assert.match(await page.locator('#preview-character').textContent(),/82 \/ 100 HP/);
    await page.locator('.encounter-preview').screenshot({path:output+'/astralis-encounter-preview.png',animations:'disabled'});
    await page.getByRole('tab',{name:'Characters',exact:true}).click();assert(await page.locator('#feature-panel-characters').isVisible());assert(!(await page.locator('#feature-panel-maps').isVisible()));await page.getByRole('tab',{name:'Combat & log',exact:true}).click();assert(await page.locator('#feature-panel-combat').isVisible());await page.getByRole('tab',{name:'Battle maps',exact:true}).click();
    await page.locator('.nav-dropdown>summary').click();await page.locator('.nav-dropdown').getByRole('button',{name:'Try the tabletop preview',exact:true}).click();assert(await page.locator('[data-preview-tour]').isVisible());await page.locator('.demo-token[data-preview-character="Mayank"]').click();assert.match(await page.locator('#preview-tour-title').textContent(),/Enable the tactical grid/);await page.locator('[data-preview-mode="tactical"]').click();await page.locator('.demo-token[data-preview-character="Bandit Scout"]').click();assert.equal(await page.locator('#preview-tour-title').textContent(),'Preview complete');assert(await page.locator('[data-tour-campaign]').isVisible());await page.getByRole('button',{name:'Close preview walkthrough',exact:true}).click();
    await page.goto(base+'/#/getting-started');await page.locator('.getting-started-steps').waitFor();assert.equal(await page.locator('.getting-started-steps>li').count(),7);await noOverflow(page,'Getting-started guide');await shot(page,'astralis-getting-started-desktop');
    await page.locator('.nav-dropdown>summary').click();await page.locator('.nav-dropdown').getByRole('button',{name:'Try the tabletop preview',exact:true}).click();await page.waitForURL(base+'/#/');await page.locator('[data-preview-tour]:not([hidden])').waitFor();assert.equal(await page.locator('#preview-tour-title').textContent(),'Select Mayank');await page.getByRole('button',{name:'Close preview walkthrough',exact:true}).click();
    await page.goto(base+'/#/register');await shot(page,'astralis-register-desktop');
    await page.getByLabel('Display name').fill('Mira');await page.getByLabel('Email address').fill(user.email);await page.getByLabel('Password').fill('test-password');await page.getByRole('button',{name:'Create account'}).click();
    await page.waitForURL('**/#/library');await page.locator('.library-empty').waitFor();await noOverflow(page,'Campaign library');
    await page.getByRole('link',{name:'New campaign',exact:true}).click();await page.getByLabel('Campaign name').fill('The Ashen Vale');await page.getByLabel('The setting').fill('Fresh tracks cross the old bridge.');await page.getByRole('button',{name:'Create campaign'}).click();
    await page.waitForURL('**/#/campaign/ui-campaign');await page.locator('.campaign-stage').waitFor();await shot(page,'astralis-lobby-desktop');
    await page.getByRole('button',{name:'Launch session',exact:true}).click();await page.waitForURL('**/#/table/ui-campaign');await page.locator('.battle-art').evaluate(img=>img.decode());
    await shot(page,'astralis-tabletop-desktop');await noOverflow(page,'Desktop tabletop');
    await page.getByRole('button',{name:'Zoom in',exact:true}).click();assert.equal(await page.locator('#map-zoom-level').textContent(),'125%');await page.getByRole('button',{name:'Pan map',exact:true}).click();const area=await page.locator('.map-viewport').boundingBox();await page.mouse.move(area.x+area.width*.6,area.y+area.height*.75);await page.mouse.down();await page.mouse.move(area.x+area.width*.6-100,area.y+area.height*.75,{steps:5});await page.mouse.up();assert(await page.locator('.map-viewport').evaluate(node=>node.scrollLeft>0),'Pan tool must drag the map');await page.getByRole('button',{name:'Zoom out',exact:true}).click();assert.equal(await page.locator('#map-zoom-level').textContent(),'100%');
    await page.getByRole('tab',{name:'Characters',exact:true}).click();assert(await page.locator('#panel-characters').isVisible());assert(!(await page.locator('#panel-actions').isVisible()));await page.getByRole('tab',{name:'Actions',exact:true}).click();
    for(let i=0;i<6;i++){await page.getByRole('button',{name:'Move east',exact:true}).click();await page.waitForFunction(x=>document.querySelector('#table-status')?.textContent.includes('moved to '+x+', 2'),i+3);}
    await page.locator('.map-token[data-actor-id="bandit-scout"]').click();await page.getByRole('button',{name:/^Attack target/}).click();
    await page.waitForFunction(()=>document.querySelector('#table-status')?.textContent.includes('Hit for 8 damage'));assert.match(await page.locator('.encounter-log').textContent(),/Roll 14 \+ 5 = 19/);
    await page.getByRole('tab',{name:/^Log/}).click();await shot(page,'astralis-encounter-log');assert.equal(await page.locator('.roll-result>.heading-4').first().textContent(),'19');await page.getByRole('tab',{name:'Actions',exact:true}).click();
    await page.getByRole('button',{name:'End turn',exact:true}).click();await page.waitForFunction(()=>document.querySelector('.table-toolbar')?.textContent.includes('Round 2'));
    await page.locator('[data-table-action="grid"]').click();assert(await page.locator('.table-map').evaluate(node=>node.classList.contains('grid-hidden')));
    campaigns.push({id:'ui-campaign-2',name:'Riverside Watch',description:'The evening patrol.'},{id:'ui-campaign-3',name:'The Old Crossing',description:'A campaign at the village bridge.'});
    await page.goto(base+'/#/library');await page.locator('.campaign-card').first().waitFor();await shot(page,'astralis-campaigns-desktop');await page.getByLabel('Search campaigns').fill('watch');await page.waitForFunction(()=>document.querySelectorAll('.campaign-card').length===1);assert.match(await page.locator('.campaign-card').textContent(),/Riverside Watch/);await page.getByLabel('Search campaigns').fill('missing campaign');await page.getByRole('button',{name:'Clear search',exact:true}).click();assert.equal(await page.locator('.campaign-card').count(),3);await page.getByLabel('Sort campaigns').selectOption('name');assert.equal(await page.locator('.campaign-card .heading-5').first().textContent(),'Riverside Watch');
    const mobile=await browser.newContext({viewport:{width:390,height:844},isMobile:true,hasTouch:true,reducedMotion:'reduce'});await attach(mobile);const phone=await mobile.newPage();
    await phone.goto(base,{waitUntil:'networkidle'});await phone.locator('.hero-heading').waitFor();await noOverflow(phone,'390px landing');const cta=await phone.locator('.hero-primary').boundingBox();assert(cta&&cta.y+cta.height<=844,'Primary action must fit on first mobile screen');await shot(phone,'astralis-home-mobile');
    await phone.getByRole('button',{name:'Open menu',exact:true}).click();await phone.locator('.mobile-menu').getByRole('link',{name:'RPG guide',exact:true}).click();await phone.waitForURL('**/#/guide');await noOverflow(phone,'390px RPG guide');
    await phone.goto(base+'/#/getting-started');await noOverflow(phone,'390px getting-started guide');await shot(phone,'astralis-getting-started-mobile');await phone.goto(base);await phone.getByRole('tab',{name:'Characters',exact:true}).click();await noOverflow(phone,'390px character feature preview');
    await phone.goto(base+'/#/login');await phone.getByLabel('Email address').fill(user.email);await phone.getByLabel('Password').fill('test-password');await phone.getByRole('button',{name:'Sign in',exact:true}).click();await phone.waitForURL('**/#/library');await phone.getByRole('link',{name:'Open The Ashen Vale',exact:true}).click();await phone.getByRole('button',{name:'Launch session',exact:true}).click();await phone.waitForURL('**/#/table/ui-campaign');
    await noOverflow(phone,'390px tabletop');assert(await phone.locator('.map-viewport').evaluate(node=>node.scrollWidth>node.clientWidth),'Mobile map must support horizontal exploration');await phone.locator('.map-viewport').evaluate(node=>node.scrollLeft=200);assert(await phone.locator('.map-viewport').evaluate(node=>node.scrollLeft===200));await shot(phone,'astralis-tabletop-mobile');await phone.getByRole('button',{name:'Fit map',exact:true}).click();assert(await phone.locator('.map-viewport').evaluate(node=>node.scrollWidth===node.clientWidth),'Fit map must fit the mobile canvas');await phone.getByRole('button',{name:'Open character and encounter panel',exact:true}).click();assert(await phone.locator('.table-sidebar').isVisible());await shot(phone,'astralis-character-mobile');await phone.getByRole('button',{name:'Close character panel',exact:true}).click();assert(!(await phone.locator('.table-sidebar').isVisible()));
    assert.deepEqual(pageErrors,[],'Unexpected browser runtime errors');
    console.log('VTT browser checks passed: desktop and 390px layouts, campaign search/sort, sidebar tabs, map pan/zoom/fit, preview, registration, campaign/session creation, movement, combat, turns, grid, and mobile panels.');
  }finally{await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
