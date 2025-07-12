package CFG

import (
	"net/http"
	"os"
	"os/exec"
	"sophuwu.site/manhttpd/neterr"
	"strings"
)

var (
	Hostname string
	Port     string
	Mandoc   string
	DbCmd    string
	ManCmd   string
	Server   http.Server
	Addr     string
)

func init() {
	var e error
	var b []byte
	var s string
	if s = os.Getenv("HOSTNAME"); s != "" {
		Hostname = s
	} else if s, e = os.Hostname(); e == nil {
		Hostname = s
	} else if b, e = os.ReadFile("/etc/hostname"); e == nil {
		Hostname = strings.TrimSpace(string(b))
	}
	if Hostname == "" {
		Hostname = "Unresolved"
	}

	if b, e = exec.Command("which", func() string {
		if s = os.Getenv("MANDOCPATH"); s != "" {
			return s
		}
		return "mandoc"
	}()).Output(); e != nil || len(b) == 0 {
		neterr.Fatal("dependency `mandoc` not found in $PATH, is it installed?\n")
	} else {
		Mandoc = strings.TrimSpace(string(b))
	}
	f := func(s string) string {
		if b, e = exec.Command("which", s).Output(); e != nil || len(b) == 0 {
			return ""
		}
		return strings.TrimSpace(string(b))
	}

	if s = f("man"); s == "" {
		neterr.Fatal("dependency `man` not found. `man` and its libraries are required for manhttpd to function.")
	} else {
		ManCmd = s
	}

	if s = f("apropos"); s == "" {
		neterr.Fatal("dependency `apropos` not found. `apropos` is required for search functionality.")
	} else {
		DbCmd = s
	}

	Port = os.Getenv("ListenPort")
	if Port == "" {
		Port = "8082"
	}
	Addr = os.Getenv("ListenAddr")
	if Addr == "" {
		Addr = "0.0.0.0"
	}

	Server = http.Server{
		Addr:    Addr + ":" + Port,
		Handler: nil,
	}
}
