package httpapp

const baseStyles = `
<style>
  @import url('https://fonts.googleapis.com/css2?family=Roboto+Narrow:wght@400;500;600;700&display=swap');
  :root {
    color-scheme: light;
    --bg: #f4efe7;
    --surface: #fffdf9;
    --ink: #1f2933;
    --muted: #52606d;
    --accent: #b35c2e;
    --border: #e7d8c8;
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    font-family: "Roboto Narrow", Roboto, "Helvetica Neue", Arial, sans-serif;
    color: var(--ink);
    background:
      radial-gradient(circle at top left, rgba(179,92,46,0.10), transparent 32%),
      linear-gradient(180deg, #f7f1e8 0%, #f4efe7 100%);
  }
  main {
    max-width: 760px;
    margin: 3rem auto;
    padding: 0 1rem;
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 20px;
    padding: 1.4rem;
    box-shadow: 0 12px 40px rgba(31,41,51,0.08);
  }
  h1, h2, h3 { margin-top: 0; }
  p { line-height: 1.55; }
  .muted { color: var(--muted); }
  .stack > * + * { margin-top: 1rem; }
  .button, button {
    display: inline-block;
    border: 0;
    border-radius: 999px;
    background: var(--accent);
    color: white;
    padding: 0.8rem 1.1rem;
    text-decoration: none;
    cursor: pointer;
    font-family: inherit;
    font-weight: 600;
  }
  input {
    width: 100%;
    padding: 0.8rem 0.9rem;
    border-radius: 12px;
    border: 1px solid var(--border);
    background: white;
  }
  pre {
    white-space: pre-wrap;
    background: #f8f5f1;
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 1rem;
    overflow-x: auto;
  }
  .text-shell {
    position: relative;
  }
  .text-actions {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 0.75rem;
  }
  .text-code {
    font-family: "Roboto Mono", "SFMono-Regular", Consolas, "Liberation Mono", monospace;
    font-weight: 700;
    font-size: 1rem;
    line-height: 1.65;
  }
  ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  li + li {
    margin-top: 0.8rem;
  }
  .row {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
  }
  .avatar {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid var(--border);
    background: #f0ebe5;
  }
  .pill {
    display: inline-block;
    font-size: 0.85rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 0.2rem 0.7rem;
    color: var(--muted);
  }
  .error {
    color: #8f2d1c;
    font-weight: 600;
  }
</style>
`

const unlockPageTemplate = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>` + baseStyles + `
</head>
<body>
  <main>
    <section class="card stack">
      <div class="pill">Geschuetzter Bereich</div>
      <h1>{{ .Title }}</h1>
      <p class="muted">{{ .Description }}</p>
      {{ if .Error }}<p class="error">{{ .Error }}</p>{{ end }}
      <form method="post" class="stack">
        <label for="passphrase">Passphrase</label>
        <input id="passphrase" name="passphrase" type="password" autofocus>
        <button type="submit">Freischalten</button>
      </form>
    </section>
  </main>
</body>
</html>`

const textPageTemplate = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>` + baseStyles + `
</head>
<body>
  <main>
    <section class="card stack">
      <div class="pill">Text</div>
      <h1>{{ .Title }}</h1>
      <p class="muted">{{ .Description }}</p>
      <div class="text-shell">
        <div class="text-actions">
          <button type="button" id="copy-button">Kopieren</button>
        </div>
        <pre id="text-content" class="text-code">{{ .Content }}</pre>
      </div>
      <p class="muted">{{ .CopyHint }}</p>
    </section>
  </main>
  <script>
    (function () {
      var button = document.getElementById('copy-button');
      var text = document.getElementById('text-content');
      if (!button || !text || !navigator.clipboard) {
        return;
      }
      var originalLabel = button.textContent;
      button.addEventListener('click', function () {
        navigator.clipboard.writeText(text.textContent || '').then(function () {
          button.textContent = 'Kopiert';
          window.setTimeout(function () {
            button.textContent = originalLabel;
          }, 1600);
        });
      });
    }());
  </script>
</body>
</html>`

const vcardPageTemplate = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>` + baseStyles + `
</head>
<body>
  <main>
    <section class="card stack">
      <div class="row">
        {{ if .Entry.ImageURL }}<img class="avatar" src="{{ .Entry.ImageURL }}" alt="{{ .Entry.FullName }}">{{ end }}
        <div class="stack">
          <div class="pill">Digitale Visitenkarte</div>
          <h1>{{ .Entry.FullName }}</h1>
          <p class="muted">{{ .Entry.JobTitle }} bei {{ .Entry.Organization }}</p>
        </div>
      </div>
      <p>{{ .Description }}</p>
      <ul>
        <li><strong>E-Mail:</strong> {{ .Entry.Email }}</li>
        <li><strong>Mobil:</strong> {{ .Entry.PhoneMobile }}</li>
        <li><strong>Adresse:</strong> {{ .Entry.Address }}</li>
        <li><strong>Website:</strong> <a href="{{ .Entry.Website }}">{{ .Entry.Website }}</a></li>
      </ul>
      <p>{{ .Entry.Note }}</p>
      <p><a class="button" href="{{ .DownloadURL }}">VCF herunterladen</a></p>
    </section>
  </main>
</body>
</html>`

const listPageTemplate = `<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>` + baseStyles + `
</head>
<body>
  <main>
    <section class="card stack">
      <div class="pill">Linkliste</div>
      <h1>{{ .Title }}</h1>
      <p class="muted">{{ .Description }}</p>
      <ul>
        {{ range .Items }}
        <li class="card">
          <h3>{{ .Label }}</h3>
          <p class="muted">{{ .Description }}</p>
          <p><span class="pill">{{ .Category }}</span></p>
          <p><a class="button" href="{{ .URL }}">Oeffnen</a></p>
        </li>
        {{ end }}
      </ul>
    </section>
  </main>
</body>
</html>`
