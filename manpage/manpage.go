package manpage

import (
	"fmt"
	"git.sophuwu.com/manhttpd/CFG"
	"git.sophuwu.com/manhttpd/embeds"
	"git.sophuwu.com/manhttpd/neterr"
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

var ManDotName = regexp.MustCompile(`^[^ .]+(\.[0-9]+[a-z]*)?$`)

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

func (m *ManPage) Html() (string, neterr.NetErr) {
	b, err := exec.Command(CFG.Mandoc, "-Thtml", "-O", "fragment", m.Path).Output()
	if err != nil {
		return "", neterr.Err500
	}
	html := LinkRemover(string(b), "")
	html = HTMLManName.ReplaceAllStringFunc(html, func(s string) string {
		mn := HTMLManName.FindStringSubmatch(s)
		return fmt.Sprintf(`<a href="?%s.%s">%s(%s)</a>`, mn[1], mn[2], mn[1], mn[2])
	})
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
	return true
}
