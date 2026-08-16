package main

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>TXQR Send</title>
  <style>
    :root {
      --bg0: #0f1419;
      --bg1: #1a2332;
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
        radial-gradient(1200px 600px at 10%% -10%%, #1e3a5f 0%%, transparent 55%%),
        radial-gradient(900px 500px at 100%% 0%%, #143028 0%%, transparent 50%%),
        linear-gradient(160deg, var(--bg0), var(--bg1));
      color: var(--ink);
      font-family: "Segoe UI", "Helvetica Neue", sans-serif;
    }
    body {
      display: grid;
      place-items: center;
      padding: 24px;
    }
    .shell {
      width: min(920px, 100%%);
      display: grid;
      gap: 20px;
    }
    header h1 {
      margin: 0;
      font-size: clamp(2rem, 4vw, 2.8rem);
      letter-spacing: -0.03em;
      font-weight: 700;
    }
    header p {
      margin: 8px 0 0;
      color: var(--muted);
      max-width: 42rem;
      line-height: 1.45;
    }
    .panel {
      display: grid;
      gap: 14px;
      padding: 18px;
      border: 1px solid var(--line);
      background: rgba(8, 12, 18, 0.45);
      backdrop-filter: blur(8px);
    }
    label {
      display: grid;
      gap: 6px;
      font-size: 0.85rem;
      color: var(--muted);
    }
    textarea, input {
      width: 100%%;
      border: 1px solid var(--line);
      background: rgba(255,255,255,0.03);
      color: var(--ink);
      border-radius: 0;
      padding: 12px 14px;
      font: inherit;
    }
    textarea {
      min-height: 160px;
      resize: vertical;
      line-height: 1.4;
    }
    .row {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;
    }
    @media (max-width: 700px) {
      .row { grid-template-columns: 1fr; }
    }
    .actions {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }
    button {
      appearance: none;
      border: 0;
      padding: 12px 18px;
      font: inherit;
      font-weight: 600;
      cursor: pointer;
      background: var(--accent);
      color: var(--accent-ink);
    }
    button.secondary {
      background: transparent;
      color: var(--ink);
      border: 1px solid var(--line);
    }
    button:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
    .stage {
      display: none;
      place-items: center;
      gap: 14px;
      padding: 20px;
      border: 1px solid var(--line);
      background: #fff;
      color: #111;
      min-height: 520px;
    }
    .stage.active { display: grid; }
    .stage img {
      width: min(520px, 90vw);
      height: auto;
      image-rendering: pixelated;
    }
    .meta {
      color: var(--muted);
      font-size: 0.9rem;
    }
    .meta strong { color: var(--ok); font-weight: 600; }
    .error {
      color: #ff8e8e;
      min-height: 1.2em;
    }
    .sending header, .sending .composer { display: none; }
    .sending .stage { display: grid; }
  </style>
</head>
<body>
  <div class="shell" id="app">
    <header>
      <h1>TXQR Send</h1>
      <p>Paste text on this Windows machine, show the animated QR stream, and let your phone reader decode and copy it.</p>
    </header>

    <section class="panel composer">
      <label>
        Text to transfer
        <textarea id="text" placeholder="Paste selected text here…">%s</textarea>
      </label>
      <div class="row">
        <label>Chunk size
          <input id="chunk" type="number" min="20" max="1000" value="%d" />
        </label>
        <label>FPS
          <input id="fps" type="number" min="1" max="20" value="%d" />
        </label>
        <label>QR size
          <input id="size" type="number" min="200" max="800" value="%d" />
        </label>
      </div>
      <div class="actions">
        <button id="send" type="button">Show QR stream</button>
        <button id="paste" class="secondary" type="button">Paste from clipboard</button>
      </div>
      <div class="error" id="error"></div>
      <div class="meta" id="hint">Tip: keep the window full-screen and steady while scanning.</div>
    </section>

    <section class="stage" id="stage">
      <img id="qr" alt="Animated TXQR stream" />
      <div class="meta" id="stats"></div>
      <div class="actions">
        <button id="back" class="secondary" type="button">Edit text</button>
      </div>
    </section>
  </div>
  <script>
    const app = document.getElementById('app');
    const textEl = document.getElementById('text');
    const errorEl = document.getElementById('error');
    const stage = document.getElementById('stage');
    const qr = document.getElementById('qr');
    const stats = document.getElementById('stats');

    async function pasteClipboard() {
      errorEl.textContent = '';
      try {
        const value = await navigator.clipboard.readText();
        textEl.value = value;
      } catch (err) {
        errorEl.textContent = 'Clipboard access blocked — paste with Ctrl+V instead.';
      }
    }

    async function send() {
      errorEl.textContent = '';
      const text = textEl.value;
      if (!text.trim()) {
        errorEl.textContent = 'Enter or paste some text first.';
        return;
      }
      const body = {
        text,
        chunk_len: Number(document.getElementById('chunk').value),
        fps: Number(document.getElementById('fps').value),
        qr_size: Number(document.getElementById('size').value),
        format: 'gif',
        redundancy: 2.0
      };
      document.getElementById('send').disabled = true;
      try {
        const res = await fetch('/api/encode', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        if (!res.ok || data.error) {
          throw new Error(data.error || ('HTTP ' + res.status));
        }
        qr.src = data.gif;
        stats.innerHTML = '<strong>' + data.bytes + ' bytes</strong> · ' +
          data.frame_count + ' frames · ' + data.fps + ' fps loop';
        app.classList.add('sending');
        stage.classList.add('active');
      } catch (err) {
        errorEl.textContent = err.message || String(err);
      } finally {
        document.getElementById('send').disabled = false;
      }
    }

    document.getElementById('paste').addEventListener('click', pasteClipboard);
    document.getElementById('send').addEventListener('click', send);
    document.getElementById('back').addEventListener('click', () => {
      app.classList.remove('sending');
      stage.classList.remove('active');
      qr.removeAttribute('src');
    });

    if (textEl.value.trim()) {
      send();
    }
  </script>
</body>
</html>
`
