/* Map interaction layer. Previews are local; committed changes use commands. */
window.AstralisSmoothPath=function(points){
  if(!points?.length)return '';
  const p=points.map(p=>({x:Math.round(p.x*400)/10,y:Math.round(p.y*400)/10}));
  let path=`M${p[0].x} ${p[0].y}`;
  for(let i=1;i<p.length-1;i++)path+=` Q${p[i].x} ${p[i].y} ${(p[i].x+p[i+1].x)/2} ${(p[i].y+p[i+1].y)/2}`;
  return path+` L${p.at(-1).x} ${p.at(-1).y}`;
};
window.AstralisCanvas=function(ctx){
  const {state,command,renderTable,escapeHTML:e,controlsActor,currentRoute,notify}=ctx;
  const editor=window.AstralisEditor?.(ctx);
  const systems=window.AstralisSystems?.(ctx);
  const w=()=>state.snapshot?.state||{},shapes=[['freehand','✎','Freehand'],['line','╱','Line'],['rectangle','▭','Rectangle'],['ellipse','◯','Ellipse'],['cone','◁','Cone'],['text','T','Text'],['pin','⌖','Pin']];
  let drag=null,space=false,suppressUntil=0,suppressPoint=null,frame=0,lastPoint=null,panelScroll=0,pendingToken=null;
  let previousTokens=new Map(),previousWidth=0;
  const isTable=()=>currentRoute().startsWith('/table/');
  const isInput=target=>target?.closest('input,textarea,select,button,[contenteditable=true]');
  const brush=()=>({layer:state.vttDrafts?.brush?.layer||'tokens',color:state.vttDrafts?.brush?.color||'#e7c98c',stroke:Number(state.vttDrafts?.brush?.stroke||2)});
  function point(event,exact=false){const board=document.querySelector('.map-world');if(!board)return null;const b=board.getBoundingClientRect(),world=w(),x=Math.max(0,Math.min(world.width-.001,(event.clientX-b.left)/b.width*world.width)),y=Math.max(0,Math.min(world.height-.001,(event.clientY-b.top)/b.height*world.height));return exact?{x:Math.round(x*1000)/1000,y:Math.round(y*1000)/1000}:{x:Math.floor(x),y:Math.floor(y)};}
  function mode(){return state.mapMode||'select';}
  function shape(){const m=mode();return m==='line'?'line':m==='rectangle'?'rectangle':m.startsWith('vtt-')?m.slice(4):null;}
  function setMode(next){document.querySelector(`[data-map-tool="${next}"]`)?.click();state.mapMode=next;state.vttStart=null;state.measureStart=null;document.querySelectorAll('[data-map-tool]').forEach(el=>{const active=el.dataset.mapTool===next;el.classList.toggle('is-active',active);el.setAttribute('aria-pressed',String(active));});document.querySelector('.map-viewport')?.classList.toggle('pan-mode',next==='pan');floating();}
  function toolOptions(){
    const drawing=shapes.some(([id])=>id===shape()),b=brush();
    return `<button type="button" class="canvas-icon" data-canvas="draw-menu" aria-label="Drawing tools" aria-expanded="${Boolean(state.canvasDrawMenu)}">✎<span>Draw</span><small>⌄</small></button><span class="canvas-divider"></span>${drawing?`<label title="Stroke color"><span class="sr-only">Stroke color</span><input type="color" aria-label="Quick stroke color" value="${e(b.color)}" data-canvas-brush="color"></label><label class="canvas-width"><span class="sr-only">Stroke width</span><select aria-label="Quick stroke width" data-canvas-brush="stroke">${[1,2,3,4,6,8,12].map(n=>`<option value="${n}" ${b.stroke===n?'selected':''}>${n} px</option>`).join('')}</select></label><label><span class="sr-only">Drawing layer</span><select aria-label="Quick drawing layer" data-canvas-brush="layer">${[['tokens','Tokens'],['map','Map'],['gm','GM private'],['foreground','Foreground']].map(([id,label])=>`<option value="${id}" ${b.layer===id?'selected':''}>${label}</option>`).join('')}</select></label><button type="button" class="canvas-icon" data-canvas="undo" aria-label="Undo drawing on current layer">↶</button>`:`<span class="canvas-tool-hint">${mode()==='pan'?'Drag to pan':mode()==='measure'?'Drag to measure':'Drag tokens · Space to pan'}</span>`}`;
  }
  function floating(){
    const map=document.querySelector('.table-map');if(!map)return;
    map.querySelector('.canvas-tools')?.remove();map.querySelector('.canvas-draw-menu')?.remove();
    if(state.isGM)map.insertAdjacentHTML('beforeend',`<div class="canvas-tools" role="toolbar" aria-label="Drawing appearance">${toolOptions()}</div>${state.canvasDrawMenu?`<div class="canvas-draw-menu" role="toolbar" aria-label="Drawing shapes">${shapes.map(([id,icon,label])=>`<button type="button" class="canvas-shape ${shape()===id?'is-active':''}" data-canvas-shape="${id}" aria-label="Draw ${label.toLowerCase()}" aria-pressed="${shape()===id}"><span aria-hidden="true">${icon}</span>${label}</button>`).join('')}</div>`:''}`);
    map.querySelector('.canvas-view-actions')?.remove();map.insertAdjacentHTML('beforeend',`<div class="canvas-view-actions"><button type="button" class="canvas-icon" data-canvas="tracker" aria-label="Toggle turn tracker" aria-pressed="${Boolean(state.localTracker||w().tracker_visible)}">☷<span>Turns</span></button><button type="button" class="canvas-icon" data-canvas="sidebar" aria-label="${state.sidebarCollapsed?'Show':'Hide'} inspector" aria-expanded="${!state.sidebarCollapsed}">${state.sidebarCollapsed?'◂':'▸'}</button></div>`);
  }
  function tracker(){
    document.querySelector('.canvas-turns')?.remove();if(state.trackerDismissed||!state.localTracker&&!w().tracker_visible)return;
    const world=w(),gm=state.isGM,rows=world.initiative?.length?world.initiative:(world.actors||[]).filter(a=>!a.hostile&&a.hp>0&&a.layer!=='map'&&a.layer!=='gm').map(a=>({token_id:a.id,score:null}));
    document.querySelector('.table-map')?.insertAdjacentHTML('beforeend',`<section class="canvas-turns" aria-label="Turn tracker"><header><div><p class="caption-4 muted">ROUND ${world.round||1}</p><h2 class="title-4">Turn order</h2></div><button type="button" class="canvas-icon" data-canvas="tracker-close" aria-label="Close turn tracker">×</button></header><ol>${rows.map((row,index)=>{const actor=world.actors?.find(a=>a.id===row.token_id);return actor?`<li class="${actor.id===world.active_actor_id?'is-current':''}"><span class="canvas-turn-avatar">${e(actor.name.slice(0,2))}</span><button type="button" class="canvas-turn-name" data-canvas-token="${e(actor.id)}">${e(actor.name)}${actor.id===world.active_actor_id?'<small>Current turn</small>':''}</button>${gm?`<input type="number" aria-label="Initiative for ${e(actor.name)}" min="-100" max="100" value="${row.score??''}" placeholder="—" data-canvas-score="${e(actor.id)}"><button type="button" class="canvas-mini" data-canvas-order="up" data-token="${e(actor.id)}" aria-label="Move ${e(actor.name)} earlier" ${index===0||row.score===null?'disabled':''}>↑</button><button type="button" class="canvas-mini" data-canvas-order="down" data-token="${e(actor.id)}" aria-label="Move ${e(actor.name)} later" ${index===rows.length-1||row.score===null?'disabled':''}>↓</button>`:`<span>${row.score??'—'}</span>`}</li>`:'';}).join('')}</ol><footer>${gm?`<button type="button" data-canvas="roll-initiative">Roll</button><button type="button" data-canvas="sort-initiative" ${world.initiative?.length?'':'disabled'}>Sort</button><button type="button" data-canvas="start-initiative" ${world.initiative?.length?'':'disabled'}>Start</button>`:''}<button type="button" data-canvas="next-turn" ${controlsActor(world.actors?.find(a=>a.id===world.active_actor_id))?'':'disabled'}>Next turn →</button></footer><p class="caption-4 muted">${gm?'Party turns · drag order with ↑ ↓':'Party turn order'}</p></section>`);
  }
  function beforeRender(){const panel=document.querySelector('.inspector-panel:not([hidden])');panelScroll=panel?.scrollTop||0;previousTokens=new Map([...document.querySelectorAll('.map-world .map-token')].map(el=>[el.dataset.actorId,el.getBoundingClientRect()]));previousWidth=document.querySelector('.map-world')?.getBoundingClientRect().width||0;}
  function prepare(){document.querySelector('.vtt-workspace')?.classList.toggle('inspector-collapsed',Boolean(state.sidebarCollapsed));}
  function afterRender(){prepare();floating();tracker();const panel=document.querySelector('.inspector-panel:not([hidden])');if(panel)panel.scrollTop=panelScroll;if(pendingToken&&state.commandPending){const el=document.querySelector(`.map-token[data-actor-id="${CSS.escape(pendingToken.id)}"]`);if(el){el.style.left=`${(pendingToken.x+pendingToken.size/2)/w().width*100}%`;el.style.top=`${(pendingToken.y+pendingToken.size/2)/w().height*100}%`;el.classList.add('is-dragging');}}else pendingToken=null;}
  function animateTokens(){
    const width=document.querySelector('.map-world')?.getBoundingClientRect().width||0;
    if(state.commandPending||state.mapInteracting||matchMedia('(prefers-reduced-motion: reduce)').matches||Math.abs(width-previousWidth)>1)return;
    document.querySelectorAll('.map-world .map-token').forEach(el=>{const old=previousTokens.get(el.dataset.actorId);if(!old)return;const next=el.getBoundingClientRect(),dx=old.left-next.left,dy=old.top-next.top;if(Math.hypot(dx,dy)>1)el.animate([{translate:`${dx}px ${dy}px`},{translate:'0px 0px'}],{duration:180,easing:'cubic-bezier(.2,.7,.2,1)'});});
  }
  function preview(){
    frame=0;const board=document.querySelector('.map-world');if(!board||!drag||!lastPoint)return;
    board.querySelector('.canvas-gesture')?.remove();const world=w(),b=brush(),p=drag.start,q=lastPoint;
    let content='';
    if(drag.kind==='freehand')content=`<path d="${window.AstralisSmoothPath(drag.path)}" fill="none" stroke="${e(b.color)}" stroke-width="${b.stroke}" stroke-linecap="round" stroke-linejoin="round"/>`;
    if(drag.kind==='shape'){const x=p.x*40+20,y=p.y*40+20,tx=q.x*40+20,ty=q.y*40+20,attrs=`fill="${e(b.color)}" fill-opacity=".06" stroke="${e(b.color)}" stroke-width="${b.stroke}"`;if(drag.shape==='line')content=`<line x1="${x}" y1="${y}" x2="${tx}" y2="${ty}" ${attrs}/>`;else if(drag.shape==='ellipse')content=`<ellipse cx="${(x+tx)/2}" cy="${(y+ty)/2}" rx="${Math.max(1,Math.abs(tx-x)/2)}" ry="${Math.max(1,Math.abs(ty-y)/2)}" ${attrs}/>`;else if(drag.shape==='cone')content=`<polygon points="${x},${y} ${tx-(ty-y)/2},${ty+(tx-x)/2} ${tx+(ty-y)/2},${ty-(tx-x)/2}" ${attrs}/>`;else content=`<rect x="${Math.min(x,tx)-20}" y="${Math.min(y,ty)-20}" width="${Math.abs(tx-x)+40}" height="${Math.abs(ty-y)+40}" ${attrs}/>`;}
    if(drag.kind==='measure'){const dx=Math.abs(q.x-p.x),dy=Math.abs(q.y-p.y),distance=(world.metric==='euclidean'?Math.hypot(dx,dy):world.metric==='chebyshev'?Math.max(dx,dy):dx+dy)*(world.scale||1);content=`<line x1="${p.x*40+20}" y1="${p.y*40+20}" x2="${q.x*40+20}" y2="${q.y*40+20}" stroke="#8bd0d9" stroke-width="2" stroke-dasharray="6 4"/><text x="${q.x*40+20}" y="${q.y*40+10}" text-anchor="middle" fill="#fff" stroke="#181818" stroke-width="3" paint-order="stroke" font-size="14">${Math.round(distance*100)/100} ${e(world.unit||'spaces')}</text>`;}
    if(drag.kind==='token'){const token=board.querySelector(`.map-token[data-actor-id="${CSS.escape(drag.token)}"]`);if(token){token.style.left=`${(q.x+drag.size/2)/world.width*100}%`;token.style.top=`${(q.y+drag.size/2)/world.height*100}%`;token.classList.add('is-dragging');}content=`<rect x="${q.x*40+2}" y="${q.y*40+2}" width="${drag.size*40-4}" height="${drag.size*40-4}" rx="4" fill="#8bd0d9" fill-opacity=".12" stroke="#8bd0d9" stroke-dasharray="4 3"/>`;}
    board.insertAdjacentHTML('beforeend',`<svg class="canvas-gesture" viewBox="0 0 ${world.width*40} ${world.height*40}" aria-hidden="true">${content}</svg>`);
  }
  function queuePreview(){if(!frame)frame=requestAnimationFrame(preview);}
  function cancel(){if(frame)cancelAnimationFrame(frame);frame=0;document.querySelector('.canvas-gesture')?.remove();if(drag?.kind==='token'){const a=w().actors?.find(a=>a.id===drag.token),el=document.querySelector(`.map-token[data-actor-id="${CSS.escape(drag.token)}"]`);if(a&&el){el.style.left=`${(a.x+drag.size/2)/w().width*100}%`;el.style.top=`${(a.y+drag.size/2)/w().height*100}%`;el.classList.remove('is-dragging');}}drag=null;state.mapInteracting=false;}
  function init(){
    document.addEventListener('click',event=>{if(event.target.closest('[data-map-tool],[data-vtt-shape]'))requestAnimationFrame(floating);});
    document.addEventListener('click',async event=>{
      if(event.target.closest('.map-viewport')&&Date.now()<suppressUntil){const same=suppressPoint&&Math.hypot(event.clientX-suppressPoint.x,event.clientY-suppressPoint.y)<4;suppressUntil=0;if(same){event.preventDefault();event.stopImmediatePropagation();return;}}
      const control=event.target.closest('[data-canvas]'),action=control?.dataset.canvas,draw=event.target.closest('[data-canvas-shape]');
      if(draw){event.stopImmediatePropagation();state.canvasDrawMenu=false;setMode('vtt-'+draw.dataset.canvasShape);return;}
      if(!control)return;event.stopImmediatePropagation();
      if(action==='draw-menu'){state.canvasDrawMenu=!state.canvasDrawMenu;floating();return;}
      if(action==='sidebar'){beforeRender();state.sidebarCollapsed=!state.sidebarCollapsed;renderTable(state.selectedCampaign);return;}
      if(action==='tracker'||action==='tracker-close'){if(state.isGM&&!state.playerPreview){await command({type:'initiative.manage',action:'visibility',visible:action==='tracker'?!w().tracker_visible:false});}else{state.trackerDismissed=action==='tracker-close';state.localTracker=action==='tracker';tracker();}return;}
      if(action==='undo'){await command({type:'draw.undo',layer:brush().layer});return;}
      if(action==='roll-initiative')await command({type:'initiative.roll'});
      if(action==='sort-initiative')await command({type:'initiative.manage',action:'sort'});
      if(action==='start-initiative')await command({type:'initiative.start'});
      if(action==='next-turn')await command({type:'turn.end',character_id:w().active_actor_id});
    },true);
    document.addEventListener('change',async event=>{
      const key=event.target.dataset.canvasBrush;if(key){state.vttDrafts??={};state.vttDrafts.brush??={};state.vttDrafts.brush[key]=key==='stroke'?Number(event.target.value):event.target.value;return;}
      if(event.target.dataset.canvasScore){if(!event.target.reportValidity()||event.target.value==='')return;await command({type:'initiative.manage',action:'score',token_id:event.target.dataset.canvasScore,score:Number(event.target.value)});}
    });
    document.addEventListener('click',async event=>{const button=event.target.closest('[data-canvas-order]');if(button)await command({type:'initiative.manage',action:button.dataset.canvasOrder,token_id:button.dataset.token});const token=event.target.closest('[data-canvas-token]');if(token)document.querySelector(`.map-token[data-actor-id="${CSS.escape(token.dataset.canvasToken)}"]`)?.click();});
    document.addEventListener('keydown',event=>{if(!isTable()||isInput(event.target))return;if(event.code==='Space'){event.preventDefault();space=true;document.querySelector('.map-viewport')?.classList.add('space-pan');}if(event.key==='Escape'){cancel();state.canvasDrawMenu=false;floating();}if(event.key.toLowerCase()==='y'){event.preventDefault();document.querySelector('[data-canvas="tracker"]')?.click();}if(event.key.toLowerCase()==='w'){event.preventDefault();document.querySelector('[data-canvas="sidebar"]')?.click();}});
    document.addEventListener('keyup',event=>{if(event.code==='Space'){space=false;document.querySelector('.map-viewport')?.classList.remove('space-pan');}});
    window.addEventListener('blur',()=>{space=false;cancel();document.querySelector('.space-pan')?.classList.remove('space-pan');});
    document.addEventListener('wheel',event=>{if(!event.target.closest('.map-viewport')||!event.ctrlKey&&!event.metaKey)return;event.preventDefault();ctx.zoom(Math.max(.25,Math.min(2.5,(state.mapZoom||1)*Math.exp(-event.deltaY*.002))),{x:event.clientX,y:event.clientY});},{passive:false});
    document.addEventListener('pointerdown',event=>{
      const viewport=event.target.closest('.map-viewport');if(!viewport||drag||state.commandPending||event.button!==0&&event.button!==1)return;
      const board=event.target.closest('.map-world');if(!board)return;
      const pan=space||event.button===1||mode()==='pan',s=shape(),p=point(event),token=event.target.closest('.map-token'),a=token?w().actors?.find(a=>a.id===token.dataset.actorId):null;
      let kind=pan?'pan':mode()==='measure'?'measure':state.isGM&&s==='freehand'?'freehand':state.isGM&&['line','rectangle','ellipse','cone'].includes(s)?'shape':mode()==='select'&&a&&controlsActor(a)&&!state.playerPreview&&(state.isGM||!a.locked&&!w().movement_paused&&a.id===w().active_actor_id)?'token':null;
      if(!kind)return;event.preventDefault();event.stopImmediatePropagation();viewport.focus({preventScroll:true});state.mapInteracting=true;
      drag={kind,shape:s,start:p,x:event.clientX,y:event.clientY,id:event.pointerId,viewport,board,left:viewport.scrollLeft,top:viewport.scrollTop,path:[point(event,true)],moved:false,token:a?.id,size:a?.size||1};lastPoint=p;viewport.setPointerCapture(event.pointerId);
    },true);
    document.addEventListener('pointermove',event=>{if(!drag||drag.id!==event.pointerId)return;event.stopImmediatePropagation();drag.moved ||= Math.hypot(event.clientX-drag.x,event.clientY-drag.y)>4;if(drag.kind==='pan'){drag.viewport.scrollLeft=drag.left+drag.x-event.clientX;drag.viewport.scrollTop=drag.top+drag.y-event.clientY;state.mapScroll={x:drag.viewport.scrollLeft,y:drag.viewport.scrollTop};return;}lastPoint=point(event);if(drag.kind==='freehand'){for(const sample of event.getCoalescedEvents?.()||[event]){const p=point(sample,true),last=drag.path.at(-1);if(p&&Math.hypot(p.x-last.x,p.y-last.y)>.025&&drag.path.length<1000)drag.path.push(p);}}if(drag.moved)queuePreview();},true);
    document.addEventListener('pointerup',async event=>{
      if(!drag||drag.id!==event.pointerId)return;event.stopImmediatePropagation();const g=drag,q=point(event);if(g.viewport.hasPointerCapture(event.pointerId))g.viewport.releasePointerCapture(event.pointerId);cancel();if(!g.moved){if(g.token&&(g.kind==='token'||g.kind==='pan'))g.board.querySelector(`.map-token[data-actor-id="${CSS.escape(g.token)}"]`)?.click();else if(g.kind==='shape'||g.kind==='measure')g.board.dispatchEvent(new MouseEvent('click',{bubbles:true,clientX:event.clientX,clientY:event.clientY}));return;}suppressUntil=Date.now()+400;suppressPoint={x:event.clientX,y:event.clientY};state.measureStart=null;state.vttStart=null;
      if(g.kind==='pan')return;
      if(g.kind==='token'){pendingToken={id:g.token,x:q.x,y:q.y,size:g.size};await command(state.isGM?{type:'token.place',token_id:g.token,...q}:{type:'character.move',character_id:g.token,target:q});return;}
      if(g.kind==='measure'){state.measureResult={a:g.start,b:q};renderTable(state.selectedCampaign);return;}
      const b=brush();if(g.kind==='shape'){await command({type:'draw.add',...b,shape:g.shape,x:g.start.x,y:g.start.y,to_x:q.x,to_y:q.y});return;}
      if(g.path.length>=2){const points=g.path.filter((p,i,list)=>i===0||Math.floor(p.x)!==Math.floor(list[i-1].x)||Math.floor(p.y)!==Math.floor(list[i-1].y)).map(p=>({x:Math.floor(p.x),y:Math.floor(p.y)})).slice(0,250);if(points.length===1)points.push({...points[0]});await command({type:'draw.add',...b,shape:'freehand',x:g.start.x,y:g.start.y,to_x:q.x,to_y:q.y,points,path:g.path});}
    },true);
    document.addEventListener('pointercancel',event=>{if(drag?.id===event.pointerId){event.stopImmediatePropagation();cancel();}},true);
    window.addEventListener('hashchange',()=>{cancel();space=false;state.canvasDrawMenu=false;});
  }
  return {prepare,afterRender:()=>{afterRender();animateTokens();editor?.afterRender();systems?.afterRender();},beforeRender,init:()=>{systems?.init();init();editor?.init();}};
};
