package tldr

import (
	"errors"
	"git.sophuwu.com/manweb/CFG"
	"git.sophuwu.com/manweb/embeds"
	"git.sophuwu.com/manweb/logs"
	"git.sophuwu.com/manweb/neterr"
	"net/http"
	"os"
	"strings"
	"time"

	"os/exec"
	"path/filepath"
)

func GitDir() string { return filepath.Join(CFG.TldrDir, "tldr.git") }

var TldrPagesMap = make(map[string]string)

func Open() {
	if !CFG.TldrPages {
		return
	}

	_ = os.MkdirAll(CFG.TldrDir, 0755)
	var cmd *exec.Cmd
	st, err := os.Stat(GitDir())
	if err != nil && os.IsNotExist(err) {
		cmd = exec.Command("/bin/git", "clone", "--bare", CFG.TldrGitSrc, GitDir())
	} else if err != nil {
		logs.CheckFatal("unable to access tldr pages repository", err)
	} else if !st.IsDir() {
		logs.Fatalf("tldr pages repository is not a directory: %s", GitDir())
	} else if st.ModTime().Before(time.Now().AddDate(0, 0, -14)) {
		cmd = exec.Command("/bin/git", "--git-dir", GitDir(), "fetch", "--all")
	}
	if cmd != nil {
		cmd.Dir = CFG.TldrDir
		err = cmd.Run()
		logs.CheckFatal("unable to clone tldr pages repository", err)
	}

	TldrPagesMap = make(map[string]string)
	fn := func(path string) {
		cmd = exec.Command("/bin/git", "--no-pager", "--git-dir", GitDir(), "ls-tree", "--name-only", "main:pages/"+path)
		cmd.Dir = CFG.TldrDir
		var b []byte
		b, err = cmd.Output()
		logs.CheckFatal("unable to list tldr pages", err)
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if len(line) == 0 || !strings.HasSuffix(line, ".md") {
				continue
			}
			name := strings.TrimSuffix(line, ".md")
			TldrPagesMap[name] = filepath.Join(path, line)
		}
	}
	fn("common")
	fn("linux")
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
	cmd := exec.Command("/bin/git", "--git-dir", GitDir(), "show", "main:pages/"+p.Path)
	cmd.Dir = CFG.TldrDir
	b, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return errors.New("Page not found: " + p.Name)
	}
	p.Content = strings.TrimSpace(string(b))
	return nil
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
