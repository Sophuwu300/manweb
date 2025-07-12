package manpage

import (
	"fmt"
	"os/exec"
	"regexp"
	"sophuwu.site/manhttpd/CFG"
	"sophuwu.site/manhttpd/neterr"
	"strings"
)

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
	b, err := exec.Command(CFG.ManCmd, arg...).Output()
	m.Path = strings.TrimSpace(string(b))
	return err
}
func (m *ManPage) Html() (string, neterr.NetErr) {
	if m.Where() != nil {
		return "", neterr.Err404
	}
	b, err := exec.Command(CFG.Mandoc, "-Thtml", "-O", "fragment", m.Path).Output()
	if err != nil {
		return "", neterr.Err500
	}
	html := LinkRemover(string(b), "")
	html = HTMLManName.ReplaceAllStringFunc(html, func(s string) string {
		m := HTMLManName.FindStringSubmatch(s)
		return fmt.Sprintf(`<a href="?%s.%s">%s(%s)</a>`, m[1], m[2], m[1], m[2])
	})
	return html, nil
}

var ManDotName = regexp.MustCompile(`^([a-zA-Z0-9_\-]+)(?:\.([0-9a-z]+))?$`)

func New(s string) (m ManPage) {
	name := ManDotName.FindStringSubmatch(s)
	if len(name) >= 2 {
		m.Name = name[1]
	}
	if len(name) >= 3 {
		m.Section = name[2]
	}
	return
}
