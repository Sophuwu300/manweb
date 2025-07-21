package embeds

import (
	"embed"
	_ "embed"
	"fmt"
	"git.sophuwu.com/manhttpd/CFG"
	"git.sophuwu.com/manhttpd/neterr"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed template/index.html
var index string

//go:embed template/help.html
var help string

//go:embed template/login.html
var LoginPage string

//go:embed static/*
var static embed.FS

type StaticFS struct {
	ContentType string
	Length      string
	Content     []byte
}

func (s *StaticFS) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", s.ContentType)
	w.Header().Set("Content-Length", s.Length)
	w.WriteHeader(http.StatusOK)
	w.Write(s.Content)
}

var constentExt = map[string]string{
	"css": "text/css",
	"js":  "text/javascript",
	"ico": "image/x-icon",
}
var files map[string]StaticFS

func openStatic() {
	d, _ := static.ReadDir("static")
	var sfs StaticFS
	files = make(map[string]StaticFS, len(d))
	ext := ""
	var ok bool
	var f fs.DirEntry
	for _, f = range d {
		sfs = StaticFS{"", "", nil}
		if f.IsDir() {
			continue
		}
		ext = filepath.Ext(f.Name())[1:]
		if sfs.ContentType, ok = constentExt[ext]; !ok {
			continue
		}
		sfs.Content, _ = static.ReadFile("static/" + f.Name())
		sfs.Length = fmt.Sprint(len(sfs.Content))
		files[ext] = sfs
	}
	if len(files) != 3 {
		neterr.Fatal("Failed to load static files, expected 3 types (css, js, ico), got", len(files))
	}
}

var t *template.Template

type Page struct {
	Title    string
	Hostname string
	Content  template.HTML
	Query    string
}

func OpenAndParse() {
	openStatic()
	var e error
	t, e = template.New("index.html").Parse(index)
	if e != nil {
		neterr.Fatal("Failed to parse index template:", e)
	}
	LoginPage = strings.ReplaceAll(LoginPage, "{{ HostName }}", func() string {
		if CFG.Hostname != "" {
			return "@" + CFG.Hostname
		}
		return ""
	}())
}

func StaticFile(name string) (*StaticFS, bool) {
	f, ok := files[name]
	return &f, ok
}

func ChkWriteError(w http.ResponseWriter, r *http.Request, err neterr.NetErr, q string) bool {
	if err == nil {
		return false
	}
	WriteError(w, r, err, q)
	return true
}

func WriteError(w http.ResponseWriter, r *http.Request, err neterr.NetErr, q string) {
	p := Page{
		Title:    err.Error().Title(),
		Hostname: CFG.HttpHostname(r),
		Content:  template.HTML(err.Error().Content()),
		Query:    q,
	}
	t.ExecuteTemplate(w, "index.html", p)
}

func WriteHtml(w http.ResponseWriter, r *http.Request, title, html string, q string, setRawQuery ...string) {
	if len(setRawQuery) > 0 {
		html += "\n" + fmt.Sprintf(`<script>SetRawQuery("%s");</script>`, setRawQuery[0]) + "\n"
	}
	p := Page{
		Title:    title,
		Hostname: CFG.HttpHostname(r),
		Content:  template.HTML(html),
		Query:    q,
	}
	t.ExecuteTemplate(w, "index.html", p)
}
func Help(w http.ResponseWriter, r *http.Request, q string) bool {
	if q == "manweb:help" {
		WriteHtml(w, r, "Help", help, q, q)
		return true
	}
	return false
}
