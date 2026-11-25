/* One handbook for the public help centre and contextual tabletop help. */
window.AstralisLearning = function(ctx) {
  const {state,escapeHTML:e,nav,footer,currentRoute,renderTable,selectInspectorTab,notify}=ctx;
  const guides=[];
  function topic(id,title,group,role,summary,steps,tip,target){guides.push({id,title,group,role,summary,steps,tip,target});}
  topic('start','Your first session','Start here','all','Choose a role, open a campaign, and learn the table at your own pace.',[
    'For a quick practice game, open Sign in and choose Try a demo table. Each visit creates a separate saved GM sandbox. Keep that browser signed in to return to it.',
    'For your own campaign, register or sign in, open My campaigns → New campaign, enter its details and open the lobby. Launch session creates the starter encounter; Resume returns to a saved one.',
    'Joining a friend? Register your own account, enter their invitation code in My campaigns, then open the campaign and resume the shared session.',
    'Open Help & tools inside the table for a guided tour or a preparation checklist. Explore panels first; saved changes happen when you submit their controls.'
  ],'The public homepage map is a preview. Use a demo or your own campaign to save your play.');
  topic('campaigns','Campaigns & invitations','Start here','gm','Create a table, invite friends, and return to your saved adventure.',[
    'Create a campaign from My campaigns. Search by name or use Sort by to find it later. Open its lobby to launch or resume a session.',
    'At the table, open Tools → Invitations. Choose expiry and join limits, create a code, and share that code with your players.',
    'Players redeem the code from their own campaign library. Refresh players in Characters after they join, then assign their tokens.',
    'List or revoke invitations in Tools. Revoking a code prevents future joins; existing campaign members stay in the campaign.'
  ],'Demo accounts have no password recovery. Use a registered account for a continuing campaign.',{tab:'tools'});
  topic('navigation','Find your way around','Start here','all','Keep the map visible while finding the controls you need.',[
    'The left rail contains map tools. More reveals terrain, fog planning, placement and other secondary controls.',
    'Use Actions for movement and combat, Characters for the roster, Tools for imports and invitations, Table for the full workspace, and Log for saved results.',
    'In Table, choose a section from Workspace section. Help & tools can also find a feature and open its panel directly.',
    'On small screens, use the inspector toggle or panel button to reveal controls. Close the inspector when you want more map space.'
  ],'Press ? for help, W to toggle the inspector, or Ctrl/Command + K to open the tool finder. These shortcuts leave normal text entry alone.');
  topic('map','Select, pan, zoom & grid','Map & editor','all','Explore the map without moving any characters.',[
    'Choose Select (1) and click a token to inspect it. Choose a hostile token to set a combat target.',
    'Choose Pan (2) and drag the viewport. Space + drag or middle-button drag temporarily pans without changing your tool.',
    'Use + / − or Ctrl/Command + wheel to zoom. Fit brings the scene into view; G toggles your local grid.',
    'Grid, zoom and panning affect your own view. Sync refreshes the saved encounter; it is not a save button.'
  ],'Escape cancels an unfinished gesture. On touch screens, use the rail and zoom buttons.',{tool:'select'});
  topic('actions','Movement, attacks & turns','Play together','all','Take an action with a character you are allowed to control.',[
    'Select a living party token. Players need a GM assignment; the GM can control every party character.',
    'On its active turn, drag the token or use the directional buttons in Actions. A traversable path, free destination and remaining movement are required. GM dragging uses placement instead.',
    'Choose a hostile target in range and press Attack target. The saved result records the roll, hit or miss and damage. The basic encounter allows one attack per turn.',
    'For the starter bridge interaction, move within two spaces of column 8, row 5 and choose Investigate bridge. End turn passes control to the next living party member.'
  ],'Cannot move? Check assignment, active turn, HP, movement lock, pause, terrain, occupied cells and remaining movement. Game profiles do not replace these basic encounter rules.',{tab:'actions'});
  topic('characters','Assign player characters','Play together','gm','Give each player control of the right party token.',[
    'Invite the player and wait for them to join the campaign.',
    'Open Characters, select the party character, and use Refresh players if the new member is missing.',
    'Choose a member under Controlled by and save. Choose Game master only to remove an assignment.',
    'The player can act with that token on its turn. Sheet assignment is separate; set its controller in Table → Sheets as well.'
  ],'Hostile NPCs remain GM-controlled. Link a sheet explicitly from Table → Sheets. Existing token control stays independent.',{tab:'characters'});
  topic('tokens','Edit tokens & status markers','Map & editor','gm','Adjust token appearance and combat values without hunting through panels.',[
    'Right-click or double-click a token to open its contextual menu. On touch screens, select it and use Table → Tokens.',
    'Apply an absolute HP value or an adjustment such as -5 or +5. Rotate, lock movement, toggle markers, or change its layer.',
    'Use Table → Tokens for size, initiative modifier, duplication and deletion. Use Tools to add a new party or hostile token.',
    'For a portrait, import an image in Table → Art, choose the target token and apply the artwork.'
  ],'Active-turn tokens have deletion/layer restrictions. Markers are reminders; they do not automatically apply rule effects.',{section:'tokens'});
  topic('scenes','Prepare & switch scenes','Map & editor','gm','Move the adventure between locations while keeping their maps intact.',[
    'Open Scenes on the rail for thumbnails and the live-scene indicator. Prepare scenes opens Table → Scenes.',
    'Name a new scene. Enable Duplicate the current map and tokens to begin with a copy, or leave it off for a blank scene.',
    'Switch to an archived scene to bring the whole table there. Maps, tokens, drawings, environment and initiative belong to each scene.',
    'Rename the active scene or delete an archived one in Table → Scenes. Sheets, quests, journals and the chosen game profile carry across scenes.'
  ],'Scene switching affects every player. Tell the group before moving them.',{section:'scenes'});
  topic('imports','Import maps & crop images','Map & editor','gm','Bring your own PNG or JPEG map to the table.',[
    'Open Tools and choose a map file up to 5 MB and 20 megapixels. Upload it, or choose the Riverside/blank background.',
    'In Table → Settings, set grid dimensions, scene scale, units and distance metric. The grid supports 2–100 cells per side.',
    'Use the map setup controls in Table → Environment for percentage crop bounds and quarter-turn rotation of an imported custom image.',
    'Check the result with Fit and the ruler, then add blocked terrain, doors and lights for gameplay.'
  ],'Cropping replaces the image; retain your original file. Pixel-to-grid alignment and hex grids are not available.',{tab:'tools'});
  topic('art','Artwork library','Map & editor','gm','Reuse images as maps or token portraits.',[
    'Open Table → Art, name an image and import a PNG/JPEG up to 2 MB.',
    'Choose the target token before using Use on token, or choose Use as map for a scene background.',
    'The library holds up to ten images. Remove an unused original to free space; already placed copies remain.',
    'Keep private material on the GM layer. Unused library originals are visible only to the GM.'
  ],'Import artwork you have permission to use.',{section:'art'});
  topic('drawings','Draw, label & pin the map','Map & editor','gm','Sketch an area or mark a point of interest.',[
    'Choose Draw on the rail, then freehand, line, rectangle, ellipse, cone, text or pin. Adjust color, width and layer in the drawing controls.',
    'Drag to preview freehand or a shape and release to save it. Two-click shape entry also works. Escape cancels an unfinished drawing.',
    'For text or pins, enter Annotation text in Table → Drawings, choose Use drawing style, then choose the shape and click the map.',
    'Undo removes the latest drawing on the selected layer. Table → Drawings lists saved drawings for individual removal.'
  ],'Drawings are saved annotations. Use Ruler for temporary distance and area templates.',{section:'drawings'});
  topic('layers','Map, tokens, GM & foreground layers','Map & editor','gm','Decide where an object appears and who can see it.',[
    'Layer tools on the rail selects the layer for new drawings.',
    'Map places decoration below tokens; Tokens shares character-level drawings; Foreground appears above tokens.',
    'GM private keeps drawings and tokens out of player views. Change a token layer from its contextual menu or Table → Tokens.',
    'Undo applies to the drawing layer you selected. The layer picker does not filter all token selection to that layer.'
  ],'Check privacy before sharing a handout or moving an object off the GM layer.',{editor:'layers'});
  topic('terrain','Blocked terrain & fog planning','Map & editor','gm','Mark impassable cells and plan the encounter.',[
    'Open More → Block and click cells to toggle blocked terrain. These cells affect legal movement and token sight.',
    'Use More → Fog for shared planning marks. Click a marked cell again to remove it.',
    'Use Place to position the token selected in Tools, without spending ordinary turn movement.',
    'Choose Select when you finish editing to prevent accidental terrain changes.'
  ],'Fog planning is visible to players. To conceal information, use Lighting and assigned-token sight.',{tool:'block'});
  topic('ruler','Measure lines & area templates','Map & editor','all','Show distance without spending a character’s movement.',[
    'Choose Ruler and select line, square, circle, cone or beam. Choose cell-center snapping or free coordinates.',
    'Drag between points or click a start and end. Circle reports radius; cone reports depth; square reports the distance between opposite corners. Beam width is one cell.',
    'Enable share-on-release or choose Share to send the measurement. Clear removes your local overlay.',
    'Distances use the scene scale and metric. Templates and measurements do not apply damage or determine movement legality.'
  ],'With token sight enabled, GM measurements stay private. Player measurements require visible endpoints and go only to that player and the GM.',{tool:'measure'});
  topic('ping','Ping a location','Play together','all','Point the group toward a location on the map.',[
    'Choose Ping on the left rail.',
    'Click a cell to save a shared location marker.',
    'The latest ping appears on the map. Choose Select to return to token interaction.'
  ],'A ping draws attention; it does not move a token or change the current scene.',{tool:'ping'});
  topic('fx','Visual effects','Map & editor','gm','Add a burst, breath, beam or nova to a scene.',[
    'Open FX on the rail and choose an effect and element color.',
    'Click for a point effect, or drag to aim a breath or beam.',
    'Connected viewers see a temporary effect in its originating scene. Reduced-motion preferences simplify the animation.'
  ],'FX is visual only. Damage, spells and conditions stay manual. GM effects stay GM-only when token sight is enabled.',{system:'fx'});
  topic('turns','Initiative & turn tracker','Play together','gm','Order the party without granting extra actions.',[
    'Open Turns on the rail or press Y to show the tracker. Roll party initiative, then sort/start in that order.',
    'Edit initiative scores, drag an entry before another or use the earlier/later buttons. Reordering preserves movement and attack allowances.',
    'Next turn advances to the next living party token. The round advances when party order wraps.',
    'In Table → Turns, pause player movement during preparation and resume when ready.'
  ],'Initiative currently contains party tokens. NPC/custom entries and published-system initiative rules remain manual.',{section:'turns'});
  topic('dice','Public, GM & private dice','Play together','all','Choose your dice and audience before rolling.',[
    'Open Dice on the rail. Select count and bonus, then a die such as d6 or d20.',
    'Choose All players, You and GM, or Only you. Only you also hides the result from other accounts acting as GM.',
    'Inspect the latest result in the tray or find saved rolls in Chat and Log.',
    'Chat also accepts simple formulas: /roll 2d6+3 or /gmroll d20. Use Rules for a selected game’s check helper.'
  ],'Sheet rolls and running personal macros are public. Choose the dice tray when you need an explicit audience.',{editor:'dice'});
  topic('systems','Game profiles, checks & rulebooks','Play together','all','Use reference tools for your group’s game.',[
    'The GM chooses Game to play in the header: Astralis Basic, D&D 2014/2024, Warhammer 40,000, Pathfinder 2e, Fate or generic percentile.',
    'Open Rules for a quick reference, tailored checks, sheet templates and the book directory. Choose the check audience before rolling.',
    'D&D supports normal/advantage/disadvantage checks; Pathfinder reports four degrees; Warhammer counts qualifying d6; Fate uses 4dF; percentile compares d100 to a target.',
    'A profile preserves current tokens and scale. Apply scale is a separate action. Search Books for included D&D SRDs or linked publisher references.'
  ],'Profiles do not automate complete combat, spells or army phases. Templates fill manual sheets; sheet values do not update token combat stats.',{system:'rules'});
  topic('sheets','Sheets, inventory, spells & resources','Characters & story','all','Record a character and track limited uses.',[
    'Open Table → Sheets. The GM creates a sheet, assigns a controller and chooses whether it is shared. Assigned players can edit their own sheet.',
    'Save background, attributes, inventory, spells and resource notes. Open an existing sheet to see its tracked entries.',
    'Add an inventory, spell or resource entry with quantity and capacity. Prepare a spell before using it; Use one spends one quantity and prevents use when empty.',
    'Edit an entry to replenish it. Roll simple dice formulas from the sheet when needed; these rolls are public.'
  ],'Place a linked token or link an existing token in the sheet’s Map tokens controls. New party tokens inherit sheet control once. Effects, equipment bonuses and map combat values remain manual.',{section:'sheets'});
  topic('lighting','Token sight, darkness & darkvision','Map & editor','gm','Let players see through their assigned living characters.',[
    'Assign a living party token in Characters before enabling sight. Open Light on the rail or Table → Lighting and set the sight radius.',
    'Enable token sight. Blocked terrain and closed doors obstruct rays. The GM retains the full map.',
    'Use Environment to add lights and token darkvision. Requiring illumination combines sight with lights/darkvision.',
    'Use the player preview in Settings to inspect a generic unassigned player view. A player with no assigned living token sees darkness; verify an assigned player’s view with that player.'
  ],'Light is cell-based and binary. Light colors mark sources; blended illumination and explored-map memory are not implemented.',{section:'lighting'});
  topic('environment','Doors, windows & light sources','Map & editor','all','Prepare an environment that affects movement and sight.',[
    'The GM opens Table → Environment to add a door or window at a cell, with open, locked and secret states.',
    'Closed doors block sight and movement. Windows allow sight while blocking movement. A door cannot close over an occupied token.',
    'Players can toggle a visible unlocked door beside their assigned living character. The GM manages locked and secret doors.',
    'The GM can add enabled light sources, choose a color/radius and set token darkvision. Enable token sight in Lighting for these visibility controls.'
  ],'If a player sees darkness, check assignment, living status, sight radius, blocking cells, light requirement and darkvision.',{section:'environment'});
  topic('chat','Chat, whispers & emotes','Play together','all','Talk to the table or send a private message.',[
    'Open Table → Chat. Choose Speak as your account or an authorized character.',
    'Enter sends a message; Shift + Enter adds a line. Use /em waves for an emote.',
    'Use /w gm message or /w "Player name" message for a whisper. Player-to-player whispers are visible only to sender and recipient.',
    'Use /roll 2d6+3 or /gmroll d20 for simple dice. Read the resulting messages in the chat history.'
  ],'Choose the dice tray for self-only dice. A GM does not automatically receive whispers between other players.',{section:'chat'});
  topic('journal','Journal handouts & session notes','Characters & story','all','Keep references and share what players should know.',[
    'The GM opens Table → Journal to create a title and plain-text handout.',
    'Leave a handout GM-private while preparing. Enable sharing when the group should be able to read it.',
    'Edit or delete handouts from the same section. Players can read shared entries.',
    'Use Tools → Session notes for one shared session note. Sheets, quests and handouts remain available when scenes change.'
  ],'Handouts are text records; file attachments and folders are not available.',{section:'journal'});
  topic('quests','Quests & objectives','Characters & story','all','Show the group what they are working toward.',[
    'The GM opens Table → Quests and enters a title, description and objectives.',
    'Write [ ] before an unfinished objective and [x] before a completed one.',
    'Choose planned, active, completed or failed, then decide whether to share the quest.',
    'Save updates as the adventure advances. Players read shared quests; the GM manages progress.'
  ],'Quests persist across scene switches. Shared status is separate from the quest’s progress status.',{section:'quests'});
  topic('tables','Weighted random tables','Characters & story','all','Draw encounters, treasure or story prompts.',[
    'The GM opens Table → Tables and names a table. Write one weighted result per line, such as 1 | Quiet road and 2 | Travelling merchant.',
    'Choose whether to share the table/results and whether to draw with replacement.',
    'Save and Draw for a result. A larger weight makes an outcome more likely.',
    'Without replacement, drawn entries become unavailable. Reset restores the pool after it is depleted.'
  ],'Players read shared definitions/results. Table creation and draws are GM controls.',{section:'tables'});
  topic('macros','Personal dice macros','Play together','all','Save a frequently used dice formula.',[
    'Open Table → Macros, enter a recognizable name and a simple formula such as 2d6+3.',
    'Save the macro, then use Roll beside its name whenever needed.',
    'Definitions belong to your account. Delete one when you no longer need it.'
  ],'Running a macro posts a public roll. Advanced expressions, scripts and automatic game effects are not supported.',{section:'macros'});
  topic('music','Shared music & local volume','Play together','all','Let the GM choose a soundtrack while each listener controls playback.',[
    'The GM opens Table → Music and supplies a direct HTTP(S) audio URL and track title.',
    'Choose Play for the table and update the soundtrack.',
    'Each listener chooses Enable table audio and adjusts their own volume. Playback joins at the current track position.',
    'Mute table audio to stop listening locally without stopping the group’s track.'
  ],'The URL must be playable by the browser; a video-page link is not a direct audio file.',{section:'music'});
  topic('media','Voice, video & microphone','Play together','all','Join a direct call only when you are ready.',[
    'Open Table → Media. Select Include camera video if you want video, then Join voice & video.',
    'Allow the browser’s microphone/camera request. Use Mute microphone during the call when needed.',
    'Choose Leave call to stop capture. Leaving the table or disconnecting also stops your devices.',
    'Use localhost or HTTPS and a network that permits direct peer connections. Check browser/device permissions if capture fails.'
  ],'A relay service is not configured. Some networks will not connect; internet-wide calling is not guaranteed.',{section:'media'});
  topic('settings','Grid, distance & player preview','Map & editor','all','Set the scene scale and control your view.',[
    'Open Table → Settings. The GM can set grid dimensions, units, scale and orthogonal, straight-line or diagonal ruler distance.',
    'Use a known map distance to choose scale and verify it with Ruler. These settings affect measurement; gameplay movement keeps its existing path rules.',
    'Zoom and Fit affect your local view. The GM can open a read-only generic player preview.',
    'Exit player preview to return to GM controls. Preview represents a player without an assigned token; it does not impersonate a specific member.'
  ],'Shrinking a grid rejects out-of-bounds objects. Move or remove those objects first.',{section:'settings'});
  topic('log','Saved results, reconnect & troubleshooting','Start here','all','Understand what happened and recover your view.',[
    'Open Log to review movement, attack rolls, discoveries and other events you are allowed to see.',
    'Shared updates normally arrive automatically. If the connection drops, let it reconnect; polling is a fallback. Sync requests the latest saved state.',
    'If an action conflicts with a newer state, review the refreshed result before trying again. Do not repeatedly submit while an action is resolving.',
    'If an input seems unchanged, finish editing or blur the field so deferred updates can appear. A dark player map usually means no living assigned token or insufficient sight/light.'
  ],'Saved actions persist; your map zoom and unsent form drafts are local view state.',{tab:'log'});
  topic('adventure','Build your first adventure','Start here','all','Follow eleven lessons from a campaign idea to a prepared, playable table.',[
    'Open the adventure course from Learn. Move between lessons using the chapter list or mobile lesson picker.',
    'Prepare original outline, location, encounter, challenge and reward worksheets. Copy or download them, or stage a private Journal draft in your GM table.',
    'Use each lesson’s tool guide, completion check and stated limitations. Checkmarks record your own progress on this browser.'
  ],'The course uses Astralis workflows. Published-system mechanics, grid alignment and compendium import remain limited.');
  const byID=id=>guides.find(g=>g.id===id)||guides[0];
  const inTable=()=>currentRoute().startsWith('/table/')&&Boolean(document.querySelector('.map-world'));
  const canOpen=g=>inTable()&&g.target&&(g.role!=='gm'||state.isGM)&&!state.playerPreview;
  const button=(text,attrs='')=>`<button type="button" class="button button-secondary button-sm" ${attrs}>${text}</button>`;
  let dialog=null,returnFocus=null,query='',role='all',tourIndex=-1;
  const adventure=window.AstralisAdventureGuide?.({...ctx,openGuide:show,launchGuide:id=>launch(byID(id))});
  const storageKey=()=>`astralis-preparation-${state.user?.id||'visitor'}-${state.selectedCampaign?.id||'none'}-${state.isGM?'gm':'player'}`;
  const tasks=()=>state.isGM?[
    ['campaigns','Invite the group'],['characters','Assign party tokens'],['systems','Choose a game profile'],['imports','Prepare the scene map'],['lighting','Check player visibility'],['sheets','Prepare character sheets'],['turns','Start the encounter']
  ]:[['actions','Find your assigned character'],['sheets','Open your sheet'],['dice','Choose your dice audience'],['chat','Say hello in chat'],['media','Check optional audio and calling']];
  function progress(){try{return JSON.parse(localStorage.getItem(storageKey())||'{}')||{};}catch{return {};}}
  function checklist(){const saved=progress(),items=tasks();return `<section class="help-checklist"><h3>Before you play</h3><p class="help-muted">${items.filter(([id])=>saved[id]===true).length} of ${items.length} checked · Saved on this browser. Mark these yourself as you prepare.</p>${items.map(([id,label])=>`<div><label><input type="checkbox" data-help-check="${id}" ${saved[id]===true?'checked':''}>${e(label)}</label>${button('How',`data-help-topic="${id}" aria-label="How to: ${e(label)}"`)}</div>`).join('')}</section>`;}
  function article(g,local=false){return `<article class="help-article"><p class="help-eyebrow">${e(g.group)} · ${g.role==='gm'?'GM manages this tool':'For players & GMs'}</p><h2 ${local?'id="help-dialog-title"':''}>${e(g.title)}</h2><p class="help-summary">${e(g.summary)}</p>${['start','adventure'].includes(g.id)?'<a class="button button-secondary button-sm" href="#/help/adventure">Build an adventure →</a>':''}${g.id==='start'&&inTable()?button('Take a table tour','data-help-tour'):''}<ol class="help-steps">${g.steps.map(s=>`<li>${e(s)}</li>`).join('')}</ol><aside class="help-note">${e(g.tip)}</aside>${canOpen(g)?button('Open this tool →',`data-help-launch="${g.id}"`):''}${g.id==='start'&&inTable()?checklist():''}<div class="help-related"><span>Continue learning</span>${guides.filter(x=>x.id!==g.id&&x.group===g.group).slice(0,4).map(x=>local?button(e(x.title),`data-help-topic="${x.id}"`):`<a href="#/help/${x.id}">${e(x.title)}</a>`).join('')}</div></article>`;}
  function cards(local=false){const q=query.trim().toLowerCase();return guides.filter(g=>(role==='all'||g.role==='all'||g.role===role)&&(!q||[g.title,g.summary,g.group,g.tip,...g.steps].join(' ').toLowerCase().includes(q))).map(g=>`<${local?'button type="button"':'a'} class="help-card" ${local?`data-help-topic="${g.id}"`:`href="#/help/${g.id}"`}><span class="help-eyebrow">${e(g.group)}${g.role==='gm'?' · GM':''}</span><strong>${e(g.title)}</strong><span>${e(g.summary)}</span></${local?'button':'a'}>`).join('');}
  function filters(){return `<div class="help-filters"><label>Find a tool or task<input type="search" data-help-search placeholder="Try doors, private dice, invite…" value="${e(query)}" autocomplete="off"></label><label>Show guides for<select data-help-role><option value="all" ${role==='all'?'selected':''}>Everyone</option><option value="player" ${role==='player'?'selected':''}>Players</option><option value="gm" ${role==='gm'?'selected':''}>Game masters</option></select></label></div><p class="help-muted" data-help-count role="status"></p><div class="help-results">${cards(Boolean(dialog?.open))}</div>`;}
  function page(id,chapter){if(id==='adventure'&&adventure)return adventure.page(chapter);query='';role='all';const g=id&&guides.find(x=>x.id===id);return `${nav('learn')}<main class="app-frame help-page" id="main-content"><header class="guide-heading"><p class="help-eyebrow">ASTRALIS HANDBOOK</p><h1>${g?e(g.title):'A little guidance. A better adventure.'}</h1><p class="help-summary">Practical guides to your campaigns and every tool at the table.</p><div class="help-page-actions"><a class="button button-secondary button-sm" href="#/help">All guides</a><a class="button button-secondary button-sm" href="#/guide">RPG glossary</a><a class="button button-primary button-sm" href="${state.selectedCampaign?.id?'#/table/'+encodeURIComponent(state.selectedCampaign.id):'#/login'}">${state.selectedCampaign?.id?'Return to table':'Try a demo table'}</a></div></header>${g?article(g):`${filters()}<section class="help-note"><h2>Start with the essentials</h2><p>GM: create a campaign → invite players → assign tokens → prepare a scene → check visibility → begin turns.</p><p>Player: redeem an invitation → find your character → open your sheet → take your turn → use chat and dice.</p><a href="#/help/start">Read the first-session guide →</a></section>`}</main>${footer()}`;}
  function refreshResults(){const root=dialog?.open?dialog:document.querySelector('.help-page');if(!root)return;const results=root.querySelector('.help-results');if(!results)return;const html=cards(Boolean(dialog?.open));results.innerHTML=html||'<p class="help-empty">No guides match. Try a shorter search or choose Everyone.</p>';const count=root.querySelector('[data-help-count]');if(count)count.textContent=`${results.querySelectorAll('.help-card').length} guides found`;}
  function ensureDialog(){if(dialog)return;dialog=document.createElement('dialog');dialog.className='help-dialog';dialog.setAttribute('aria-labelledby','help-dialog-title');document.body.append(dialog);dialog.addEventListener('close',()=>{if(returnFocus?.isConnected)returnFocus.focus({preventScroll:true});else document.querySelector('[data-help-open],.tool-help')?.focus({preventScroll:true});});dialog.addEventListener('click',event=>{if(event.target===dialog){const r=dialog.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)dialog.close();}});}
  function show(id='start',search=false){endTour();ensureDialog();if(!dialog.open)returnFocus=document.activeElement;query='';role='all';dialog.innerHTML=`<header class="help-dialog-header"><button type="button" class="button button-quiet button-sm" data-help-find>Find a tool</button><a href="#/help" data-help-handbook>Full handbook ↗</a>${button('Close','data-help-close aria-label="Close help"')}</header><div class="help-dialog-content">${search?`<h2 id="help-dialog-title">Find a tool</h2>${filters()}`:article(byID(id),true)}</div>`;if(!dialog.open)dialog.showModal();if(search){refreshResults();dialog.querySelector('[data-help-search]').focus();}else dialog.querySelector('[data-help-close]').focus();}
  function launch(g){if(!canOpen(g))return;dialog?.close();state.sidebarCollapsed=false;document.querySelector('.vtt-workspace')?.classList.remove('inspector-collapsed');if(g.target.tab){selectInspectorTab(g.target.tab);}else if(g.target.section){state.vttSection=g.target.section;selectInspectorTab('table');renderTable(state.selectedCampaign);}else{const t=g.target;const selector=t.tool?`[data-map-tool="${t.tool}"]`:t.editor?`[data-editor="${t.editor}"]`:`[data-system-action="${t.system}"]`;const control=document.querySelector(selector);if(!control){notify('Use the tabletop rail to open this tool.');return;}if(control.closest('.editor-secondary-tools')?.hidden)document.querySelector('[data-editor="more"]')?.click();control.click();control.focus({preventScroll:true});return;}document.querySelector('.table-sidebar')?.classList.add('is-open');const panel=document.querySelector(`#panel-${g.target.section?'table':g.target.tab}`);panel?.setAttribute('tabindex','-1');panel?.focus({preventScroll:true});}
  const tourSteps=[
    ['.game-toolbar','Your map tools','Select and Pan explore the scene. Draw, Ruler, Dice and More keep specialist tools nearby.'],
    ['.map-viewport','The map is your workspace','Select a token to inspect it. Fit and zoom change your view. Saved moves still follow turn and permission rules.'],
    ['#tab-actions','Actions & combat','Movement, attacks and interaction live here. You can act with your assigned living character on its turn.'],
    ['#tab-characters','Your party','Inspect the roster here. The GM assigns each party character to a campaign member.'],
    ['#tab-table','The full workspace','Use the section picker for sheets, chat, journals, environment, music and more.'],
    ['[data-help-open]','Help whenever you need it','Press ? for guides or Ctrl/Command + K to find a tool. Opening help and this tour does not change your campaign.']
  ];
  function endTour(){tourIndex=-1;document.querySelectorAll('.help-tour-target').forEach(el=>el.classList.remove('help-tour-target'));document.querySelector('.help-tour')?.remove();}
  function tour(index){endTour();if(!inTable())return;tourIndex=index;const [selector,title,copy]=tourSteps[index];if(index>=2&&index<=4){state.sidebarCollapsed=false;document.querySelector('.vtt-workspace')?.classList.remove('inspector-collapsed');selectInspectorTab(['actions','characters','table'][index-2]);document.querySelector('.table-sidebar')?.classList.add('is-open');}document.querySelector(selector)?.classList.add('help-tour-target');const panel=document.createElement('section');panel.className='help-tour';panel.setAttribute('role','region');panel.setAttribute('aria-label','Table walkthrough');panel.innerHTML=`<div class="help-tour-heading"><span>TABLE TOUR · ${index+1} / ${tourSteps.length}</span>${button('Close','data-help-tour-close aria-label="Close table tour"')}</div><h2>${e(title)}</h2><p>${e(copy)}</p><div>${index?button('Back','data-help-tour-back'):''}${button(index===tourSteps.length-1?'Finish':'Next','data-help-tour-next')}</div>`;document.body.append(panel);panel.querySelector('[data-help-tour-next]').focus({preventScroll:true});}
  function afterRender(){if(!inTable())return;adventure?.afterRender();const toolbar=document.querySelector('.game-toolbar .tool-help');if(toolbar){toolbar.href='#/help';toolbar.setAttribute('aria-label','Help and tool guides');}if(!document.querySelector('[data-help-open]')){document.querySelector('.game-header')?.insertAdjacentHTML('beforeend',button('Help & tools','data-help-open title="Help (?) · Find a tool (Ctrl/Command + K)"'));}
    const contexts={actions:'actions',characters:'characters',tools:'imports',log:'log',table:state.vttSection|| (state.isGM?'scenes':'sheets')};for(const [panelID,id] of Object.entries(contexts)){const panel=document.querySelector('#panel-'+panelID);if(!panel||panel.querySelector('.help-context'))continue;const g=byID(id);panel.insertAdjacentHTML('afterbegin',`<div class="help-context"><span>${e(g.summary)}</span>${button('How to use this',`data-help-topic="${g.id}"`)}</div>`);}
    if(tourIndex>=0)document.querySelector(tourSteps[tourIndex][0])?.classList.add('help-tour-target');
  }
  function init(){adventure?.init();document.addEventListener('click',event=>{const target=event.target.closest('[data-help-open],[data-help-topic],[data-help-find],[data-help-close],[data-help-launch],[data-help-tour],[data-help-tour-next],[data-help-tour-back],[data-help-tour-close],.tool-help');if(!target)return;event.preventDefault();if(target.hasAttribute('data-help-close')){dialog?.close();return;}if(target.hasAttribute('data-help-tour-close')){endTour();document.querySelector('[data-help-open]')?.focus();return;}if(target.hasAttribute('data-help-tour-next')){if(tourIndex===tourSteps.length-1){endTour();document.querySelector('[data-help-open]')?.focus();}else tour(tourIndex+1);return;}if(target.hasAttribute('data-help-tour-back')){tour(Math.max(0,tourIndex-1));return;}if(target.hasAttribute('data-help-tour')){dialog?.close();tour(0);return;}if(target.hasAttribute('data-help-launch')){launch(byID(target.dataset.helpLaunch));return;}show(target.dataset.helpTopic||'start',target.hasAttribute('data-help-find'));});
    document.addEventListener('input',event=>{if(event.target.matches('[data-help-search]')){query=event.target.value;refreshResults();}});
    document.addEventListener('change',event=>{if(event.target.matches('[data-help-role]')){role=event.target.value;refreshResults();}if(event.target.matches('[data-help-check]')){try{const saved=progress();saved[event.target.dataset.helpCheck]=event.target.checked;localStorage.setItem(storageKey(),JSON.stringify(saved));const scroll=dialog.querySelector('.help-dialog-content').scrollTop;const checkedID=event.target.dataset.helpCheck;dialog.querySelector('.help-checklist').outerHTML=checklist();dialog.querySelector(`[data-help-check="${checkedID}"]`)?.focus({preventScroll:true});dialog.querySelector('.help-dialog-content').scrollTop=scroll;}catch{notify('Browser storage is unavailable. This checklist cannot be saved.');}}});
    document.addEventListener('keydown',event=>{if(!inTable()||event.target.closest('input,textarea,select,[contenteditable="true"]'))return;if(event.key==='Escape'&&tourIndex>=0){endTour();document.querySelector('[data-help-open]')?.focus();return;}if(event.key==='?'||(event.key.toLowerCase()==='k'&&(event.ctrlKey||event.metaKey))){event.preventDefault();show('start',event.key!=='?');}});
    window.addEventListener('hashchange',()=>{dialog?.close();endTour();});
  }
  return {page,init,afterRender,refreshResults};
};
