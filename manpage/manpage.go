package manpage

import (
	"fmt"
	"git.sophuwu.com/manweb/CFG"
	"git.sophuwu.com/manweb/embeds"
	"git.sophuwu.com/manweb/neterr"
	"git.sophuwu.com/manweb/stats"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type ManPage struct {
	Name    string
	Section string
	Path    string
}

func (m *ManPage) Url() string {
	if m.Section != "" && m.Name != "" {
		return fmt.Sprintf("%s.%s", m.Name, m.Section)
	}
	return ""
}

func (m *ManPage) Title() string {
	if m.Section != "" && m.Name != "" {
		return fmt.Sprintf("man %s.%s", m.Name, m.Section)
	}
	return ""
}

func ext(s string) (string, string) {
	if n := filepath.Ext(s); n != "" {
		return s[:len(s)-len(n)], n[1:]
	}
	return s, ""
}

var ManDotName = regexp.MustCompile(`^[^ ]+?(\.[0-9]+[a-z]*)?$`)

func (m *ManPage) Find(q string) bool {
	if !ManDotName.MatchString(q) {
		return false
	}
	b, err := exec.Command(CFG.ManCmd, "--where", q).Output()
	if err != nil {
		return false
	}
	m.Path = strings.TrimSpace(string(b))
	m.Name = filepath.Base(m.Path)
	m.Name, m.Section = ext(m.Name)
	if m.Section == "gz" {
		m.Name, m.Section = ext(m.Name)
	}
	return !(m.Section == "" || m.Name == "" || m.Path == "")
}

var LinkRemover = regexp.MustCompile(`(<a [^>]*>)|(</a>)`).ReplaceAllString
var HTMLManName = regexp.MustCompile(`(?:<b>)?([a-zA-Z0-9_.:\-]+)(?:</b>)?\(([0-9][0-9a-z]*)\)`)
var htmlRep = `<a href="?$1.$2">$1($2)</a>`

func (m *ManPage) Html() (string, neterr.NetErr) {
	b, err := exec.Command(CFG.Mandoc, "-Thtml", "-O", "fragment", m.Path).Output()
	if err != nil {
		return "", neterr.Err500
	}
	html := LinkRemover(string(b), "")
	html = HTMLManName.ReplaceAllString(html, htmlRep)
	return html, nil
}

func Http(w http.ResponseWriter, r *http.Request, q string) bool {
	var m ManPage
	if !m.Find(q) {
		return false
	}
	html, err := m.Html()
	if embeds.ChkWriteError(w, r, err, q) {
		return true
	}
	embeds.WriteHtml(w, r, m.Title(), html, q, m.Url())
	stats.Count(m.Url())
	return true
}
