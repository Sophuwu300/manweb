package tldr

import (
	"errors"
	"fmt"
	"git.sophuwu.com/manhttpd/CFG"
	"git.sophuwu.com/manhttpd/embeds"
	"git.sophuwu.com/manhttpd/neterr"
	"net/http"
	"os"
	"strings"
	"time"

	"os/exec"
	"path/filepath"
)

func GitDir() string { return filepath.Join(CFG.TldrDir, "tldr.git") }

func getGitList(path string, mp *map[string]string) error {
	cmd := exec.Command("/bin/git", "--git-dir", GitDir(), "--no-color", "show", "main:"+path)
	cmd.Dir = CFG.TldrDir
	b, err := cmd.Output()
	if err != nil {
		return err
	}
	var k, v string
	for _, v = range strings.Split(string(b), "\n") {
		if strings.HasSuffix(v, ".md") {
			v = strings.TrimSpace(v)
			k = strings.TrimSuffix(v, ".md")
			v = filepath.Join(path, v)
			(*mp)[k] = v
		}
	}
	return nil
}

var shspt = `#!/bin/bash

cd '{{ .TldrDir }}'
if [[ -d '{{ .GitDir }}' ]]; then
	rm -rf '{{ .GitDir }}'
fi
set -e

git clone --filter=blob:none --no-checkout '{{ .TldrGitSrc }}' '{{ .GitDir }}'
cd '{{ .GitDir }}'
git sparse-checkout init
git sparse-checkout set pages/linux/*.md pages/common/*.md
git checkout main
mkdir -p '{{ .TldrDir }}/pages'
find "{{ .GitDir }}/pages/common" -type f -name "*.md" -exec cp "{}" '{{ .TldrDir }}/pages/' \;
find "{{ .GitDir }}/pages/linux" -type f -name "*.md" -exec cp "{}" '{{ .TldrDir }}/pages/' \;
`

var TldrPagesMap = make(map[string]string)

var PageDir string

func updateTldrPages() error {
	sh := strings.ReplaceAll(shspt, "{{ .TldrDir }}", CFG.TldrDir)
	sh = strings.ReplaceAll(sh, "{{ .GitDir }}", GitDir())
	sh = strings.ReplaceAll(sh, "{{ .TldrGitSrc }}", CFG.TldrGitSrc)
	upcmd := filepath.Join(CFG.TldrDir, "update.sh")
	_ = os.Remove(upcmd)
	err := os.WriteFile(upcmd, []byte(sh), 0755)
	if err != nil {
		return err
	}
	cmd := exec.Command(upcmd)
	cmd.Env = os.Environ()
	cmd.Dir = CFG.TldrDir
	var b []byte
	b, err = cmd.CombinedOutput()
	fmt.Println("update tldr pages:", string(b))
	if err != nil {
		return err
	}
	return nil
}

func Open() error {
	if !CFG.TldrPages {
		return nil
	}
	PageDir = filepath.Join(CFG.TldrDir, "pages")

	update := false
	st, err := os.Stat(PageDir)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(PageDir, 0755); err != nil {
			return err
		}
		update = true
	} else if time.Now().After(st.ModTime().AddDate(0, 0, 14)) {
		update = true
	}
	if update {
		if err = updateTldrPages(); err != nil {
			return fmt.Errorf("failed to update tldr pages: %w", err)
		}
	}

	TldrPagesMap = make(map[string]string)
	var de []os.DirEntry
	de, err = os.ReadDir(PageDir)
	if err != nil {
		return err
	}
	var name string
	for _, d := range de {
		if d.IsDir() {
			continue
		}
		name = strings.TrimSuffix(d.Name(), ".md")
		TldrPagesMap[name] = filepath.Join(PageDir, d.Name())
	}
	return nil
}

type TldrPage struct {
	Name    string
	Path    string
	Content string
}

func (p *TldrPage) findPath() error {
	if p.Name == "" {
		return errors.New("tldr page name cannot be empty")
	}
	var ok bool
	p.Path, ok = TldrPagesMap[p.Name]
	if !ok {
		return errors.New("tldr page not found: " + p.Name)
	}
	return nil
}

func (p *TldrPage) open() error {
	err := p.findPath()
	if err != nil {
		return err
	}
	// cmd := exec.Command("/bin/git", "--git-dir", GitDir(), "--no-color", "show", "main:"+p.Path)
	// cmd.Dir = CFG.TldrDir
	// b, err := cmd.Output()
	b, err := os.ReadFile(p.Path)
	p.Content = string(b)
	return err
}

/*
	tldr page format:
	# tldr page name
	> tldr page description
	- command 1
		`command 1 usage`
	- command 2
		`command 2 usage`

	Specification:
		lines beginning with `#` are titles
		lines beginning with `>` are descriptions
			inline urls, inside <> tags
			may contain inline code
		lines beginning with `-` list elements
		 	may contain inline code
			`codeblocks` on same level

*/

func inlineLink(s string) string {
	i := strings.Index(s, "<")
	if i < 0 {
		return s
	}
	j := strings.Index(s[i:], ">")
	if j < 0 {
		return s
	}
	j += i
	return s[:i] + `<a href="` + s[i+1:j] + `">` + s[i+1:j] + `</a>` + inlineLink(s[j+1:])
}
func inlineCode(s string) string {
	i := strings.Index(s, "`")
	if i < 0 {
		return s
	}
	j := strings.Index(s[i+1:], "`")
	if j < 0 {
		return s
	}
	j += i + 1
	return s[:i] + `<code>` + s[i+1:j] + `</code>` + inlineCode(s[j+1:])
}

func inline(s string) string {
	s = inlineLink(s)
	s = inlineCode(s)
	return s
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func (p *TldrPage) HTML() (string, error) {
	s := ""
	lines := strings.Split(p.Content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		i := strings.Index(line, " ")
		if strings.HasPrefix(line, "#") {
			if i < 0 {
				return "", errors.New("invalid tldr page format: missing space after title")
			}
			s += `<h1 class="tldr">` + htmlEscape(line[i+1:]) + "</h1>\n"
		} else if strings.HasPrefix(line, ">") {
			if i < 0 {
				return "", errors.New("invalid tldr page format: missing space after description")
			}
			s += `<p class="desc tldr">` + inline(htmlEscape(line[i+1:])) + "</p>\n"
		} else if strings.HasPrefix(line, "-") {
			if i < 0 {
				return "", errors.New("invalid tldr page format: missing space after list item")
			}
			s += `<p class="list-item tldr">` + inline(htmlEscape(line[i+1:])) + "</p>\n"
		} else if strings.HasPrefix(line, "`") && strings.HasSuffix(line, "`") {
			s += `<pre class="list-item tldr">` + htmlEscape(line[1:len(line)-1]) + "</pre>\n"
		} else {
			s += `<p class="list-item tldr">` + inline(htmlEscape(line)) + "</p>\n"
		}
	}
	if len(s) == 0 {
		return "", errors.New("invalid tldr page format: no content found")
	}
	s = `<div class="tldr-page">` + s + "</div>\n"
	return s, nil
}

func OpenTldrPage(name string) (*TldrPage, neterr.NetErr) {
	if name == "" {
		return nil, neterr.Err400
	}
	page := &TldrPage{Name: name}
	err := page.open()
	if err != nil {
		return nil, neterr.Err404
	}
	return page, nil
}

func Http(w http.ResponseWriter, r *http.Request, q string) bool {
	if filepath.Ext(q) != ".tldr" {
		return false
	}
	name := strings.TrimSuffix(q, ".tldr")
	page, nerr := OpenTldrPage(name)
	if embeds.ChkWriteError(w, r, nerr, q) {
		return true
	}
	html, err := page.HTML()
	if err != nil {
		embeds.WriteError(w, r, neterr.Err500, q)
		return true
	}
	embeds.WriteHtml(w, r, "TLDR: "+name, html, q, q)
	return true
}
