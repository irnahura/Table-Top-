/* Contextual editor controls. All saved actions use the existing command API. */
window.AstralisEditor=function(ctx){
  const {state,command,renderTable,escapeHTML:e,notify,selectInspectorTab,controlsActor}=ctx;
  const world=()=>state.snapshot?.state||{};
  let lastTap={id:'',at:0},dragTurn='';
  const btn=(label,action,attrs='')=>`<button type="button" data-editor="${action}" ${attrs}>${label}</button>`;
  const icon=(symbol,label,action)=>`<button type="button" class="map-tool editor-rail-button" data-editor="${action}" aria-label="${label}" title="${label}"><span aria-hidden="true">${symbol}</span><span class="caption-4">${label.replace('Open ','').replace(' tools','')}</span></button>`;
  function close(){state.editorPanel='';state.editorToken='';state.canvasDrawMenu=false;document.querySelectorAll('.editor-popover,.editor-token-hud,.editor-scenes,.canvas-draw-menu').forEach(el=>el.remove());const tools=document.querySelector('.editor-secondary-tools');if(tools)tools.hidden=true;state.editorMore=false;syncRail();}
  function syncRail(){document.querySelectorAll('.editor-rail-button').forEach(el=>{el.setAttribute('aria-pressed',String(el.dataset.editor===state.editorPanel||el.dataset.editor==='more'&&(state.editorMore||['block','fog','place','vtt-fx'].includes(state.mapMode))));if(el.dataset.editor==='more')el.setAttribute('aria-expanded',String(Boolean(state.editorMore)));});const draw=document.querySelector('.game-toolbar [data-editor-draw]');if(draw){const active=['line','rectangle','vtt-freehand','vtt-text','vtt-pin','vtt-ellipse','vtt-cone','vtt-line','vtt-rectangle'].includes(state.mapMode);draw.classList.toggle('is-active',active);draw.setAttribute('aria-pressed',String(active));}}
  function rail(){
    const rail=document.querySelector('.game-toolbar'),map=document.querySelector('.table-map');if(!rail||!map||rail.dataset.grouped)return;rail.dataset.grouped='true';
    const primary=new Set(['select','pan','vtt-freehand','measure','ping']);
    const extras=document.createElement('div');extras.className='editor-secondary-tools';extras.hidden=!state.editorMore;extras.setAttribute('role','toolbar');extras.setAttribute('aria-label','Secondary map tools');
    extras.innerHTML='<div class="editor-menu-heading"><span>More tools</span><button type="button" data-editor="close" aria-label="Close more tools">×</button></div>';
    const grid=document.createElement('div');grid.className='editor-tool-grid';extras.append(grid);
    [...rail.querySelectorAll('[data-map-tool],[data-vtt-open]')].forEach(el=>{if(!primary.has(el.dataset.mapTool))grid.append(el);});
    map.append(extras);rail.querySelectorAll('.tool-divider').forEach(el=>el.remove());
    const extrasMarkup=`${icon('⚄','Open dice','dice')}${icon('☷','Turn order','turns')}${state.isGM?icon('▱','Scenes','scenes')+icon('▧','Layer tools','layers'):''}${icon('•••','More tools','more')}`;
    const ping=rail.querySelector('[data-map-tool="ping"]');ping?.insertAdjacentHTML('afterend',extrasMarkup);
    rail.querySelectorAll('.map-tool').forEach(el=>{if(!el.title)el.title=el.getAttribute('aria-label')||el.textContent.trim();});
    const draw=rail.querySelector('[data-map-tool="vtt-freehand"]');if(draw)draw.dataset.editorDraw='true';
    syncRail();
  }
  function sidebar(){
    const panel=document.querySelector('#panel-table');if(!panel)return;
    const nav=panel.querySelector('.vtt-section-nav');if(!nav||panel.querySelector('.editor-section-picker'))return;
    const select=document.createElement('select');select.className='editor-section-picker';select.setAttribute('aria-label','Workspace section');
    [...nav.querySelectorAll('[data-vtt-section]')].forEach(button=>{const option=new Option(button.textContent.trim(),button.dataset.vttSection);option.selected=button.getAttribute('aria-pressed')==='true';select.append(option);});
    select.addEventListener('change',()=>panel.querySelector(`[data-vtt-section="${CSS.escape(select.value)}"]`)?.click());
    nav.before(select);nav.querySelector('[aria-pressed="true"]')?.scrollIntoView({block:'nearest',inline:'nearest'});
  }
  function popover(){
    document.querySelector('.editor-popover')?.remove();const type=state.editorPanel;if(!['dice','layers'].includes(type))return;
    const w=world(),brush=state.vttDrafts?.brush||{};
    let body='';
    if(type==='dice'){
      const last=[...(state.snapshot?.events||[])].reverse().find(ev=>ev.event_type==='dice.roll');
      body=`<div class="editor-dice-grid">${[4,6,8,10,12,20,100].map(n=>btn(`d${n}`,'roll',`data-sides="${n}" aria-label="Roll d${n}"`)).join('')}</div><form data-editor-form="dice"><label>Dice count<input aria-label="Quick dice count" name="count" type="number" min="1" max="20" required value="${state.quickDiceCount||1}"></label><label>Bonus<input aria-label="Quick dice bonus" name="bonus" type="number" min="-100" max="100" required value="${state.quickDiceBonus||0}"></label><button>Roll d20</button></form>${last?`<div class="editor-roll-result"><strong>${e(last.payload.total)}</strong><span>${e(last.payload.formula)}<small>${e((last.payload.rolls||[]).join(' + '))}</small></span></div>`:'<p class="editor-muted">Server rolls are public and saved in the encounter log.</p>'}`;
    }else body=`<p class="editor-muted">Choose the layer for new drawings. Token layers are changed from the token menu.</p><div class="editor-layers">${[['tokens','◉','Tokens','Shared characters and drawings'],['gm','◇','GM private','Visible only to the game master'],['map','▧','Map','Background decorations'],['foreground','▱','Foreground','Above tokens']].map(([id,symbol,title,description])=>`<button type="button" data-editor="layer" data-layer="${id}" aria-pressed="${(brush.layer||'tokens')===id}"><span>${symbol}</span><span>${title}<small>${description}</small></span><i>${(brush.layer||'tokens')===id?'✓':''}</i></button>`).join('')}</div>`;
    document.querySelector('.table-map')?.insertAdjacentHTML('beforeend',`<section class="editor-popover" aria-label="${type==='dice'?'Quick dice tray':'Drawing layers'}"><header><div><span class="editor-kicker">${type==='dice'?'ROLL FOR IT':'WORKSPACE'}</span><h2>${type==='dice'?'Dice tray':'Drawing layer'}</h2></div>${btn('×','close',`aria-label="Close ${type==='dice'?'dice tray':'layer menu'}"`)}</header>${body}</section>`);
  }
  function scenes(){
    document.querySelector('.editor-scenes')?.remove();if(state.editorPanel!=='scenes'||!state.isGM)return;
    const w=world(),page=(id,name,image,active)=>`<button type="button" class="editor-scene-card ${active?'is-current':''}" ${active?'aria-current="true"':'data-editor="scene-switch"'} data-id="${e(id)}"><span class="editor-scene-image" ${image?`style="background-image:url('${e(image)}')"`:''}></span><span>${e(name)}<small>${active?'Players are here':'Click to bring everyone here'}</small></span>${active?'<b>LIVE</b>':''}</button>`;
    document.querySelector('.table-map')?.insertAdjacentHTML('beforeend',`<section class="editor-scenes" aria-label="Scene switcher"><header><div><span class="editor-kicker">YOUR CAMPAIGN</span><h2>Scenes</h2></div>${btn('Prepare scenes','prepare-scenes')}${btn('×','close','aria-label="Close scene switcher"')}</header><div class="editor-scene-list">${page(w.scene_id,w.scene_name||'Riverside village',w.image_data||(w.map_image==='blank'?'':'/assets/riverside-battlemap.png'),true)}${(w.pages||[]).map(p=>page(p.id,p.name,p.state?.image_data||(p.state?.map_image==='blank'?'':'/assets/riverside-battlemap.png'),false)).join('')}<button type="button" class="editor-scene-new" data-editor="prepare-scenes"><span>＋</span>New scene</button></div></section>`);
  }
  function configuration(a,changes){return{type:'token.configure',token_id:a.id,layer:a.layer||'tokens',rotation:a.rotation||0,size:a.size||1,initiative_bonus:a.initiative_bonus||0,locked:Boolean(a.locked),markers:a.markers||[],...changes};}
  function tokenHud(){
    document.querySelector('.editor-token-hud')?.remove();if(!state.editorToken)return;
    const a=world().actors?.find(a=>a.id===state.editorToken),map=document.querySelector('.table-map'),token=document.querySelector(`.map-token[data-actor-id="${CSS.escape(state.editorToken)}"]`);if(!a||!map||!token){state.editorToken='';return;}
    const b=map.getBoundingClientRect(),t=token.getBoundingClientRect(),left=Math.max(8,Math.min(b.width-278,t.x+t.width/2-b.x-135)),top=Math.max(8,Math.min(b.height-370,t.bottom-b.y+12));
    const gm=state.isGM,editable=gm&&!state.playerPreview,markers=['poisoned','prone','stunned','blessed','concentrating','invisible'];
    map.insertAdjacentHTML('beforeend',`<section class="editor-token-hud" aria-label="Token actions for ${e(a.name)}" style="left:${left}px;top:${top}px"><header><span class="editor-hud-avatar ${a.hostile?'hostile':''}">${e(a.name.slice(0,2))}</span><div><h2>${e(a.name)}</h2><span class="editor-muted">${a.hostile?'Hostile NPC':a.controller_name||'Game master control'}</span></div>${btn('×','token-close','aria-label="Close token actions"')}</header><div class="editor-hud-health"><span>Hit points</span><strong>${a.hp} <small>/ ${a.max_hp}</small></strong><i><b style="width:${Math.max(0,Math.min(100,a.hp/a.max_hp*100))}%"></b></i></div>${editable?`<form data-editor-form="hp"><input aria-label="Token hit points or adjustment" name="hp" placeholder="${a.hp} or −5 / +5" required maxlength="8" inputmode="text"><button>Apply HP</button></form><div class="editor-token-controls">${btn('↻ Rotate','rotate')}${btn(a.locked?'Unlock':'Lock movement','lock')}${btn('＋ Turn','add-turn',a.hostile?'disabled':'')}</div><div class="editor-marker-picker" role="group" aria-label="Status markers">${markers.map(m=>btn(m,'marker',`data-marker="${m}" aria-pressed="${(a.markers||[]).includes(m)}"`)).join('')}</div><label class="editor-layer-label">Token layer<select aria-label="Token layer shortcut" data-editor-token-layer>${[['tokens','Tokens'],['gm','GM private'],['map','Map'],['foreground','Foreground']].map(([id,label])=>`<option value="${id}" ${id===(a.layer||'tokens')?'selected':''}>${label}</option>`).join('')}</select></label>`:''}<footer>${btn('Open inspector','inspect-token')}${controlsActor(a)&&a.id===world().active_actor_id?btn('End turn →','token-end'):''}</footer></section>`);
  }
  function afterRender(){rail();sidebar();popover();scenes();tokenHud();document.querySelectorAll('.map-world .map-token').forEach(token=>{const a=world().actors?.find(a=>a.id===token.dataset.actorId);if(!a||token.querySelector('.editor-token-health'))return;token.insertAdjacentHTML('beforeend',`<span class="editor-token-health" aria-hidden="true"><i style="width:${Math.max(0,Math.min(100,a.hp/a.max_hp*100))}%"></i></span>`);});if(state.isGM&&world().initiative?.length)document.querySelectorAll('.canvas-turns li').forEach(row=>{row.draggable=true;row.dataset.editorTurn=row.querySelector('[data-canvas-score]')?.dataset.canvasScore||'';row.title='Drag to reorder; arrow buttons also work';});}
  function open(panel){close();state.editorPanel=panel;state.canvasDrawMenu=false;document.querySelector('.canvas-draw-menu')?.remove();popover();scenes();syncRail();}
  function init(){
    document.addEventListener('click',async event=>{
      const control=event.target.closest('[data-editor]');if(!control)return;event.preventDefault();event.stopImmediatePropagation();if(state.commandPending)return;
      const action=control.dataset.editor,a=world().actors?.find(a=>a.id===state.editorToken);
      if(['dice','layers','scenes'].includes(action)){const was=state.editorPanel===action;close();if(!was)open(action);return;}
      if(action==='close'){close();return;}
      if(action==='more'){const next=!state.editorMore;close();state.editorMore=next;const menu=document.querySelector('.editor-secondary-tools');if(menu)menu.hidden=!next;syncRail();return;}
      if(action==='turns'){document.querySelector('[data-canvas="tracker"]')?.click();return;}
      if(action==='layer'){state.vttDrafts??={};state.vttDrafts.brush={...state.vttDrafts.brush,layer:control.dataset.layer};popover();const select=document.querySelector('[data-canvas-brush="layer"]');if(select)select.value=control.dataset.layer;return;}
      if(action==='scene-switch'){await command({type:'scene.switch',scene_id:control.dataset.id});return;}
      if(action==='prepare-scenes'){close();state.vttSection='scenes';selectInspectorTab('table');renderTable(state.selectedCampaign);document.querySelector('.table-sidebar')?.classList.add('is-open');return;}
      if(action==='roll'){const form=document.querySelector('[data-editor-form="dice"]');if(!form?.reportValidity())return;state.quickDiceCount=Number(form.elements.count.value);state.quickDiceBonus=Number(form.elements.bonus.value);await command({type:'dice.roll',count:state.quickDiceCount,sides:Number(control.dataset.sides),bonus:state.quickDiceBonus});return;}
      if(action==='token-close'){state.editorToken='';tokenHud();return;}
      if(!a)return;
      if(action==='inspect-token'){selectInspectorTab('actions');document.querySelector('.table-sidebar')?.classList.add('is-open');state.editorToken='';tokenHud();return;}
      if(action==='token-end'){await command({type:'turn.end',character_id:a.id});return;}
      if(!state.isGM||state.playerPreview)return;
      if(action==='rotate')await command(configuration(a,{rotation:((a.rotation||0)+45)%360}));
      if(action==='lock')await command(configuration(a,{locked:!a.locked}));
      if(action==='marker'){const markers=new Set(a.markers||[]);markers.has(control.dataset.marker)?markers.delete(control.dataset.marker):markers.add(control.dataset.marker);await command(configuration(a,{markers:[...markers]}));}
      if(action==='add-turn')await command({type:'initiative.manage',action:'score',token_id:a.id,score:a.initiative_bonus||0});
    },true);
    document.addEventListener('click',event=>{
      if(event.target.closest('[data-map-tool]'))requestAnimationFrame(syncRail);
      const token=event.target.closest('.map-world .map-token');if(token){const now=performance.now();if(lastTap.id===token.dataset.actorId&&now-lastTap.at<350){close();state.editorToken=token.dataset.actorId;renderTable(state.selectedCampaign);lastTap={id:'',at:0};return;}lastTap={id:token.dataset.actorId,at:now};}
      if(event.isTrusted&&event.target.closest('[data-editor-draw]')){close();if(!state.canvasDrawMenu)document.querySelector('[data-canvas="draw-menu"]')?.click();}
      if(event.target.closest('.map-world')&&!event.target.closest('.map-token'))close();
      const target=event.target.closest('.editor-secondary-tools [data-map-tool]');if(target){state.editorMore=false;document.querySelector('.editor-secondary-tools').hidden=true;syncRail();}
    });
    const showToken=event=>{const token=event.target.closest('.map-world .map-token');if(!token)return;event.preventDefault();event.stopImmediatePropagation();close();const a=world().actors?.find(a=>a.id===token.dataset.actorId);if(!a)return;state.editorToken=a.id;if(a.hostile)state.selectedTarget=a.id;else{state.selectedActor=a.id;state.selectedTarget=null;}renderTable(state.selectedCampaign);};
    document.addEventListener('contextmenu',showToken,true);document.addEventListener('dblclick',showToken,true);
    document.addEventListener('submit',async event=>{const form=event.target.closest('[data-editor-form]');if(!form)return;event.preventDefault();event.stopImmediatePropagation();if(state.commandPending||!form.reportValidity())return;
      if(form.dataset.editorForm==='dice'){state.quickDiceCount=Number(form.elements.count.value);state.quickDiceBonus=Number(form.elements.bonus.value);await command({type:'dice.roll',count:state.quickDiceCount,sides:20,bonus:state.quickDiceBonus});return;}
      const a=world().actors?.find(a=>a.id===state.editorToken);if(!a||!state.isGM||state.playerPreview)return;const input=form.elements.hp.value.trim().replace('−','-');if(!/^[+-]?\d+$/.test(input)){notify('Enter hit points or an adjustment such as -5 or +5.');return;}const hp=/^[+-]/.test(input)?a.hp+Number(input):Number(input);if(hp<0||hp>a.max_hp){notify(`Hit points must be between 0 and ${a.max_hp}.`);return;}await command({type:'token.update',token_id:a.id,name:a.name,hp});
    },true);
    document.addEventListener('change',async event=>{if(!event.target.matches('[data-editor-token-layer]')||state.commandPending)return;const a=world().actors?.find(a=>a.id===state.editorToken);if(a&&state.isGM&&!state.playerPreview)await command(configuration(a,{layer:event.target.value}));});
    document.addEventListener('keydown',event=>{if(event.key==='Escape')close();});
    document.addEventListener('dragstart',event=>{const row=event.target.closest('[data-editor-turn]');if(!row||event.target.closest('input,button')||!state.isGM)return;dragTurn=row.dataset.editorTurn;event.dataTransfer.setData('text/plain',dragTurn);event.dataTransfer.effectAllowed='move';row.classList.add('is-reordering');});
    document.addEventListener('dragover',event=>{const row=event.target.closest('[data-editor-turn]');if(row&&dragTurn&&row.dataset.editorTurn!==dragTurn){event.preventDefault();event.dataTransfer.dropEffect='move';document.querySelectorAll('.editor-drop-target').forEach(el=>el.classList.remove('editor-drop-target'));row.classList.add('editor-drop-target');}});
    document.addEventListener('drop',async event=>{const row=event.target.closest('[data-editor-turn]');if(!row||!dragTurn||!state.isGM)return;event.preventDefault();const token=dragTurn;dragTurn='';document.querySelectorAll('.editor-drop-target,.is-reordering').forEach(el=>el.classList.remove('editor-drop-target','is-reordering'));if(token!==row.dataset.editorTurn)await command({type:'initiative.manage',action:'before',token_id:token,before_id:row.dataset.editorTurn});});
    document.addEventListener('dragend',()=>{dragTurn='';document.querySelectorAll('.editor-drop-target,.is-reordering').forEach(el=>el.classList.remove('editor-drop-target','is-reordering'));});
    window.addEventListener('hashchange',close);
  }
  return{afterRender,init};
};
