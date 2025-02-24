package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed template/index.html
var index string

//go:embed template/help.html
var help string

//go:embed template/theme.css
var css []byte

//go:embed template/scripts.js
var scripts []byte

//go:embed template/favicon.ico
var favicon []byte

var CFG struct {
	Hostname string
	Port     string
	Mandoc   string
	DbCmd    string
	ManCmd   string
	Server   http.Server
	Addr     string
}

func Fatal(v ...interface{}) {
	fmt.Fprintln(os.Stderr, "manhttpd exited due to an error it could not recover from.")
	fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprint(v...))
	os.Exit(1)
}

func GetCFG() {
	var e error
	var b []byte
	var s string
	if s = os.Getenv("HOSTNAME"); s != "" {
		CFG.Hostname = s
	} else if s, e = os.Hostname(); e == nil {
		CFG.Hostname = s
	} else if b, e = os.ReadFile("/etc/hostname"); e == nil {
		CFG.Hostname = strings.TrimSpace(string(b))
	}
	if CFG.Hostname == "" {
		CFG.Hostname = "Unresolved"
	}
	index = strings.ReplaceAll(index, "{{ hostname }}", CFG.Hostname)

	if b, e = exec.Command("which", func() string {
		if s = os.Getenv("MANDOCPATH"); s != "" {
			return s
		}
		return "mandoc"
	}()).Output(); e != nil || len(b) == 0 {
		Fatal("dependency `mandoc` not found in $PATH, is it installed?\n")
	} else {
		CFG.Mandoc = strings.TrimSpace(string(b))
	}
	f := func(s string) string {
		if b, e = exec.Command("which", s).Output(); e != nil || len(b) == 0 {
			return ""
		}
		return strings.TrimSpace(string(b))
	}

	if s = f("man"); s == "" {
		Fatal("dependency `man` not found. `man` and its libraries are required for manhttpd to function.")
	} else {
		CFG.ManCmd = s
	}

	if s = f("apropos"); s == "" {
		Fatal("dependency `apropos` not found. `apropos` is required for search functionality.")
	} else {
		CFG.DbCmd = s
	}

	CFG.Port = os.Getenv("ListenPort")
	if CFG.Port == "" {
		CFG.Port = "8082"
	}
	CFG.Addr = os.Getenv("ListenAddr")
	if CFG.Addr == "" {
		CFG.Addr = "0.0.0.0"
	}
}

func main() {
	GetCFG()
	server := http.Server{
		Addr:    CFG.Addr + ":" + CFG.Port,
		Handler: http.HandlerFunc(MuxHandler),
	}
	_ = server.ListenAndServe()
}

func WriteFile(w http.ResponseWriter, file []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprint(len(file)))
	w.WriteHeader(http.StatusOK)
	w.Write(file)
}

func MuxHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	l := len(p)
	if l >= 9 {
		p = p[l-9:]
	}
	switch p {
	case "style.css":
		WriteFile(w, css, "text/css")
	case "script.js":
		WriteFile(w, scripts, "text/javascript")
	case "icons.ico":
		WriteFile(w, favicon, "image/x-icon")
	default:
		IndexHandler(w, r)
	}
	r.Body.Close()
}

func WriteHtml(w http.ResponseWriter, r *http.Request, title, html string, q string) {
	out := strings.ReplaceAll(index, "{{ host }}", r.Host)
	out = strings.ReplaceAll(out, "{{ title }}", title)
	out = strings.ReplaceAll(out, "{{ query }}", strings.ReplaceAll(q, "`", "\\`"))
	out = strings.ReplaceAll(out, "{{ content }}", html)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)
}

var LinkRemover = regexp.MustCompile(`(<a [^>]*>)|(</a>)`).ReplaceAllString
var HTMLManName = regexp.MustCompile(`(?:<b>)?([a-zA-Z0-9_.:\-]+)(?:</b>)?\(([0-9][0-9a-z]*)\)`)

type ManPage struct {
	Name    string
	Section string
	Desc    string
	Path    string
}

func (m *ManPage) Where() error {
	var arg = []string{"-w", m.Name}
	if m.Section != "" {
		arg = []string{"-w", "-s" + m.Section, m.Name}
	}
	b, err := exec.Command("man", arg...).Output()
	m.Path = strings.TrimSpace(string(b))
	return err
}
func (m *ManPage) Html() (string, NetErr) {
	if m.Where() != nil {
		return "", e404
	}
	b, err := exec.Command(CFG.Mandoc, "-Thtml", "-O", "fragment", m.Path).Output()
	if err != nil {
		return "", e500
	}
	html := LinkRemover(string(b), "")
	html = HTMLManName.ReplaceAllStringFunc(html, func(s string) string {
		m := HTMLManName.FindStringSubmatch(s)
		return fmt.Sprintf(`<a href="?%s.%s">%s(%s)</a>`, m[1], m[2], m[1], m[2])
	})
	return html, nil
}

var ManDotName = regexp.MustCompile(`^([a-zA-Z0-9_\-]+)(?:\.([0-9a-z]+))?$`)

func NewManPage(s string) (m ManPage) {
	name := ManDotName.FindStringSubmatch(s)
	if len(name) >= 2 {
		m.Name = name[1]
	}
	if len(name) >= 3 {
		m.Section = name[2]
	}
	return
}

var RxWords = regexp.MustCompile(`("[^"]+")|([^ ]+)`).FindAllString
var RxWhatIs = regexp.MustCompile(`([a-zA-Z0-9_\-]+) [(]([0-9a-z]+)[)][\- ]+(.*)`).FindAllStringSubmatch

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.Form.Get("q")
	if q == "" {
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}
	if func() bool {
		m := NewManPage(q)
		return m.Where() == nil
	}() {
		http.Redirect(w, r, "?"+q, http.StatusFound)
		return
	}

	var args = RxWords("-lw "+q, -1)

	for i := range args {
		args[i] = strings.TrimSpace(args[i])
		args[i] = strings.TrimPrefix(args[i], `"`)
		args[i] = strings.TrimSuffix(args[i], `"`)
		if (args[i] == "-r" || args[i] == "-w") && args[0] != "-l" {
			args[0] = "-l"
		}
	}

	cmd := exec.Command(CFG.DbCmd, args...)
	b, e := cmd.Output()
	if len(b) < 1 || e != nil {
		fmt.Println(e)
		e404.Write(w, r, q)
		return
	}
	var output string
	for _, line := range RxWhatIs(string(b), -1) { // strings.Split(string(b), "\n") {
		if len(line) == 4 {
			output += fmt.Sprintf(`<p><a href="?%s.%s">%s (%s)</a> - %s</p>%c`, line[1], line[2], line[1], line[2], line[3], 10)
		}
	}
	WriteHtml(w, r, "Search", output, q)
}

var (
	e400 = HTCode(400, "Bad Request",
		"Your request cannot be understood by the server.",
		"Check that you are using a release version of manhttpd.",
		"Please check spelling and try again.",
		"Otherwise browser extensions or proxies may be hijacking your requests.",
	)
	e404 = HTCode(404, "Not Found",
		"The requested does match any known page names. Please check your spelling and try again.",
		`If you cannot find the page using your system's man command, then you may need to update your manDB or apt-get <b>&lt;package&gt;-doc</b>.`,
		"If you can open a page using the cli but not in manhttpd, your service is misconfigured. For best results set user and group to your login user:",
		`You can edit &lt;<b>/etc/systemd/system/manhttpd.service</b>&gt; and set "<b>User</b>=&lt;<b>your-user</b>&gt;" and "<b>Group</b>=&lt;<b>your-group</b>&gt;".`,
		`Usually root user will work just fine, however root does not index user pages. If manuals are installed without superuser, they are saved to &lt;<b>$HOME/.local/share/man/</b>&gt;.`,
		`If you want user pages you have to run manhttpd as your login user. If you really want to run the service as root with user directories, at your own risk: adding users' homes into the global path &lt;<b>/etc/manpath.config</b>&gt; is usually safe but may cause catastrophic failure on some systems.`,
	)
	e500 = HTCode(500, "Internal Server Error",
		"The server encountered an error and could not complete your request.",
		"Make sure you are using a release version of manhttpd.",
	)
)

func HTCode(code int, name string, desc ...string) HTErr {
	return HTErr{code, name, desc}
}
func (h HTErr) Write(w http.ResponseWriter, r *http.Request, s ...string) {
	if len(s) == 0 {
		s = append(s, r.URL.RawQuery)
	}
	WriteHtml(w, r, h.Title(), h.Content(), s[0])
}
func (h HTErr) Is(err error, w http.ResponseWriter, r *http.Request, s ...string) bool {
	if err == nil {
		return false
	}
	h.Write(w, r, s...)
	return true
}

type HTErr struct {
	Code int
	Name string
	Desc []string
}
type NetErr interface {
	Error() HTErr
	Write(w http.ResponseWriter, r *http.Request, s ...string)
}

func (e HTErr) Error() HTErr {
	return e
}
func (e HTErr) Title() string {
	return fmt.Sprintf("%d %s", e.Code, e.Name)
}
func (e HTErr) Content() string {
	s := fmt.Sprintf("<h1>%3d</h1><h2>%s</h2><br>\n", e.Code, e.Name)
	for d := range e.Desc {
		s += fmt.Sprintf("<p>%s</p><br>\n", e.Desc[d])
	}

	s += `<script>SetRawQuery()</script>
`
	return s
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Base(r.URL.Path)
	path = strings.TrimSuffix(path, "/")

	if r.Method == "POST" {
		err := r.ParseForm()
		if !e400.Is(err, w, r) {
			searchHandler(w, r)
		}
		return
	}

	name := r.URL.RawQuery

	if name == "manweb-help.html" {
		WriteHtml(w, r, "Help", help, name)
		return
	}

	var nerr NetErr
	title := "Index"
	var html string
	if name != "" {
		man := NewManPage(name)
		html, nerr = man.Html()
		if nerr != nil {
			nerr.Write(w, r, name)
			return
		}
		title = man.Name
	}
	WriteHtml(w, r, title, html, name)
	return
}
