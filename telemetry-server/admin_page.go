package main

import "github.com/gofiber/fiber/v2"

// RegisterAdminPage serves the literary disguise page at /.
// The page always shows a literary quote. Passkey login is hidden
// behind a subtle interaction (5 clicks on the quote).
// After authentication, the user is redirected to /g/ (Grafana proxy).
func RegisterAdminPage(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(pageHTML)
	})
}

const pageHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>远方</title>
<meta name="color-scheme" content="light dark">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg: #fafaf9; --fg: #44403c; --fg2: #a8a29e; --fg3: #d6d3d1;
    --card-bg: #ffffff; --card-border: #e7e5e4; --input-bg: #f5f5f4;
    --accent: #16a34a; --accent-hover: #15803d;
    --err-bg: #fef2f2; --err-fg: #dc2626; --ok-bg: #f0fdf4; --ok-fg: #16a34a;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #0a0a0c; --fg: #d4d4d8; --fg2: #71717a; --fg3: #27272a;
      --card-bg: #161b22; --card-border: #30363d; --input-bg: #0d1117;
      --accent: #238636; --accent-hover: #2ea043;
      --err-bg: #3d1f28; --err-fg: #f85149; --ok-bg: #1b3a2d; --ok-fg: #56d364;
    }
  }

  body {
    font-family: "Noto Serif SC", "Source Han Serif CN", "Songti SC", Georgia, serif;
    background: var(--bg); color: var(--fg);
    min-height: 100vh; display: flex; align-items: center; justify-content: center;
    overflow: hidden; transition: background 0.3s, color 0.3s;
  }

  /* --- Literary page --- */
  .literary { text-align: center; padding: 3rem; max-width: 560px; animation: fadeIn 2s ease; }
  .literary blockquote {
    font-size: 1.5rem; line-height: 2.2; letter-spacing: 0.08em;
    color: var(--fg2); font-style: normal; cursor: default;
    user-select: none; -webkit-user-select: none; transition: color 0.6s;
  }
  .literary blockquote:hover { color: var(--fg); }
  .literary .author { margin-top: 1.5rem; font-size: 0.85rem; color: var(--fg2); letter-spacing: 0.12em; opacity: 0.6; }
  .literary .breath {
    width: 40px; height: 1px; background: var(--fg3);
    margin: 2rem auto; animation: breathe 4s ease-in-out infinite;
  }
  @keyframes breathe { 0%,100% { opacity: 0.3; width: 40px; } 50% { opacity: 1; width: 80px; } }
  @keyframes fadeIn { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: none; } }

  /* --- Auth overlay --- */
  .overlay {
    position: fixed; inset: 0; background: color-mix(in srgb, var(--bg) 85%, transparent);
    backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px);
    display: flex; align-items: center; justify-content: center;
    animation: fadeIn 0.3s ease; z-index: 10;
  }
  .card {
    background: var(--card-bg); border: 1px solid var(--card-border);
    border-radius: 12px; padding: 2rem; max-width: 400px; width: 90%;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  }
  .card h2 { font-size: 1.1rem; margin-bottom: 0.3rem; }
  .card .desc { font-size: 0.8rem; color: var(--fg2); margin-bottom: 1rem; line-height: 1.5; }
  .btn {
    display: inline-block; padding: 0.55rem 1.2rem; border-radius: 6px; border: none;
    font-size: 0.88rem; cursor: pointer; transition: all 0.2s; font-family: inherit;
  }
  .btn:disabled { opacity: 0.35; cursor: not-allowed; }
  .btn-primary { background: var(--accent); color: #fff; }
  .btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
  .btn-ghost { background: transparent; color: var(--fg2); font-size: 0.8rem; padding: 0.4rem 0.8rem; }
  .input {
    width: 100%; padding: 0.5rem 0.7rem; border-radius: 6px;
    border: 1px solid var(--card-border); background: var(--input-bg);
    color: var(--fg); font-size: 0.85rem; margin-bottom: 0.7rem; font-family: inherit;
  }
  .msg { font-size: 0.78rem; margin-top: 0.7rem; padding: 0.45rem 0.6rem; border-radius: 6px; }
  .msg-ok { background: var(--ok-bg); color: var(--ok-fg); }
  .msg-err { background: var(--err-bg); color: var(--err-fg); }
  .hidden { display: none !important; }
</style>
</head>
<body>

<!-- Literary page (always visible) -->
<div class="literary">
  <blockquote id="quote" onclick="onQuoteClick()">
    我们走了很远的路，<br>
    不过是为了回到最初的地方，<br>
    重新认识它。
  </blockquote>
  <div class="breath"></div>
  <div class="author">—— T.S. 艾略特</div>
</div>

<!-- Auth overlay (hidden until triggered) -->
<div id="auth-overlay" class="overlay hidden" onclick="if(event.target===this)closeOverlay()">

  <!-- Setup form -->
  <div id="phase-setup" class="card hidden">
    <h2>初始化</h2>
    <p class="desc">注册管理凭证以访问控制台。</p>
    <input id="setup-token" class="input" type="password" placeholder="Setup Token" autocomplete="off" />
    <button class="btn btn-primary" onclick="doRegister()">注册 Passkey</button>
    <button class="btn btn-ghost" onclick="closeOverlay()">取消</button>
    <div id="msg-setup" class="msg hidden"></div>
  </div>

</div>

<script>
const B = location.origin;
let hasCredential = false;
let clickCount = 0, clickTimer = null;

// Detect state on load (lightweight, no ceremony created).
(async () => {
  try {
    const r = await fetch(B+'/auth/passkey/status');
    if (r.ok) { hasCredential = (await r.json()).hasCredential; }
  } catch {}
})();

function onQuoteClick() {
  clickCount++;
  clearTimeout(clickTimer);
  clickTimer = setTimeout(() => clickCount = 0, 3000);
  if (clickCount >= 5) {
    clickCount = 0;
    if (hasCredential) {
      // Already registered — launch Passkey immediately, no card.
      doLogin();
    } else {
      // Not registered — show setup token input.
      document.getElementById('auth-overlay').classList.remove('hidden');
      document.getElementById('phase-setup').classList.remove('hidden');
      document.getElementById('phase-login').classList.add('hidden');
    }
  }
}

function closeOverlay() {
  document.getElementById('auth-overlay').classList.add('hidden');
}

function b2u(buf){return btoa(String.fromCharCode(...new Uint8Array(buf))).replace(/\+/g,'-').replace(/\//g,'_').replace(/=+$/,'');}
function u2b(s){s=s.replace(/-/g,'+').replace(/_/g,'/');while(s.length%4)s+='=';return Uint8Array.from(atob(s),c=>c.charCodeAt(0)).buffer;}

function showMsg(id,text,ok){
  const el=document.getElementById(id);
  el.textContent=text; el.className='msg '+(ok?'msg-ok':'msg-err'); el.classList.remove('hidden');
}

// --- Register ---
async function doRegister(){
  const token=document.getElementById('setup-token').value.trim();
  if(!token){showMsg('msg-setup','请输入 Setup Token',false);return;}
  try{
    const r1=await fetch(B+'/auth/passkey/register/begin',{method:'POST',headers:{'X-Setup-Token':token}});
    if(!r1.ok){showMsg('msg-setup','操作失败，请重试',false);return;}
    const{publicKey:opts,sessionId}=await r1.json();
    opts.publicKey.challenge=u2b(opts.publicKey.challenge);
    opts.publicKey.user.id=u2b(opts.publicKey.user.id);
    if(opts.publicKey.excludeCredentials)opts.publicKey.excludeCredentials=opts.publicKey.excludeCredentials.map(c=>({...c,id:u2b(c.id)}));
    const cred=await navigator.credentials.create(opts);
    const body={id:cred.id,rawId:b2u(cred.rawId),type:cred.type,response:{attestationObject:b2u(cred.response.attestationObject),clientDataJSON:b2u(cred.response.clientDataJSON)}};
    const r2=await fetch(B+'/auth/passkey/register/finish',{method:'POST',headers:{'Content-Type':'application/json','X-Session-Id':sessionId},body:JSON.stringify(body)});
    if(!r2.ok){showMsg('msg-setup','操作失败，请重试',false);return;}
    hasCredential=true;
    location.href='/g/';
  }catch(e){showMsg('msg-setup','操作失败，请重试',false);}
}

// --- Login (called directly from quote click, no card shown) ---
async function doLogin(){
  try{
    const r1=await fetch(B+'/auth/passkey/login/begin',{method:'POST'});
    if(!r1.ok) throw new Error((await r1.json()).error);
    const{publicKey:opts,sessionId}=await r1.json();
    opts.publicKey.challenge=u2b(opts.publicKey.challenge);
    if(opts.publicKey.allowCredentials)opts.publicKey.allowCredentials=opts.publicKey.allowCredentials.map(c=>({...c,id:u2b(c.id)}));
    const assertion=await navigator.credentials.get(opts);
    const body={id:assertion.id,rawId:b2u(assertion.rawId),type:assertion.type,response:{authenticatorData:b2u(assertion.response.authenticatorData),clientDataJSON:b2u(assertion.response.clientDataJSON),signature:b2u(assertion.response.signature),userHandle:assertion.response.userHandle?b2u(assertion.response.userHandle):''}};
    const r2=await fetch(B+'/auth/passkey/login/finish',{method:'POST',headers:{'Content-Type':'application/json','X-Session-Id':sessionId},body:JSON.stringify(body)});
    if(!r2.ok) throw new Error((await r2.json()).error);
    // Success — redirect straight to Grafana, no intermediate UI.
    location.href='/g/';
  }catch(e){
    // Fail silently — maintain disguise, no error shown.
  }
}
</script>
</body>
</html>`
