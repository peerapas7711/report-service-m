package httpserver

import "github.com/gofiber/fiber/v2"

func payslipPage(c *fiber.Ctx) error {
	c.Type("html", "utf-8")
	return c.SendString(payslipPageHTML)
}

const payslipPageHTML = `<!doctype html>
<html lang="th">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Payslip</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f4f7f6;
      --surface: #ffffff;
      --panel: #eef5f3;
      --line: #d8e2df;
      --text: #13201d;
      --muted: #60706b;
      --primary: #0b6f63;
      --primary-dark: #08564d;
      --accent: #be7c2d;
      --shadow: 0 18px 48px rgba(21, 41, 36, .12);
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      min-height: 100vh;
      background: var(--bg);
      color: var(--text);
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      font-size: 14px;
      letter-spacing: 0;
    }

    .shell {
      min-height: 100vh;
      display: grid;
      grid-template-columns: minmax(300px, 380px) minmax(0, 1fr);
    }

    aside {
      background: var(--surface);
      border-right: 1px solid var(--line);
      padding: 24px;
      display: flex;
      flex-direction: column;
      gap: 22px;
    }

    main {
      padding: 24px;
      min-width: 0;
    }

    header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 14px;
    }

    h1 {
      margin: 0;
      font-size: 24px;
      line-height: 1.15;
      font-weight: 760;
    }

    .badge {
      color: var(--primary-dark);
      background: #dcefed;
      border: 1px solid #badbd6;
      border-radius: 999px;
      padding: 5px 10px;
      font-size: 12px;
      font-weight: 680;
      white-space: nowrap;
    }

    form {
      display: grid;
      gap: 16px;
    }

    fieldset {
      border: 0;
      margin: 0;
      padding: 0;
      display: grid;
      gap: 12px;
    }

    legend {
      padding: 0 0 2px;
      color: var(--muted);
      font-size: 12px;
      font-weight: 720;
      text-transform: uppercase;
    }

    label {
      display: grid;
      gap: 7px;
      color: var(--muted);
      font-size: 12px;
      font-weight: 680;
    }

    select,
    input {
      width: 100%;
      min-height: 42px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
      color: var(--text);
      padding: 0 12px;
      font: inherit;
      font-size: 14px;
      outline: none;
    }

    select:focus,
    input:focus {
      border-color: var(--primary);
      box-shadow: 0 0 0 3px rgba(11, 111, 99, .14);
    }

    .segmented {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 8px;
    }

    .segmented input {
      position: absolute;
      opacity: 0;
      pointer-events: none;
    }

    .segmented label {
      min-height: 42px;
      display: grid;
      place-items: center;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
      color: var(--text);
      cursor: pointer;
      font-size: 14px;
    }

    .segmented input:checked + span {
      width: 100%;
      height: 100%;
      display: grid;
      place-items: center;
      border-radius: 7px;
      background: var(--panel);
      color: var(--primary-dark);
      box-shadow: inset 0 0 0 1px #9ecac3;
    }

    .actions {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 10px;
      padding-top: 4px;
    }

    button {
      min-height: 44px;
      border: 0;
      border-radius: 8px;
      padding: 0 14px;
      font: inherit;
      font-weight: 760;
      cursor: pointer;
    }

    .primary {
      background: var(--primary);
      color: #fff;
    }

    .primary:hover {
      background: var(--primary-dark);
    }

    .secondary {
      background: #fff;
      color: var(--text);
      border: 1px solid var(--line);
    }

    .secondary:hover {
      border-color: #b4c5c1;
      background: #f9fbfa;
    }

    .preview-shell {
      min-height: calc(100vh - 48px);
      display: grid;
      grid-template-rows: auto minmax(420px, 1fr);
      gap: 16px;
    }

    .preview-top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
    }

    .preview-title {
      display: flex;
      align-items: baseline;
      gap: 10px;
      min-width: 0;
    }

    .preview-title h2 {
      margin: 0;
      font-size: 18px;
      line-height: 1.2;
    }

    .preview-title span {
      color: var(--muted);
      white-space: nowrap;
    }

    .open-link {
      color: var(--primary-dark);
      text-decoration: none;
      font-weight: 760;
      white-space: nowrap;
    }

    .viewer {
      overflow: hidden;
      background: var(--surface);
      border: 1px solid var(--line);
      border-radius: 8px;
      box-shadow: var(--shadow);
    }

    iframe {
      width: 100%;
      height: 100%;
      min-height: 640px;
      border: 0;
      display: block;
      background: #fff;
    }

    @media (max-width: 900px) {
      .shell {
        grid-template-columns: 1fr;
      }

      aside {
        border-right: 0;
        border-bottom: 1px solid var(--line);
      }

      main {
        padding: 18px;
      }

      .preview-shell {
        min-height: auto;
      }
    }

    @media (max-width: 520px) {
      aside {
        padding: 18px;
      }

      .actions,
      .segmented {
        grid-template-columns: 1fr;
      }

      .preview-top {
        align-items: flex-start;
        flex-direction: column;
      }
    }
  </style>
</head>
<body>
  <div class="shell">
    <aside>
      <header>
        <h1>Payslip</h1>
        <span class="badge">PDF</span>
      </header>

      <form id="payslipForm" method="get" action="/preview/payslip" target="payslipPreview">
        <fieldset>
          <legend>Source</legend>
          <label>
            Company
            <select name="mock" id="mock">
              <option value="hopinn">Hop Inn</option>
              <option value="tigersoft">TigerSoft</option>
              <option value="bluewave">Bluewave</option>
              <option value="kubota">Kubota</option>
            </select>
          </label>

        </fieldset>

        <fieldset>
          <legend>Page</legend>
          <div class="segmented" role="radiogroup" aria-label="Orientation">
            <label>
              <input type="radio" name="orientation" value="P" checked>
              <span>Portrait</span>
            </label>
            <label>
              <input type="radio" name="orientation" value="L">
              <span>Landscape</span>
            </label>
          </div>
        </fieldset>

        <fieldset>
          <legend>Override</legend>
          <label>
            Company name
            <input name="company_name" id="companyName" placeholder="Company name">
          </label>

          <label>
            Logo URL
            <input name="logo" id="logo" placeholder="https://example.com/logo.png">
          </label>
        </fieldset>

        <input type="hidden" name="download" id="download" value="">

        <div class="actions">
          <button class="primary" type="submit" data-mode="preview">Preview</button>
          <button class="secondary" type="submit" data-mode="download">Download</button>
        </div>
      </form>
    </aside>

    <main>
      <section class="preview-shell" aria-label="Payslip preview">
        <div class="preview-top">
          <div class="preview-title">
            <h2 id="previewName">Hop Inn</h2>
            <span id="previewMeta">Portrait</span>
          </div>
          <a class="open-link" id="openLink" href="/preview/payslip?mock=hopinn&orientation=P" target="_blank" rel="noreferrer">Open PDF</a>
        </div>

        <div class="viewer">
          <iframe id="payslipPreview" name="payslipPreview" title="Payslip PDF" src="/preview/payslip?mock=hopinn&orientation=P"></iframe>
        </div>
      </section>
    </main>
  </div>

  <script>
    const form = document.getElementById('payslipForm');
    const download = document.getElementById('download');
    const openLink = document.getElementById('openLink');
    const previewName = document.getElementById('previewName');
    const previewMeta = document.getElementById('previewMeta');
    const labels = {
      hopinn: 'Hop Inn',
      tigersoft: 'TigerSoft',
      bluewave: 'Bluewave',
      kubota: 'Kubota'
    };
    function currentUrl(includeDownload) {
      const data = new FormData(form);
      if (!includeDownload) {
        data.delete('download');
      } else {
        data.set('download', '1');
      }

      for (const [key, value] of Array.from(data.entries())) {
        if (typeof value === 'string' && value.trim() === '') {
          data.delete(key);
        }
      }

      return '/preview/payslip?' + new URLSearchParams(data).toString();
    }

    function syncMeta() {
      const data = new FormData(form);
      const mock = data.get('mock') || 'hopinn';
      const orientation = data.get('orientation') === 'L' ? 'Landscape' : 'Portrait';
      previewName.textContent = labels[mock] || mock;
      previewMeta.textContent = orientation;
      openLink.href = currentUrl(false);
    }

    form.addEventListener('submit', (event) => {
      const mode = event.submitter?.dataset.mode || 'preview';
      download.value = mode === 'download' ? '1' : '';
      if (mode === 'download') {
        form.target = '_blank';
      } else {
        form.target = 'payslipPreview';
      }
      syncMeta();
    });

    form.addEventListener('change', syncMeta);
    form.addEventListener('input', syncMeta);
    syncMeta();
  </script>
</body>
</html>`
