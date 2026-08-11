package reporter

import (
	"fmt"
	"html/template"
	"io"
)

var htmlReportTemplate = template.Must(template.New("report").Funcs(template.FuncMap{
	"number":  formatHTMLNumber,
	"display": displayString,
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    :root {
      color-scheme: light dark;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      color: #172033;
      background: #f3f6fb;
    }
    * { box-sizing: border-box; }
    body { margin: 0; background: #f3f6fb; }
    main { width: min(960px, calc(100% - 32px)); margin: 0 auto; padding: 48px 0; }
    header { margin-bottom: 24px; }
    h1 { margin: 0; font-size: clamp(1.75rem, 4vw, 2.5rem); letter-spacing: -0.04em; }
    .measured-at { margin: 8px 0 0; color: #65718a; font-size: 0.9rem; }
    h2 { margin: 0; font-size: 0.8rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; color: #65718a; }
    .speed-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
    .card, .panel { background: #fff; border: 1px solid #dde3ed; border-radius: 16px; box-shadow: 0 8px 24px rgba(26, 39, 64, 0.06); }
    .speed-card { position: relative; overflow: hidden; padding: 24px; }
    .speed-card::before { position: absolute; inset: 0 0 auto; height: 5px; content: ""; }
    .download::before { background: #0f9f8f; }
    .upload::before { background: #6366f1; }
    .speed-value { margin: 22px 0 0; font-variant-numeric: tabular-nums; }
    .speed-value strong { font-size: clamp(2.75rem, 8vw, 4.75rem); line-height: 0.9; letter-spacing: -0.06em; }
    .speed-value span { margin-left: 8px; color: #65718a; font-weight: 650; }
    .panel { margin-top: 16px; padding: 24px; }
    .metric-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 12px; margin-top: 18px; }
    .metric { padding: 16px; border-radius: 12px; background: #f3f6fb; }
    .metric dt { color: #65718a; font-size: 0.78rem; font-weight: 650; }
    .metric dd { margin: 8px 0 0; font-size: 1.35rem; font-weight: 750; font-variant-numeric: tabular-nums; }
    .metric dd span { color: #65718a; font-size: 0.75rem; font-weight: 650; }
    .network { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px 24px; margin: 18px 0 0; }
    .network div { min-width: 0; }
    .network dt { color: #65718a; font-size: 0.78rem; font-weight: 650; }
    .network dd { overflow-wrap: anywhere; margin: 4px 0 0; font-weight: 650; }
    .warnings { border-color: #f1c36d; background: #fffaf0; }
    .warnings ul { margin: 16px 0 0; padding-left: 20px; }
    .warnings li + li { margin-top: 8px; }
    @media (max-width: 720px) {
      main { padding: 28px 0; }
      .speed-grid { grid-template-columns: 1fr; }
      .metric-grid { grid-template-columns: repeat(2, 1fr); }
      .network { grid-template-columns: 1fr; }
    }
    @media (prefers-color-scheme: dark) {
      :root { color: #edf2fa; background: #101522; }
      body { background: #101522; }
      h2, .measured-at, .speed-value span, .metric dt, .metric dd span, .network dt { color: #9da9bf; }
      .card, .panel { background: #171e2e; border-color: #2a3448; box-shadow: none; }
      .metric { background: #101522; }
      .warnings { background: #2a2213; border-color: #765b27; }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <h1>{{.Title}}</h1>
      <p class="measured-at">Measured <time data-unix-ms="{{.MeasuredAtUnixMs}}">{{.MeasuredAtUnixMs}}</time></p>
    </header>

    <section class="speed-grid" aria-label="Transfer speed">
      <article class="card speed-card download">
        <h2>Download</h2>
        <p class="speed-value"><strong>{{number .DownloadMbps 2}}</strong><span>Mbps</span></p>
      </article>
      <article class="card speed-card upload">
        <h2>Upload</h2>
        <p class="speed-value"><strong>{{number .UploadMbps 2}}</strong><span>Mbps</span></p>
      </article>
    </section>

    <section class="panel" aria-labelledby="quality-heading">
      <h2 id="quality-heading">Connection quality</h2>
      <dl class="metric-grid">
        <div class="metric"><dt>Unloaded latency</dt><dd>{{number .UnloadedLatency 2}} <span>ms</span></dd></div>
        <div class="metric"><dt>Download latency</dt><dd>{{number .LoadedDownLatency 2}} <span>ms</span></dd></div>
        <div class="metric"><dt>Upload latency</dt><dd>{{number .LoadedUpLatency 2}} <span>ms</span></dd></div>
        <div class="metric"><dt>Jitter</dt><dd>{{number .Jitter 2}} <span>ms</span></dd></div>
        <div class="metric"><dt>Packet loss</dt><dd>{{number .PacketLoss 1}} <span>%</span></dd></div>
      </dl>
    </section>

    <section class="panel" aria-labelledby="network-heading">
      <h2 id="network-heading">Network</h2>
      <dl class="network">
        <div><dt>Server location</dt><dd>{{display .ServerColo}}</dd></div>
        <div><dt>ASN</dt><dd>{{display .NetworkASN}}</dd></div>
        <div><dt>Provider</dt><dd>{{display .NetworkASOrg}}</dd></div>
        <div><dt>IP address</dt><dd>{{display .IP}}</dd></div>
      </dl>
    </section>

    {{if .Warnings}}
    <section class="panel warnings" aria-labelledby="warnings-heading">
      <h2 id="warnings-heading">Warnings</h2>
      <ul>{{range .Warnings}}<li>{{.}}</li>{{end}}</ul>
    </section>
    {{end}}
  </main>
  <script>
    const measuredAt = document.querySelector("[data-unix-ms]");
    const measuredAtDate = new Date(Number(measuredAt.dataset.unixMs));
    measuredAt.dateTime = measuredAtDate.toISOString();
    measuredAt.textContent = measuredAtDate.toLocaleString();
  </script>
</body>
</html>
`))

type htmlReportData struct {
	Result
	Title string
}

// PrintHTML writes a self-contained HTML report to w.
func PrintHTML(w io.Writer, r Result, title string) error {
	if title == "" {
		title = "Internet Speed Report"
	} else {
		title = "Internet Speed Report - " + title
	}
	return htmlReportTemplate.Execute(w, htmlReportData{Result: r, Title: title})
}

func formatHTMLNumber(value *float64, precision int) string {
	if value == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.*f", precision, *value)
}
