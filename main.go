package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"sophuwu.site/manhttpd/CFG"
	"sophuwu.site/manhttpd/embeds"
	"sophuwu.site/manhttpd/manpage"
	"sophuwu.site/manhttpd/neterr"
	"strings"
)

func main() {
	CFG.Server.Handler = ManHandler{}
	err := CFG.Server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

var RxWords = regexp.MustCompile(`("[^"]+")|([^ ]+)`).FindAllString
var RxWhatIs = regexp.MustCompile(`([a-zA-Z0-9_\-]+) [(]([0-9a-z]+)[)][\- ]+(.*)`).FindAllStringSubmatch

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if neterr.Err400.Is(err) {
		embeds.WriteError(w, r, neterr.Err400)
		return
	}
	q := r.Form.Get("q")
	if q == "" {
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}
	if strings.HasPrefix(q, "manweb:") {
		http.Redirect(w, r, "?"+q, http.StatusFound)
		return
	}
	if func() bool {
		m := manpage.New(q)
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
		embeds.WriteError(w, r, neterr.Err404)
		return
	}
	var output string
	for _, line := range RxWhatIs(string(b), -1) { // strings.Split(string(b), "\n") {
		if len(line) == 4 {
			output += fmt.Sprintf(`<p><a href="?%s.%s">%s (%s)</a> - %s</p>%c`, line[1], line[2], line[1], line[2], line[3], 10)
		}
	}
	embeds.WriteHtml(w, r, "Search", output, q)
}

type ManHandle interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type ManHandler struct {
}

func (m ManHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("static") {
		StaticHandler(w, r)
		return
	}

	if r.Method == "POST" {
		SearchHandler(w, r)
		return
	}
	name := r.URL.RawQuery
	if name == "manweb:help" {
		embeds.Help(w, r)
		return
	}

	var nerr neterr.NetErr
	title := "Index"
	var html string
	if name != "" {
		man := manpage.New(name)
		html, nerr = man.Html()
		if nerr != nil {
			embeds.WriteError(w, r, nerr)
			return
		}
		title = man.Name
	}
	embeds.WriteHtml(w, r, title, html, name)
	return
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("static")
	if f, ok := embeds.StaticFile(q); ok {
		w.Header().Set("Content-Type", f.ContentType)
		w.Header().Set("Content-Length", f.Length)
		f.WriteTo(w)
		return
	}
}
