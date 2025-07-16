package CFG

import (
	"context"
	"errors"
	"fmt"
	"git.sophuwu.com/manhttpd/neterr"
	"golang.org/x/sys/unix"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

var (
	Mandoc      string = "mandoc"
	DbCmd       string = "apropos" // or "whatis"
	ManCmd      string = "man"
	Hostname    string = ""
	Port        string = "8082"
	Addr        string = "0.0.0.0"
	RequireAuth bool   = false
	PasswdFile  string = "/var/lib/manhttpd/authuwu"
	TldrPages   bool   = false
	TldrDir     string = "/var/lib/manhttpd/tldr"
	EnableStats bool   = false
	StatisticDB string = "/var/lib/manhttpd/manhttpd.db"
	UseTLS      bool   = false
	TLSCertFile string = ""
	TLSKeyFile  string = ""
)

func which(s string) (string, error) {
	c := exec.Command("which", s)
	b, e := c.CombinedOutput()
	return strings.TrimSpace(string(b)), e
}

func checkCmds() {
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
}

func getHostname() {
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
}

func getEnvs() {
	var s string
	s = os.Getenv("ListenPort")
	if s != "" {
		Port = s
	}
	s = os.Getenv("ListenAddr")
	if s != "" {
		Addr = s
	}
}

func ParseConfig() {
	checkCmds()

	if len(os.Args) > 1 && strings.HasSuffix(os.Args[1], ".conf") {
		ConfFile = os.Args[1]
	}

	err := parse()
	if err != nil {
		if !errors.Is(err, NoConfError) {
			neterr.Fatal("Failed to parse configuration file:", err)
		}
		getEnvs()
	}

	if Hostname == "" {
		getHostname()
	}
}

func HttpHostname(r *http.Request) string {
	if Hostname != "" {
		return Hostname
	}
	if r.Host != "" {
		return r.Host
	}
	if r.URL != nil && r.URL.Host != "" {
		return r.URL.Host
	}
	if r.TLS != nil && r.TLS.ServerName != "" {
		return r.TLS.ServerName
	}
	return ""
}

func ListenAndServe(h http.Handler) {
	server := http.Server{
		Addr:    Addr + ":" + Port,
		Handler: h,
	}
	var err error
	go func() {
		if UseTLS && TLSCertFile != "" && TLSKeyFile != "" {
			err = server.ListenAndServeTLS(TLSCertFile, TLSKeyFile)
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, unix.SIGINT, unix.SIGTERM, unix.SIGQUIT, unix.SIGKILL, unix.SIGSTOP)
	s := <-sigchan
	println("stopping: got signal", s.String())
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Println("Error stopping server: %v", err)
	}
}
