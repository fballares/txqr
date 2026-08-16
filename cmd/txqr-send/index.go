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
      gap: 8px;
      place-items: center;
      align-content: center;
      padding: 8px 10px 4px;
      background: #fff;
    }
    .stage.multi {
      grid-template-columns: 1fr 1fr;
      column-gap: 12px;
      padding-left: 12px;
      padding-right: 12px;
    }
    .slot {
      display: grid;
      gap: 4px;
      justify-items: center;
    }
    .slot label {
      font-size: 0.7rem;
      font-weight: 700;
      letter-spacing: 0.08em;
      color: #444;
      text-transform: uppercase;
    }
    .stage img {
      width: min(280px, 42vw);
      height: auto;
      image-rendering: pixelated;
      background: #fff;
    }
    .stage:not(.multi) img {
      width: min(360px, 88vw);
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
    <strong>Looping continuously</strong> —
    dual mode uses <strong>LEFT</strong> + <strong>RIGHT</strong> QR slots; iPhone reads both sides.
    Drag this window to your phone stand.
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
    let frames = [];
    let tickN = 0;
    let loops = 0;
    let streams = 1;
    let timer = null;

    function frameIndex(tick, stream, streamCount, n) {
      if (n <= 0) return 0;
      const offset = Math.floor((n * stream) / streamCount);
      return (tick + offset) %% n;
    }

    function showStatic(src) {
      stage.classList.remove('multi');
      stage.innerHTML = '<img alt="TXQR" src="' + src + '" />';
      loopEl.textContent = 'Single QR · Esc closes';
    }

    function paint() {
      if (!frames.length) return;
      for (let s = 0; s < streams; s++) {
        const img = document.getElementById('qr' + s);
        if (!img) continue;
        const idx = frameIndex(tickN, s, streams, frames.length);
        img.src = frames[idx];
      }
      tickN += 1;
      if (frames.length && tickN %% frames.length === 0) loops += 1;
      const mode = streams > 1 ? 'LEFT+RIGHT QRs' : 'single QR';
      loopEl.textContent = mode + ' · loop ' + (loops + 1);
    }

    function showAnimated(list, fps, streamCount) {
      frames = list;
      tickN = 0;
      loops = 0;
      // Trust server auto-selection; never show dual for a 1-frame set.
      streams = Math.max(1, Math.min(streamCount || 1, 4));
      if (frames.length < 2) streams = 1;
      stage.classList.toggle('multi', streams > 1);
      // Stream 0 = LEFT, stream 1 = RIGHT (stable placement for the phone).
      const labels = ['LEFT', 'RIGHT', 'L2', 'R2'];
      let html = '';
      for (let s = 0; s < streams; s++) {
        html += '<div class="slot">' +
          (streams > 1 ? '<label>' + (labels[s] || ('S' + s)) + '</label>' : '') +
          '<img id="qr' + s + '" alt="TXQR ' + (labels[s] || s) + '" src="' +
          frames[frameIndex(0, s, streams, frames.length)] + '" /></div>';
      }
      stage.innerHTML = html;
      const ms = Math.max(80, Math.round(1000 / (fps || 6)));
      if (timer) clearInterval(timer);
      timer = setInterval(paint, ms);
      paint();
    }

    async function load() {
      try {
        const res = await fetch('/api/latest?format=frames');
        const data = await res.json();
        if (data.error && !data.image && !(data.frames && data.frames.length)) {
          stage.innerHTML = '<div class="err">' + data.error + '</div>';
          return;
        }
        const sc = data.streams || 1;
        const kind = data.static
          ? 'single static QR'
          : (data.frame_count + ' frames @ ' + data.fps + ' fps · ' + (sc > 1 ? (sc + ' concurrent') : 'single'));
        stats.textContent = data.bytes + ' bytes · ' + kind;
        if (data.static || !data.frames || data.frames.length <= 1) {
          showStatic(data.image || (data.frames && data.frames[0]));
        } else if (data.frames && data.frames.length) {
          showAnimated(data.frames, data.fps, sc);
        } else {
          stage.classList.remove('multi');
          stage.innerHTML = '<img alt="TXQR stream" src="' + data.image + '" />';
          loopEl.textContent = 'GIF looping · Esc closes';
        }
      } catch (err) {
        stage.innerHTML = '<div class="err">' + (err.message || err) + '</div>';
      }
    }
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') window.close();
    });
    load();
  </script>
</body>
</html>
`
