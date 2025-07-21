package CFG

import (
	"context"
	"errors"
	"fmt"
	"git.sophuwu.com/gophuwu/flags"
	"git.sophuwu.com/manhttpd/neterr"
	"golang.org/x/sys/unix"
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
	TldrGitSrc  string = "https://github.com/tldr-pages/tldr.git"
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

	s, err := flags.GetStringFlag("conf")
	if err != nil {
		neterr.Fatal("Failed to get configuration file flag:", err)
	}
	if s != "" {
		ConfFile = s
	}

	err = parse()
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
	var err error
	server := http.Server{
		Addr:    Addr + ":" + Port,
		Handler: h,
	}
	sigchan := make(chan os.Signal)
	go func() {
		signal.Notify(sigchan, unix.SIGINT, unix.SIGTERM, unix.SIGQUIT, unix.SIGKILL, unix.SIGSTOP)
		<-sigchan
		fmt.Println("Stopping server...")
		err = server.Shutdown(context.Background())
		if err != nil {
			fmt.Println("Error stopping server: %v", err)
		}
	}()
	fmt.Println("Starting server on", server.Addr)
	if UseTLS && TLSCertFile != "" && TLSKeyFile != "" {
		err = server.ListenAndServeTLS(TLSCertFile, TLSKeyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Error starting server:", err)
	}
	fmt.Println("Server stopped.")
}
