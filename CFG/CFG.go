package CFG

import (
	"context"
	"errors"
	"git.sophuwu.com/gophuwu/flags"
	"git.sophuwu.com/manhttpd/logs"
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
	errfmt := "dependency `%s` not found"
	Mandoc, err = which("mandoc")
	if err != nil || len(Mandoc) == 0 {
		logs.Fatalf(errfmt, "mandoc")
	}
	ManCmd, err = which("man")
	if err != nil || len(ManCmd) == 0 {
		logs.Fatalf(errfmt, "man")
	}
	DbCmd, err = which("apropos")
	if err != nil || len(DbCmd) == 0 {
		DbCmd, err = which("whatis")
		if err != nil || len(DbCmd) == 0 {
			logs.Fatalf(errfmt, "apropos")
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
	logs.CheckFatal("getting conf flag", err)
	if s != "" {
		ConfFile = s
	}

	err = parse()
	if err != nil {
		if !errors.Is(err, NoConfError) {
			logs.Fatal("parsing conf file", err)
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
		logs.Log("Stopping server...")
		err = server.Shutdown(context.Background())
		if err != nil {
			logs.Log("Error stopping server: %v", err)
		}
	}()
	logs.Log("Starting server on", server.Addr)
	if UseTLS && TLSCertFile != "" && TLSKeyFile != "" {
		err = server.ListenAndServeTLS(TLSCertFile, TLSKeyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logs.Log("Error starting server:", err)
	}
	logs.Log("Server stopped.")
}
