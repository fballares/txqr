package main

// HTML for settings/coach and the QR overlay popup.
// Placeholders are filled with fmt.Fprintf from server handlers.
// Literal % in CSS/JS must be written as %%.

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>txqr Send</title>
<style>
:root {
  --ink: #0f1419;
  --paper: #f2efe8;
  --accent: #c45c26;
  --line: #d4cfc4;
  --muted: #5c6570;
  --ok: #1f7a4c;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  font-family: "Segoe UI", "Helvetica Neue", sans-serif;
  color: var(--ink);
  background:
    radial-gradient(1200px 600px at 10%% -10%%, #fff8ef 0%%, transparent 55%%),
    radial-gradient(900px 500px at 100%% 0%%, #e8f0ea 0%%, transparent 50%%),
    var(--paper);
  min-height: 100%%;
}
.wrap { max-width: 720px; margin: 0 auto; padding: 2.5rem 1.25rem 4rem; }
.brand {
  font-family: Consolas, "Courier New", monospace;
  font-size: clamp(2.4rem, 8vw, 3.4rem);
  font-weight: 700;
  letter-spacing: -0.04em;
  margin: 0 0 0.35rem;
  line-height: 1;
}
.lead { color: var(--muted); font-size: 1.05rem; margin: 0 0 2rem; max-width: 36rem; }
.panel {
  border-top: 1px solid var(--line);
  padding: 1.4rem 0 1.6rem;
}
.panel h2 {
  margin: 0 0 0.45rem;
  font-size: 1.15rem;
  font-weight: 700;
}
.panel p { margin: 0 0 0.9rem; color: var(--muted); line-height: 1.45; }
.coach {
  display: grid;
  gap: 0.75rem;
  margin: 0 0 1rem;
  padding: 0;
  list-style: none;
  counter-reset: step;
}
.coach li {
  display: grid;
  grid-template-columns: 2rem 1fr;
  gap: 0.75rem;
  align-items: start;
}
.coach li::before {
  counter-increment: step;
  content: counter(step);
  font-family: Consolas, "Courier New", monospace;
  font-weight: 700;
  width: 2rem; height: 2rem;
  border-radius: 999px;
  background: var(--ink);
  color: #fff;
  display: grid;
  place-items: center;
  font-size: 0.85rem;
}
kbd {
  font-family: Consolas, "Courier New", monospace;
  font-size: 0.85em;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 6px;
  padding: 0.12rem 0.4rem;
}
.row { display: flex; flex-wrap: wrap; gap: 0.6rem; margin-top: 0.75rem; }
button, .btn {
  font: inherit;
  border: 0;
  border-radius: 10px;
  padding: 0.7rem 1rem;
  background: var(--ink);
  color: #fff;
  cursor: pointer;
  text-decoration: none;
  display: inline-block;
}
button.secondary, .btn.secondary { background: #fff; color: var(--ink); border: 1px solid var(--line); }
button:disabled { opacity: 0.45; cursor: default; }
.status {
  font-family: Consolas, "Courier New", monospace;
  font-size: 0.9rem;
  color: var(--muted);
  min-height: 1.2em;
  margin-top: 0.75rem;
}
.status.ok { color: var(--ok); }
.status.err { color: #a32020; }
label { display: block; font-size: 0.92rem; margin: 0.7rem 0 0.3rem; }
input[type=text], select, textarea {
  width: 100%%;
  max-width: 22rem;
  font: inherit;
  padding: 0.55rem 0.7rem;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: #fff;
}
textarea { max-width: 100%%; min-height: 120px; resize: vertical; }
.hint { font-size: 0.85rem; color: var(--muted); margin-top: 0.35rem; }
</style>
</head>
<body>
<main class="wrap">
  <h1 class="brand">txqr</h1>
  <p class="lead">Air-gapped clipboard transfer: Windows encodes text as animated QR codes; your iPhone reconstructs and copies it. Nothing leaves the room except light.</p>

  <section class="panel" id="first-run">
    <h2>First run</h2>
    <ol class="coach">
      <li><div>Install <strong>TXQR Reader</strong> on your iPhone (Xcode → open <code>ios/TXQRReader</code>).</div></li>
      <li><div>On this PC, keep <strong>txqr-send</strong> running in the tray.</div></li>
      <li><div>Copy text, press <kbd>%s</kbd>, aim the phone at the floating codes.</div></li>
      <li><div>Wait for the ring to fill → tap Copy. Dual QR: phone sorts Left then Right automatically.</div></li>
    </ol>
    <button type="button" class="secondary" id="dismiss-coach">Got it — hide this</button>
  </section>

  <section class="panel">
    <h2>Transfer now</h2>
    <p>Uses the current clipboard. Prefer UTF-8 text under a few hundred KB.</p>
    <div class="row">
      <button type="button" id="send">Send clipboard</button>
      <a class="btn secondary" href="/popup" target="_blank" rel="noopener">Open overlay</a>
      <a class="btn secondary" href="/popup?stand=1" target="_blank" rel="noopener">Phone-stand mode</a>
    </div>
    <div class="status" id="status" aria-live="polite"></div>
  </section>

  <section class="panel">
    <h2>Paste &amp; encode</h2>
    <p>Optional: paste text here instead of using the clipboard.</p>
    <textarea id="text" placeholder="Paste text to encode…"></textarea>
    <div class="row">
      <button type="button" id="encode">Encode pasted text</button>
    </div>
    <div class="status" id="encode-status" aria-live="polite"></div>
  </section>

  <section class="panel">
    <h2>Settings</h2>
    <form id="settings">
      <label for="hotkey">Global hotkey (restart required)</label>
      <input id="hotkey" name="hotkey" type="text" value="%s" autocomplete="off" spellcheck="false">
      <p class="hint">Examples: Ctrl+Shift+Q · Alt+Q · Ctrl+Alt+T. Change via <code>-hotkey</code> flag or restart after saving.</p>

      <label for="streams">Parallel QR streams</label>
      <select id="streams" name="streams">
        <option value="0"%s>Auto (1 or 2)</option>
        <option value="1"%s>1 — single QR</option>
        <option value="2"%s>2 — Left + Right</option>
        <option value="3"%s>3</option>
        <option value="4"%s>4</option>
      </select>
      <p class="hint">Auto uses dual QR only when it clearly helps. Left = stream 0, Right = stream 1.</p>

      <div class="row">
        <button type="submit">Save settings</button>
      </div>
    </form>
    <div class="status" id="settings-status" aria-live="polite"></div>
  </section>

  <section class="panel">
    <h2>If something goes wrong</h2>
    <p><strong>Empty clipboard</strong> — copy text first, then hotkey.<br>
    <strong>Overlay behind other windows</strong> — tray → Show overlay, or Phone-stand mode.<br>
    <strong>Phone stuck at 0%%</strong> — closer, less glare, torch on, keep codes fully in frame.<br>
    <strong>CRC failed</strong> — keep scanning; sender loops until a clean reconstruct verifies.</p>
  </section>
</main>
<script>
(function () {
  const status = document.getElementById("status");
  const encodeStatus = document.getElementById("encode-status");
  const settingsStatus = document.getElementById("settings-status");
  const coach = document.getElementById("first-run");
  if (localStorage.getItem("txqr_coach_done") === "1") coach.style.display = "none";
  document.getElementById("dismiss-coach").onclick = function () {
    localStorage.setItem("txqr_coach_done", "1");
    coach.style.display = "none";
  };

  function summarize(data) {
    return "Ready — " + data.bytes + " bytes · " + data.frame_count + " frames · " +
      data.streams + " QR · ~" + (data.eta_seconds || 0).toFixed(1) + "s est · CRC " + (data.crc_hex || data.crc32 || "");
  }

  document.getElementById("send").addEventListener("click", async function () {
    status.className = "status";
    status.textContent = "Encoding…";
    try {
      const res = await fetch("/api/send", { method: "POST" });
      const data = await res.json();
      if (!res.ok || data.error) throw new Error(data.error || res.statusText);
      status.className = "status ok";
      status.textContent = summarize(data);
      window.open(data.popup || "/popup", "txqr-overlay", "noopener");
    } catch (err) {
      status.className = "status err";
      status.textContent = String(err.message || err);
    }
  });

  document.getElementById("encode").addEventListener("click", async function () {
    encodeStatus.className = "status";
    encodeStatus.textContent = "Encoding…";
    try {
      const res = await fetch("/api/encode", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: document.getElementById("text").value, auto: true })
      });
      const data = await res.json();
      if (!res.ok || data.error) throw new Error(data.error || res.statusText);
      encodeStatus.className = "status ok";
      encodeStatus.textContent = summarize(data);
      window.open("/popup", "txqr-overlay", "noopener");
    } catch (err) {
      encodeStatus.className = "status err";
      encodeStatus.textContent = String(err.message || err);
    }
  });

  document.getElementById("settings").addEventListener("submit", async function (e) {
    e.preventDefault();
    settingsStatus.className = "status";
    settingsStatus.textContent = "Saving…";
    const body = new URLSearchParams({
      hotkey: document.getElementById("hotkey").value.trim(),
      streams: document.getElementById("streams").value
    });
    try {
      const res = await fetch("/api/settings", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body
      });
      const data = await res.json();
      if (!res.ok || data.error) throw new Error(data.error || res.statusText);
      settingsStatus.className = "status ok";
      settingsStatus.textContent = data.message || "Saved.";
    } catch (err) {
      settingsStatus.className = "status err";
      settingsStatus.textContent = String(err.message || err);
    }
  });
})();
</script>
</body>
</html>
`

const popupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>txqr</title>
<style>
:root {
  --bg: #07090c;
  --fg: #f4f1ea;
  --muted: #8b939e;
  --accent: #e07a3d;
  --guide: rgba(224, 122, 61, 0.85);
}
* { box-sizing: border-box; }
html, body {
  margin: 0;
  height: 100%%;
  background: var(--bg);
  color: var(--fg);
  font-family: "Segoe UI", system-ui, sans-serif;
  overflow: hidden;
  user-select: none;
}
body.stand {
  display: flex;
  align-items: center;
  justify-content: center;
}
.stage {
  height: 100%%;
  display: flex;
  flex-direction: column;
  padding: 10px 12px 8px;
}
body.stand .stage {
  width: min(96vw, 1200px);
  height: min(92vh, 900px);
}
.hud {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
  flex-shrink: 0;
}
.brand {
  font-weight: 700;
  letter-spacing: 0.04em;
  font-size: 13px;
  text-transform: uppercase;
}
.meta {
  font-size: 12px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
  text-align: right;
}
.meta strong { color: var(--fg); font-weight: 600; }
.codes {
  flex: 1;
  display: grid;
  gap: 14px;
  min-height: 0;
  align-items: stretch;
}
.codes.n1 { grid-template-columns: 1fr; }
.codes.n2 { grid-template-columns: 1fr 1fr; }
.codes.n3 { grid-template-columns: 1fr 1fr 1fr; }
.codes.n4 { grid-template-columns: 1fr 1fr; grid-template-rows: 1fr 1fr; }
.slot {
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}
.guide {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 2;
}
.guide .corner {
  position: absolute;
  width: 22px;
  height: 22px;
  border-color: var(--guide);
  border-style: solid;
  border-width: 0;
}
.guide .tl { top: 2px; left: 2px; border-top-width: 3px; border-left-width: 3px; }
.guide .tr { top: 2px; right: 2px; border-top-width: 3px; border-right-width: 3px; }
.guide .bl { bottom: 28px; left: 2px; border-bottom-width: 3px; border-left-width: 3px; }
.guide .br { bottom: 28px; right: 2px; border-bottom-width: 3px; border-right-width: 3px; }
.label {
  text-align: center;
  font-size: 11px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--accent);
  margin-bottom: 4px;
  font-weight: 700;
}
.frame-wrap {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 4px;
  padding: 6px;
}
.frame-wrap svg {
  width: 100%%;
  height: 100%%;
  max-width: 100%%;
  max-height: 100%%;
  object-fit: contain;
  display: block;
  shape-rendering: crispEdges;
}
.foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--muted);
}
.bar {
  flex: 1;
  height: 4px;
  background: #1c222a;
  border-radius: 2px;
  overflow: hidden;
}
.bar > i {
  display: block;
  height: 100%%;
  width: 0%%;
  background: var(--accent);
  transition: width 0.15s linear;
}
.paused .bar > i { background: #888; }
.hint { white-space: nowrap; }
.empty {
  margin: auto;
  text-align: center;
  color: var(--muted);
  padding: 2rem;
  grid-column: 1 / -1;
}
</style>
</head>
<body class="%s">
<div class="stage" id="root">
  <div class="hud">
    <div class="brand">txqr</div>
    <div class="meta" id="meta">Waiting…</div>
  </div>
  <div id="codes" class="codes"></div>
  <div class="foot">
    <div class="bar" aria-hidden="true"><i id="loopbar"></i></div>
    <div class="hint" id="hint">Aim phone · %s</div>
  </div>
</div>
<script>
(function () {
  const labels = ["LEFT", "RIGHT", "C", "D"];
  const codes = document.getElementById("codes");
  const meta = document.getElementById("meta");
  const hint = document.getElementById("hint");
  const loopbar = document.getElementById("loopbar");
  const root = document.getElementById("root");
  const hotkey = %q;
  let streamCount = %d;
  let total = 0, fps = 8, bytes = 0, crc = "", paused = false, eta = 0;
  let tick = 0, lastTs = 0, raf = 0, slots = [];
  let svgCache = {}, inflight = {};

  function frameIndex(t, stream, nStreams, n) {
    if (n <= 0) return 0;
    const offset = Math.floor((n * stream) / nStreams);
    return (t + offset) %% n;
  }

  function buildSlots(n) {
    codes.innerHTML = "";
    slots = [];
    n = Math.max(0, Math.min(4, n | 0));
    codes.className = "codes n" + Math.max(1, n || 1);
    if (!n || !total) {
      codes.innerHTML = '<div class="empty">No active transfer.<br>Copy text and press ' + hotkey + '.</div>';
      return;
    }
    for (let s = 0; s < n; s++) {
      const slot = document.createElement("div");
      slot.className = "slot";
      slot.innerHTML =
        (n > 1 ? '<div class="label">' + (labels[s] || ("S" + (s + 1))) + '</div>' : '') +
        '<div class="guide"><span class="corner tl"></span><span class="corner tr"></span>' +
        '<span class="corner bl"></span><span class="corner br"></span></div>' +
        '<div class="frame-wrap"><div class="host" id="host' + s + '"></div></div>';
      codes.appendChild(slot);
      slots.push(slot.querySelector(".host"));
    }
  }

  async function svgAt(i) {
    if (svgCache[i]) return svgCache[i];
    if (inflight[i]) return inflight[i];
    inflight[i] = fetch("/api/frame?i=" + i)
      .then(r => { if (!r.ok) throw new Error("frame"); return r.text(); })
      .then(svg => { svgCache[i] = svg; delete inflight[i]; return svg; })
      .catch(err => { delete inflight[i]; throw err; });
    return inflight[i];
  }

  function setHostSVG(host, svg) {
    host.innerHTML = svg;
    const el = host.firstElementChild;
    if (el) {
      el.removeAttribute("width");
      el.removeAttribute("height");
      el.setAttribute("preserveAspectRatio", "xMidYMid meet");
      el.style.width = "100%%";
      el.style.height = "100%%";
    }
  }

  async function paint() {
    if (!total || !slots.length) return;
    const jobs = slots.map((host, s) => {
      const idx = frameIndex(tick, s, streamCount, total);
      return svgAt(idx).then(svg => setHostSVG(host, svg));
    });
    await Promise.all(jobs);
  }

  function updateHUD() {
    if (!total) {
      meta.textContent = "Idle";
      loopbar.style.width = "0%%";
      return;
    }
    const loopSec = (total / Math.max(fps, 1)).toFixed(1);
    const frame = (tick %% total) + 1;
    meta.innerHTML = "<strong>" + bytes + "</strong> B · <strong>" + streamCount + "</strong> QR · " +
      "frame <strong>" + frame + "/" + total + "</strong> · ~" + loopSec + "s/loop" +
      (eta ? " · ~" + eta.toFixed(1) + "s est" : "") +
      (crc ? " · CRC <strong>" + crc + "</strong>" : "") +
      (paused ? " · <strong>PAUSED</strong>" : "");
    loopbar.style.width = ((frame / total) * 100).toFixed(1) + "%%";
    hint.textContent = paused ? "Paused — Space or tray Resume" : "Aim phone · " + hotkey;
  }

  async function refreshStatus() {
    try {
      const res = await fetch("/api/status");
      const st = await res.json();
      const nextTotal = st.total || st.frame_count || 0;
      const nextStreams = st.streams || streamCount || 1;
      const nextCRC = st.crc32 || st.crc_hex || "";
      const changed = nextTotal !== total || nextStreams !== streamCount || nextCRC !== crc;
      total = nextTotal;
      fps = st.fps || 8;
      bytes = st.bytes || 0;
      crc = nextCRC;
      eta = st.eta_seconds || 0;
      paused = !!st.paused;
      streamCount = total ? Math.max(1, Math.min(4, nextStreams)) : 0;
      root.classList.toggle("paused", paused);
      if (changed) {
        svgCache = {};
        tick = 0;
        buildSlots(streamCount);
      }
      updateHUD();
    } catch (_) {}
  }

  async function loop(ts) {
    raf = requestAnimationFrame(loop);
    if (!lastTs) lastTs = ts;
    if (paused || !total) return;
    const interval = 1000 / Math.max(fps, 1);
    if (ts - lastTs >= interval) {
      lastTs = ts;
      await paint();
      tick++;
      updateHUD();
    }
  }

  document.addEventListener("keydown", function (e) {
    if (e.code === "Space") {
      e.preventDefault();
      fetch("/api/control", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: paused ? "resume" : "pause" })
      }).then(refreshStatus);
    }
    if (e.key === "Escape") window.close();
  });

  buildSlots(streamCount);
  refreshStatus();
  setInterval(refreshStatus, 800);
  raf = requestAnimationFrame(loop);
})();
</script>
</body>
</html>
`
