// main.go
// Go + HTMX Environment/App Matrix — modern UI, external source sync, and drift detection.
//
// Features
// - Modern, responsive UI (pure CSS, no build tooling).
// - Shows Apps × Environments with: Version, Image, Updated time.
// - External sync endpoint (generic Fetcher interface). Includes a demo fetcher and an optional HTTP JSON fetcher
//   controlled via env var FETCH_URL. See "External Integrations" notes at the bottom.
// - Drift detection per App: highlights environments that are behind the highest version.
// - Filters (app/env/version contains), CSV export, periodic auto-refresh.
//
// Run
//   go mod init env-matrix && go mod tidy
//   go run .
//   open http://localhost:8080
//
// Optional: set FETCH_URL to an HTTP endpoint returning JSON records:
//   [ {"env":"prod","app":"billing","image":"registry.example.com/billing:1.9.5","version":"1.9.5"}, ... ]

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ==========================
// Data Model & In-Memory DB
// ==========================

type Store struct {
	mu      sync.RWMutex
	apps    map[string]struct{}        // set of apps
	envs    map[string]struct{}        // set of envs
	matrix  map[string]map[string]Cell // env -> app -> Cell
	updated time.Time
}

type Cell struct {
	Version string
	Image   string
	Updated time.Time
	Drift   bool // true if behind the max version for this app
	Latest  bool // true if this is the max version for this app
}

func NewStore() *Store {
	return &Store{apps: map[string]struct{}{}, envs: map[string]struct{}{}, matrix: map[string]map[string]Cell{}}
}

func (s *Store) ensureEnv(env string) {
	if _, ok := s.envs[env]; !ok {
		s.envs[env] = struct{}{}
		if _, ok := s.matrix[env]; !ok {
			s.matrix[env] = map[string]Cell{}
		}
	}
}
func (s *Store) ensureApp(app string) {
	if _, ok := s.apps[app]; !ok {
		s.apps[app] = struct{}{}
	}
}

func (s *Store) Set(env, app, version, image string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env, app, version, image = strings.TrimSpace(env), strings.TrimSpace(app), strings.TrimSpace(version), strings.TrimSpace(image)
	if env == "" || app == "" {
		return
	}
	s.ensureEnv(env)
	s.ensureApp(app)
	row := s.matrix[env]
	row[app] = Cell{Version: version, Image: image, Updated: time.Now()}
	s.updated = time.Now()
}

func (s *Store) AddEnv(env string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	env = strings.TrimSpace(env)
	if env == "" {
		return
	}
	s.ensureEnv(env)
	s.updated = time.Now()
}
func (s *Store) AddApp(app string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app = strings.TrimSpace(app)
	if app == "" {
		return
	}
	s.ensureApp(app)
	s.updated = time.Now()
}
func (s *Store) LatestUpdated() time.Time { s.mu.RLock(); defer s.mu.RUnlock(); return s.updated }

func (s *Store) Snapshot(filterApp, filterEnv, contains string) (apps []string, envs []string, m map[string]map[string]Cell) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	apps = make([]string, 0, len(s.apps))
	for a := range s.apps {
		if filterApp == "" || containsFold(a, filterApp) {
			apps = append(apps, a)
		}
	}
	sort.Strings(apps)
	envs = make([]string, 0, len(s.envs))
	for e := range s.envs {
		if filterEnv == "" || containsFold(e, filterEnv) {
			envs = append(envs, e)
		}
	}
	sort.Strings(envs)
	m = map[string]map[string]Cell{}
	for _, e := range envs {
		row := map[string]Cell{}
		for _, a := range apps {
			c := s.matrix[e][a]
			if contains == "" || containsFold(c.Version, contains) || containsFold(c.Image, contains) {
				row[a] = c
			}
		}
		m[e] = row
	}
	return
}

func (s *Store) ToCSV() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	apps := keysSorted(s.apps)
	envs := keysSorted(s.envs)
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	head := append([]string{"App/Env"}, envs...)
	_ = w.Write(head)
	for _, a := range apps {
		row := []string{a}
		for _, e := range envs {
			c := s.matrix[e][a]
			row = append(row, fmt.Sprintf("%s (%s)", c.Version, c.Image))
		}
		_ = w.Write(row)
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// Drift calculation — compute max version per app and mark cells.
func (s *Store) RecomputeDrift() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// find max per app
	maxByApp := map[string]semver{}
	for env, row := range s.matrix {
		_ = env
		for app, c := range row {
			v := parseSemver(c.Version)
			if cur, ok := maxByApp[app]; !ok || v.gt(cur) {
				maxByApp[app] = v
			}
		}
	}
	// mark cells
	for env, row := range s.matrix {
		for app, c := range row {
			v := parseSemver(c.Version)
			max := maxByApp[app]
			c.Drift = v.valid && max.valid && v.lt(max)
			c.Latest = v.valid && max.valid && v.eq(max) && c.Version != ""
			row[app] = c
		}
		s.matrix[env] = row
	}
}

// ==========================
// External Fetcher (generic)
// ==========================

type Record struct{ Env, App, Image, Version string }

type Fetcher interface{ Fetch() ([]Record, error) }

// Demo fetcher (replace later)
type DemoFetcher struct{}

func (DemoFetcher) Fetch() ([]Record, error) {
	return []Record{
		{"dev", "auth", "registry/acme/auth:1.4.0", "1.4.0"},
		{"test", "auth", "registry/acme/auth:1.3.2", "1.3.2"},
		{"staging", "auth", "registry/acme/auth:1.3.2", "1.3.2"},
		{"prod", "auth", "registry/acme/auth:1.3.1", "1.3.1"},
		{"dev", "billing", "registry/acme/billing:2.0.0-rc2", "2.0.0-rc2"},
		{"prod", "billing", "registry/acme/billing:1.9.5", "1.9.5"},
		{"dev", "orders", "registry/acme/orders:0.9.2", "0.9.2"},
		{"test", "orders", "registry/acme/orders:0.9.0", "0.9.0"},
		{"prod", "catalog", "registry/acme/catalog:3.4.8", "3.4.8"},
	}, nil
}

// HTTP JSON fetcher (generic schema shown in header comment)
type HTTPFetcher struct {
	URL    string
	Client *http.Client
}

func (h HTTPFetcher) Fetch() ([]Record, error) {
	if h.URL == "" {
		return nil, errors.New("empty URL")
	}
	resp, err := h.Client.Get(h.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch %s: %s: %s", h.URL, resp.Status, string(b))
	}
	var recs []Record
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&recs); err != nil {
		return nil, err
	}
	return recs, nil
}

var fetcher Fetcher = DemoFetcher{}

// ==========================
// Templates (modern UI)
// ==========================

var baseTmpl = template.Must(template.New("base").Funcs(template.FuncMap{
	"nowISO": func() string { return time.Now().Format(time.RFC3339) },
}).Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Env/App Matrix</title>
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    :root{ --bg:#0b1020; --card:#111733; --muted:#9aa3b2; --text:#e6ecff; --accent:#7c9cff; --ok:#22c55e; --warn:#f59e0b; --bad:#ef4444; --chip:#1b2347; --chip-br:#243059; }
    *{box-sizing:border-box}
    body{margin:0; background:linear-gradient(180deg,#0b1020 0%, #0e1430 100%); color:var(--text); font-family:Inter, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif}
    header{display:flex; align-items:center; justify-content:space-between; padding:20px 28px; position:sticky; top:0; backdrop-filter:saturate(140%) blur(6px); background:rgba(11,16,32,0.7); border-bottom:1px solid #1e2648}
    h1{font-size:20px; margin:0}
    .muted{color:var(--muted); font-size:12px}
    .wrap{display:flex; gap:10px; align-items:center; flex-wrap:wrap}
    .btn{padding:8px 12px; border:1px solid #2a3568; background:linear-gradient(180deg,#1c2550,#1a2348); color:var(--text); border-radius:10px; cursor:pointer}
    .btn.primary{border-color:#2f42a4; background:linear-gradient(180deg,#2a3f9a,#253a90)}
    .btn.ghost{background:transparent}
    .container{padding:20px 28px}
    .card{background:linear-gradient(180deg,#0e1533, #0f1638); border:1px solid #263062; border-radius:16px; padding:14px; box-shadow:0 5px 20px rgba(5,10,30,0.25)}
    input[type=text]{padding:10px 12px; border:1px solid #2a3568; background:#0c1230; color:var(--text); border-radius:10px}
    table{border-collapse:separate; border-spacing:0; width:100%; font-size:14px}
    thead th{position:sticky; top:0; background:#131a3a; color:#c9d4ff; z-index:1; border-bottom:1px solid #2a3568}
    th, td{border-right:1px solid #223064; border-bottom:1px solid #223064; padding:10px; text-align:left}
    th:first-child{position:sticky; left:0; background:#131a3a; z-index:2}
    tbody th{background:#0f1638; color:#d3dbff}
    .td-cell{min-width:220px}
    .grid-wrap{overflow:auto; max-height:70vh; border:1px solid #223064; border-radius:14px}
    .version{font-weight:700}
    .image{font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; font-size:12px; color:#b8c1ff}
    .updated{color:var(--muted); font-size:11px}
    .chip{display:inline-flex; align-items:center; gap:6px; padding:3px 8px; font-size:11px; border-radius:999px; border:1px solid var(--chip-br); background:var(--chip)}
    .chip.ok{border-color:#194a2a; background:rgba(23,120,59,0.18)}
    .chip.warn{border-color:#4b3b14; background:rgba(133,95,10,0.18)}
    .chip.bad{border-color:#4b161b; background:rgba(160,32,43,0.18)}
    .row{display:flex; flex-direction:column; gap:4px}
  </style>
</head>
<body>
  <header>
    <div>
      <h1>Environment • Application Matrix</h1>
      <div class="muted">Last updated: {{ .Updated.Format "2006-01-02 15:04:05" }}</div>
    </div>
    <div class="wrap">
      <a href="/export.csv" class="btn">Export CSV</a>
      <form class="wrap" hx-post="/sync" hx-target="#matrix" hx-swap="outerHTML">
        <button type="submit" class="btn primary">Sync from Source</button>
      </form>
      <form class="wrap" hx-get="/matrix" hx-target="#matrix" hx-trigger="every 20s" hx-swap="outerHTML">
        <button type="submit" class="btn ghost">Auto-refresh 20s</button>
      </form>
    </div>
  </header>

  <div class="container">
    <section class="wrap" style="margin-bottom:12px">
      <div class="card">
        <form class="wrap" hx-get="/matrix" hx-target="#matrix" hx-include="#qApp,#qEnv,#qVer" hx-swap="outerHTML">
          <input id="qApp" name="app" type="text" placeholder="Filter apps (e.g. api)" />
          <input id="qEnv" name="env" type="text" placeholder="Filter envs (e.g. prod)" />
          <input id="qVer" name="contains" type="text" placeholder="Filter version/image contains" />
          <button class="btn" type="submit">Apply</button>
        </form>
      </div>
      <div class="card">
        <form class="wrap" hx-post="/add-env" hx-target="#matrix" hx-swap="outerHTML">
          <input type="text" name="env" placeholder="Add environment (e.g. prod)" required />
          <button class="btn primary" type="submit">Add Env</button>
        </form>
      </div>
      <div class="card">
        <form class="wrap" hx-post="/add-app" hx-target="#matrix" hx-swap="outerHTML">
          <input type="text" name="app" placeholder="Add application (e.g. billing)" required />
          <button class="btn primary" type="submit">Add App</button>
        </form>
      </div>
    </section>

    <section>
      <div class="grid-wrap" id="grid">
        {{ template "matrix" . }}
      </div>
      <div class="muted" style="margin-top:8px">Legend: <span class="chip ok">latest</span> <span class="chip warn">drift</span></div>
    </section>
  </div>
</body>
</html>`))

var matrixTmpl = template.Must(baseTmpl.New("matrix").Parse(`
<table id="matrix">
  <thead>
    <tr>
      <th>Application / Environment</th>
      {{ range .Envs }}<th>{{ . }}</th>{{ end }}
    </tr>
  </thead>
  <tbody>
    {{ range .Apps }}
      {{ $app := . }}
      <tr>
        <th>{{ $app }}</th>
        {{ range $.Envs }}
          {{ $env := . }}
          {{ $cell := (index (index $.Matrix $env) $app) }}
          <td class="td-cell" id="cell-{{ $env }}-{{ $app }}">
            <div class="row">
              <div>
                <span class="version">{{ $cell.Version }}</span>
                {{ if $cell.Latest }}<span class="chip ok">latest</span>{{ end }}
                {{ if $cell.Drift }}<span class="chip warn">drift</span>{{ end }}
              </div>
              {{ if $cell.Image }}<div class="image" title="{{ $cell.Image }}">{{ $cell.Image }}</div>{{ end }}
              {{ if $cell.Updated.IsZero }}
                <div class="updated">—</div>
              {{ else }}
                <div class="updated">Updated {{ $cell.Updated.Format "2006-01-02 15:04:05" }}</div>
              {{ end }}
            </div>
          </td>
        {{ end }}
      </tr>
    {{ else }}
      <tr><td colspan="999">No applications yet. Add one above.</td></tr>
    {{ end }}
  </tbody>
</table>
`))

// ==========================
// HTTP Handlers
// ==========================

type PageData struct {
	Apps    []string
	Envs    []string
	Matrix  map[string]map[string]Cell
	Updated time.Time
}

var store = NewStore()

func main() {
	// Choose fetcher
	if url := os.Getenv("FETCH_URL"); url != "" {
		fetcher = HTTPFetcher{URL: url, Client: &http.Client{Timeout: 10 * time.Second}}
	}
	seedDemoData()
	store.RecomputeDrift()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/matrix", handleMatrix)
	http.HandleFunc("/add-env", handleAddEnv)
	http.HandleFunc("/add-app", handleAddApp)
	http.HandleFunc("/sync", handleSync)
	http.HandleFunc("/export.csv", handleExport)

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func seedDemoData() {
	store.AddEnv("dev")
	store.AddEnv("test")
	store.AddEnv("staging")
	store.AddEnv("prod")
	store.AddApp("auth")
	store.AddApp("billing")
	store.AddApp("orders")
	store.AddApp("catalog")
	store.Set("dev", "auth", "1.3.2", "registry/acme/auth:1.3.2")
	store.Set("test", "auth", "1.3.1", "registry/acme/auth:1.3.1")
	store.Set("staging", "auth", "1.3.0", "registry/acme/auth:1.3.0")
	store.Set("prod", "auth", "1.2.9", "registry/acme/auth:1.2.9")
	store.Set("dev", "billing", "2.0.0-rc1", "registry/acme/billing:2.0.0-rc1")
	store.Set("prod", "billing", "1.9.5", "registry/acme/billing:1.9.5")
	store.Set("dev", "orders", "0.9.1", "registry/acme/orders:0.9.1")
	store.Set("test", "orders", "0.9.0", "registry/acme/orders:0.9.0")
	store.Set("prod", "catalog", "3.4.8", "registry/acme/catalog:3.4.8")
}

func readFilters(r *http.Request) (app, env, contains string) {
	app = strings.TrimSpace(r.URL.Query().Get("app"))
	env = strings.TrimSpace(r.URL.Query().Get("env"))
	contains = strings.TrimSpace(r.URL.Query().Get("contains"))
	return
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	appQ, envQ, contains := readFilters(r)
	apps, envs, matrix := store.Snapshot(appQ, envQ, contains)
	data := PageData{Apps: apps, Envs: envs, Matrix: matrix, Updated: store.LatestUpdated()}
	if err := baseTmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func handleMatrix(w http.ResponseWriter, r *http.Request) {
	appQ, envQ, contains := readFilters(r)
	apps, envs, matrix := store.Snapshot(appQ, envQ, contains)
	data := PageData{Apps: apps, Envs: envs, Matrix: matrix, Updated: store.LatestUpdated()}
	if err := matrixTmpl.ExecuteTemplate(w, "matrix", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func handleAddEnv(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	store.AddEnv(r.Form.Get("env"))
	store.RecomputeDrift()
	handleMatrix(w, r)
}

func handleAddApp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	store.AddApp(r.Form.Get("app"))
	store.RecomputeDrift()
	handleMatrix(w, r)
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	recs, err := fetcher.Fetch()
	if err != nil {
		http.Error(w, fmt.Sprintf("sync error: %v", err), 502)
		return
	}
	for _, rec := range recs {
		store.Set(rec.Env, rec.App, rec.Version, rec.Image)
	}
	store.RecomputeDrift()
	handleMatrix(w, r)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	csvBytes, err := store.ToCSV()
	if err != nil {
		http.Error(w, "failed to export", 500)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=env-matrix.csv")
	_, _ = w.Write(csvBytes)
}

// ==========================
// Version parsing & compare (semver-ish)
// ==========================

type semver struct {
	major, minor, patch int
	pre                 string
	valid               bool
}

func parseSemver(s string) semver {
	s = strings.TrimSpace(s)
	if s == "" {
		return semver{}
	}
	main, pre := s, ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		main, pre = s[:i], s[i+1:]
	}
	parts := strings.Split(main, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return semver{}
	}
	nums := make([]int, 3)
	for i := 0; i < 3; i++ {
		if i < len(parts) {
			n, err := strconv.Atoi(parts[i])
			if err != nil {
				return semver{}
			}
			nums[i] = n
		}
	}
	return semver{major: nums[0], minor: nums[1], patch: nums[2], pre: pre, valid: true}
}

func (a semver) cmp(b semver) int {
	if !a.valid && !b.valid {
		return 0
	}
	if !a.valid {
		return -1
	}
	if !b.valid {
		return 1
	}
	if a.major != b.major {
		if a.major < b.major {
			return -1
		} else {
			return 1
		}
	}
	if a.minor != b.minor {
		if a.minor < b.minor {
			return -1
		} else {
			return 1
		}
	}
	if a.patch != b.patch {
		if a.patch < b.patch {
			return -1
		} else {
			return 1
		}
	}
	// Handle pre-release: empty pre is greater than any pre (e.g., 1.0.0 > 1.0.0-rc1)
	if a.pre == b.pre {
		return 0
	}
	if a.pre == "" {
		return 1
	}
	if b.pre == "" {
		return -1
	}
	// lexical tie-breaker for pre
	if a.pre < b.pre {
		return -1
	} else if a.pre > b.pre {
		return 1
	} else {
		return 0
	}
}
func (a semver) gt(b semver) bool { return a.cmp(b) > 0 }
func (a semver) lt(b semver) bool { return a.cmp(b) < 0 }
func (a semver) eq(b semver) bool { return a.cmp(b) == 0 }

// ==========================
// Helpers
// ==========================

func containsFold(hay, needle string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}
func keysSorted(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
