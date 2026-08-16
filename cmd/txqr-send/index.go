package main

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>TXQR Send</title>
  <style>
    :root {
      --bg0: #10151c;
      --bg1: #182232;
      --ink: #e8eef7;
      --muted: #8b9bb4;
      --accent: #3d9cf0;
      --accent-ink: #041018;
      --line: rgba(232, 238, 247, 0.12);
      --ok: #3ecf8e;
    }
    * { box-sizing: border-box; }
    html, body {
      margin: 0;
      min-height: 100%%;
      background:
        radial-gradient(1100px 560px at 8%% -12%%, #1e3a5f 0%%, transparent 55%%),
        radial-gradient(800px 460px at 100%% 0%%, #143028 0%%, transparent 50%%),
        linear-gradient(160deg, var(--bg0), var(--bg1));
      color: var(--ink);
      font-family: "Segoe UI", "Helvetica Neue", sans-serif;
    }
    body { display: grid; place-items: center; padding: 24px; }
    .shell { width: min(880px, 100%%); display: grid; gap: 18px; }
    header h1 {
      margin: 0;
      font-size: clamp(2rem, 4vw, 2.7rem);
      letter-spacing: -0.03em;
    }
    header p { margin: 8px 0 0; color: var(--muted); line-height: 1.45; max-width: 40rem; }
    kbd {
      display: inline-block;
      padding: 2px 8px;
      border: 1px solid var(--line);
      background: rgba(255,255,255,0.04);
      font: inherit;
      font-size: 0.92em;
    }
    .panel {
      display: grid; gap: 14px; padding: 18px;
      border: 1px solid var(--line);
      background: rgba(8, 12, 18, 0.45);
    }
    label { display: grid; gap: 6px; font-size: 0.85rem; color: var(--muted); }
    textarea, input {
      width: 100%%; border: 1px solid var(--line);
      background: rgba(255,255,255,0.03); color: var(--ink);
      padding: 12px 14px; font: inherit;
    }
    textarea { min-height: 140px; resize: vertical; }
    .row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
    @media (max-width: 700px) { .row { grid-template-columns: 1fr; } }
    .actions { display: flex; flex-wrap: wrap; gap: 10px; }
    button {
      appearance: none; border: 0; padding: 12px 18px;
      font: inherit; font-weight: 600; cursor: pointer;
      background: var(--accent); color: var(--accent-ink);
    }
    button.secondary {
      background: transparent; color: var(--ink); border: 1px solid var(--line);
    }
    .error { color: #ff8e8e; min-height: 1.2em; }
    .meta { color: var(--muted); font-size: 0.9rem; }
    .meta strong { color: var(--ok); }
  </style>
</head>
<body>
  <div class="shell">
    <header>
      <h1>TXQR Send</h1>
      <p>Runs in the background. Copy any text, press <kbd>%s</kbd>, and a QR popup appears for your phone reader.</p>
    </header>
    <section class="panel">
      <label>Manual text (optional)
        <textarea id="text" placeholder="Or paste here to preview without the hotkey…"></textarea>
      </label>
      <div class="row">
        <label>Chunk override (0 = auto)
          <input id="chunk" type="number" min="0" max="1000" value="0" />
        </label>
        <label>FPS
          <input id="fps" type="number" min="1" max="15" value="%d" />
        </label>
        <label>QR size
          <input id="size" type="number" min="200" max="800" value="%d" />
        </label>
      </div>
      <div class="actions">
        <button id="send" type="button">Show QR popup</button>
        <button id="paste" class="secondary" type="button">Paste clipboard</button>
      </div>
      <div class="error" id="error"></div>
      <div class="meta" id="hint">Auto mode picks chunk/FPS/redundancy from paste size for reliable phone scanning.</div>
    </section>
  </div>
  <script>
    const errorEl = document.getElementById('error');
    const textEl = document.getElementById('text');

    document.getElementById('paste').onclick = async () => {
      errorEl.textContent = '';
      try { textEl.value = await navigator.clipboard.readText(); }
      catch { errorEl.textContent = 'Clipboard blocked — use Ctrl+V.'; }
    };

    document.getElementById('send').onclick = async () => {
      errorEl.textContent = '';
      const chunk = Number(document.getElementById('chunk').value);
      const body = {
        text: textEl.value,
        chunk_len: chunk,
        fps: Number(document.getElementById('fps').value),
        qr_size: Number(document.getElementById('size').value),
        auto: !chunk
      };
      try {
        const res = await fetch('/api/encode', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        if (!res.ok || data.error) throw new Error(data.error || ('HTTP ' + res.status));
        window.open('/popup?t=' + Date.now(), 'txqr-popup', 'noopener,width=720,height=780');
      } catch (err) {
        errorEl.textContent = err.message || String(err);
      }
    };
  </script>
</body>
</html>
`

const popupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>TXQR Overlay</title>
  <style>
    html, body {
      margin: 0; height: 100%%;
      background: #ffffff;
      color: #111;
      font-family: "Segoe UI", "Helvetica Neue", sans-serif;
      user-select: none;
      overflow: hidden;
    }
    body {
      display: grid;
      grid-template-rows: auto 1fr auto;
      height: 100%%;
      min-height: 100%%;
    }
    .banner {
      padding: 10px 12px;
      background: #0f1720;
      color: #e8eef7;
      font-size: 0.82rem;
      line-height: 1.35;
    }
    .banner strong { color: #3ecf8e; font-weight: 650; }
    .stage {
      display: grid;
      gap: 10px;
      place-items: center;
      align-content: center;
      padding: 12px;
      background: #ffffff;
      min-height: 0;
      height: 100%%;
    }
    .stage.multi {
      grid-template-columns: 1fr 1fr;
      column-gap: 16px;
    }
    .slot {
      display: grid;
      gap: 6px;
      justify-items: center;
      width: 100%%;
      height: 100%%;
      min-height: 0;
      align-content: center;
    }
    .slot label {
      font-size: 0.7rem;
      font-weight: 700;
      letter-spacing: 0.08em;
      color: #333;
      text-transform: uppercase;
    }
    .qr-host {
      /* Extra white quiet margin around the vector QR for phone cameras. */
      box-sizing: border-box;
      width: min(100%%, calc(100vh - 150px));
      max-width: 100%%;
      aspect-ratio: 1 / 1;
      padding: 7%%;
      background: #ffffff;
      border: 1px solid #eee;
    }
    .stage.multi .qr-host {
      width: min(100%%, calc((100vh - 150px) * 0.92));
    }
    .qr-host svg {
      display: block;
      width: 100%%;
      height: 100%%;
      shape-rendering: crispEdges;
    }
    footer {
      display: flex;
      justify-content: space-between;
      gap: 8px;
      flex-wrap: wrap;
      padding: 8px 12px 10px;
      border-top: 1px solid #e6e6e6;
      background: #f7f7f7;
      font-size: 0.78rem;
      color: #444;
    }
    .err { color: #b00020; text-align: center; padding: 18px; font-size: 0.9rem; grid-column: 1 / -1; }
    kbd {
      padding: 1px 5px;
      border: 1px solid #ccc;
      background: #fff;
      font: inherit;
      font-size: 0.85em;
    }
  </style>
</head>
<body>
  <div class="banner">
    <strong>Vector QR</strong> — high-contrast SVG, scales as you resize.
    Dual mode: <strong>LEFT</strong> + <strong>RIGHT</strong>. Drag to your phone stand.
  </div>
  <div class="stage" id="stage">
    <div class="err" id="status">Preparing QR stream…</div>
  </div>
  <footer>
    <div id="stats">Hotkey <kbd>%s</kbd> · auto 1↔%d QR</div>
    <div id="loop">Esc closes</div>
  </footer>
  <script>
    const stage = document.getElementById('stage');
    const stats = document.getElementById('stats');
    const loopEl = document.getElementById('loop');
    const defaultStreams = %d;
    let frameCount = 0;
    let inlineFrames = [];
    let svgCache = {};
    let tickN = 0;
    let loops = 0;
    let streams = 1;
    let timer = null;
    let inflight = {};

    function frameIndex(tick, stream, streamCount, n) {
      if (n <= 0) return 0;
      const offset = Math.floor((n * stream) / streamCount);
      return (tick + offset) %% n;
    }

    async function svgAt(i) {
      if (inlineFrames[i]) return inlineFrames[i];
      if (svgCache[i]) return svgCache[i];
      if (inflight[i]) return inflight[i];
      inflight[i] = fetch('/api/frame?i=' + i)
        .then(r => {
          if (!r.ok) throw new Error('frame ' + i);
          return r.text();
        })
        .then(svg => {
          svgCache[i] = svg;
          delete inflight[i];
          return svg;
        })
        .catch(err => {
          delete inflight[i];
          throw err;
        });
      return inflight[i];
    }

    function setHostSVG(host, svg) {
      host.innerHTML = svg;
      const el = host.querySelector('svg');
      if (el) {
        el.removeAttribute('width');
        el.removeAttribute('height');
        el.setAttribute('width', '100%%');
        el.setAttribute('height', '100%%');
        el.setAttribute('preserveAspectRatio', 'xMidYMid meet');
      }
    }

    async function showStatic(svg) {
      stage.classList.remove('multi');
      stage.innerHTML = '<div class="slot"><div class="qr-host" id="host0"></div></div>';
      setHostSVG(document.getElementById('host0'), svg);
      loopEl.textContent = 'Single vector QR · resize freely · Esc closes';
    }

    async function paint() {
      if (frameCount <= 0) return;
      const jobs = [];
      for (let s = 0; s < streams; s++) {
        const idx = frameIndex(tickN, s, streams, frameCount);
        const host = document.getElementById('host' + s);
        if (!host) continue;
        jobs.push(svgAt(idx).then(svg => setHostSVG(host, svg)));
      }
      await Promise.all(jobs);
      tickN += 1;
      if (tickN %% frameCount === 0) loops += 1;
      const mode = streams > 1 ? 'LEFT+RIGHT SVG' : 'single SVG';
      loopEl.textContent = mode + ' · loop ' + (loops + 1) + ' · crisp vector';
    }

    function showAnimated(count, fps, streamCount, inline) {
      frameCount = count;
      inlineFrames = inline || [];
      svgCache = {};
      tickN = 0;
      loops = 0;
      streams = Math.max(1, Math.min(streamCount || 1, 4));
      if (frameCount < 2) streams = 1;
      stage.classList.toggle('multi', streams > 1);
      const labels = ['LEFT', 'RIGHT', 'L2', 'R2'];
      let html = '';
      for (let s = 0; s < streams; s++) {
        html += '<div class="slot">' +
          (streams > 1 ? '<label>' + (labels[s] || ('S' + s)) + '</label>' : '') +
          '<div class="qr-host" id="host' + s + '"></div></div>';
      }
      stage.innerHTML = html;
      const ms = Math.max(90, Math.round(1000 / (fps || 6)));
      if (timer) clearInterval(timer);
      paint();
      timer = setInterval(() => { paint(); }, ms);
    }

    async function load() {
      try {
        const res = await fetch('/api/latest?format=frames');
        const data = await res.json();
        if (data.error && !data.image && !(data.frames && data.frames.length) && !data.frame_count) {
          stage.innerHTML = '<div class="err">' + data.error + '</div>';
          return;
        }
        const sc = data.streams || 1;
        const kind = data.static
          ? 'single vector QR'
          : (data.frame_count + ' SVG frames @ ' + data.fps + ' fps · ' + (sc > 1 ? 'LEFT+RIGHT' : 'single'));
        stats.textContent = data.bytes + ' bytes · ' + kind + ' · high contrast';
        if (data.static || data.frame_count <= 1) {
          const svg = (data.frames && data.frames[0]) || data.image;
          await showStatic(svg);
        } else {
          showAnimated(data.frame_count, data.fps, sc, data.frames || []);
        }
      } catch (err) {
        stage.innerHTML = '<div class="err">' + (err.message || err) + '</div>';
      }
    }
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') window.close();
    });
    window.addEventListener('resize', () => {
      // SVG is viewBox-based; hosts already use %% width — force a repaint for safety.
      if (frameCount > 0) paint();
    });
    load();
  </script>
</body>
</html>
`
