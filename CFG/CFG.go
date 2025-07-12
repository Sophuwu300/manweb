package CFG

import (
	"fmt"
	"os"
	"os/exec"
	"sophuwu.site/manhttpd/neterr"
	"strings"
)

var (
	Mandoc string = "mandoc"
	DbCmd  string = "apropos" // or "whatis"
	ManCmd string = "man"

	Hostname string
)

func which(s string) (string, error) {
	c := exec.Command("which", s)
	b, e := c.CombinedOutput()
	return strings.TrimSpace(string(b)), e
}

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

	var err error
	Mandoc, err = which("mandoc")
	fmt.Println(Mandoc, err)
	if err != nil || len(Mandoc) == 0 {
		neterr.Fatal("dependency `mandoc` not found in $PATH, is it installed?\n")
	}
	ManCmd, err = which("man")
	if err != nil {
		neterr.Fatal("dependency `man` not found in $PATH, is it installed?\n")
	}
	DbCmd, err = which("apropos")
	if err != nil || len(DbCmd) == 0 {
		DbCmd, err = which("whatis")
		if err != nil || len(DbCmd) == 0 {
			neterr.Fatal("dependency `apropos` or `whatis` not found in $PATH, is it installed?\n")
		}
	}

	if len(os.Args) > 1 && strings.HasSuffix(os.Args[1], ".conf") {
		ConfFile = os.Args[1]
	}
	DefaultConf, err = ParseEtcConf()
	if err != nil {
		neterr.Fatal("Failed to parse configuration file:", err)
	}
}
