package server

const editorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>marky editor</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:system-ui,sans-serif;height:100vh;display:flex;flex-direction:column;background:#1a1a2e;color:#e0e0e0}
header{padding:12px 20px;background:#16213e;border-bottom:1px solid #0f3460;display:flex;align-items:center;justify-content:space-between}
header h1{font-size:18px;color:#e94560;font-weight:700;letter-spacing:1px}
#conn-status{font-size:12px;color:#a0a0b0}
.editor-area{display:flex;flex:1;overflow:hidden}
.pane{flex:1;display:flex;flex-direction:column;overflow:hidden}
.pane:first-child{border-right:1px solid #0f3460}
.pane-header{padding:8px 16px;font-size:11px;color:#a0a0b0;background:#16213e;border-bottom:1px solid #0f3460;text-transform:uppercase;letter-spacing:1px}
#editor{flex:1;padding:20px;font-family:'Cascadia Code','Fira Code',monospace;font-size:14px;line-height:1.6;background:#1a1a2e;color:#e0e0e0;border:none;outline:none;resize:none;tab-size:2}
#preview{flex:1;padding:20px;overflow-y:auto;line-height:1.7}
footer{padding:10px 20px;background:#16213e;border-top:1px solid #0f3460;display:flex;align-items:center;gap:12px}
#save-btn{padding:6px 20px;background:#e94560;color:#fff;border:none;border-radius:4px;cursor:pointer;font-size:14px}
#save-btn:hover{background:#c73652}
#save-status{font-size:13px;color:#a0a0b0}
#preview h1,#preview h2,#preview h3,#preview h4,#preview h5,#preview h6{margin:.8em 0 .4em;color:#e0e0e0}
#preview p{margin:.5em 0}
#preview code{background:#0f3460;padding:2px 6px;border-radius:3px;font-family:monospace;font-size:13px}
#preview pre{background:#0f3460;padding:16px;border-radius:6px;overflow-x:auto;margin:.8em 0}
#preview pre code{background:none;padding:0}
#preview blockquote{border-left:4px solid #e94560;padding-left:12px;color:#a0a0b0;margin:.8em 0}
#preview ul,#preview ol{padding-left:1.5em;margin:.5em 0}
#preview li{margin:.2em 0}
#preview hr{border:none;border-top:1px solid #0f3460;margin:1em 0}
#preview a{color:#e94560}
#preview strong{color:#fff;font-weight:700}
</style>
</head>
<body>
<header><h1>marky editor</h1><span id="conn-status">connecting…</span></header>
<div class="editor-area">
  <div class="pane"><div class="pane-header">Markdown Source</div><textarea id="editor" spellcheck="false"></textarea></div>
  <div class="pane"><div class="pane-header">Preview</div><div id="preview"></div></div>
</div>
<footer><button id="save-btn">Save</button><span id="save-status"></span></footer>
<script>
function md2html(s){
  var blocks=[];
  s=s.replace(/` + "`" + `{3}(\w*)\n([\s\S]*?)` + "`" + `{3}/g,function(_,lang,code){
    var i=blocks.length;
    blocks.push('<pre><code class="lang-'+lang+'">'+code.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')+'</code></pre>');
    return'\x00block'+i+'\x00';
  });
  s=s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  s=s.replace(/` + "`" + `([^` + "`" + `]+)` + "`" + `/g,'<code>$1</code>');
  s=s.replace(/^#{6} (.+)$/gm,'<h6>$1</h6>');
  s=s.replace(/^#{5} (.+)$/gm,'<h5>$1</h5>');
  s=s.replace(/^#{4} (.+)$/gm,'<h4>$1</h4>');
  s=s.replace(/^#{3} (.+)$/gm,'<h3>$1</h3>');
  s=s.replace(/^#{2} (.+)$/gm,'<h2>$1</h2>');
  s=s.replace(/^# (.+)$/gm,'<h1>$1</h1>');
  s=s.replace(/^\&gt; (.+)$/gm,'<blockquote>$1</blockquote>');
  s=s.replace(/^---$/gm,'<hr>');
  s=s.replace(/\*\*(.+?)\*\*/g,'<strong>$1</strong>');
  s=s.replace(/\*(.+?)\*/g,'<em>$1</em>');
  s=s.replace(/\[([^\]]+)\]\(([^)]+)\)/g,'<a href="$2" target="_blank">$1</a>');
  s=s.replace(/^(\d+\. .+\n?)+/gm,function(m){return'<ol>'+m.replace(/^\d+\. (.+)$/gm,'<li>$1</li>')+'</ol>';});
  s=s.replace(/^(- .+\n?)+/gm,function(m){return'<ul>'+m.replace(/^- (.+)$/gm,'<li>$1</li>')+'</ul>';});
  s=s.replace(/\n\n+/g,'</p><p>');
  s=s.replace(/^(?!<[a-z\/])(.+)$/gm,'<p>$1</p>');
  s=s.replace(/\x00block(\d+)\x00/g,function(_,i){return blocks[+i];});
  return s;
}

var editor=document.getElementById('editor');
var preview=document.getElementById('preview');
var connStatus=document.getElementById('conn-status');
var saveStatus=document.getElementById('save-status');

fetch('/content').then(function(r){return r.json();}).then(function(d){
  editor.value=d.content;
  preview.innerHTML=md2html(d.content);
});

editor.addEventListener('input',function(){preview.innerHTML=md2html(editor.value);});

function save(){
  saveStatus.textContent='Saving…';
  fetch('/save',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({content:editor.value})})
    .then(function(r){if(!r.ok)saveStatus.textContent='Save failed ('+r.status+')';})
    .catch(function(e){saveStatus.textContent='Error: '+e.message;});
}

document.getElementById('save-btn').addEventListener('click',save);
editor.addEventListener('keydown',function(e){
  if((e.ctrlKey||e.metaKey)&&e.key==='s'){e.preventDefault();save();}
});

var ws=new WebSocket('ws://'+location.host+'/ws');
ws.onopen=function(){connStatus.textContent='connected';};
ws.onmessage=function(e){
  var msg=JSON.parse(e.data);
  if(msg.status==='saved'){saveStatus.textContent='Saved at '+msg.time;}
  if(msg.status==='error'){saveStatus.textContent='Error: '+msg.error;}
};
ws.onclose=function(){connStatus.textContent='disconnected';};
</script>
</body>
</html>`
