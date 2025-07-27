package main

import (
	"fmt"
	"git.sophuwu.com/authuwu"
	"git.sophuwu.com/authuwu/userpass"
	"git.sophuwu.com/gophuwu/flags"
	"git.sophuwu.com/manhttpd/CFG"
	"git.sophuwu.com/manhttpd/embeds"
	"git.sophuwu.com/manhttpd/logs"
	"git.sophuwu.com/manhttpd/manpage"
	"git.sophuwu.com/manhttpd/neterr"
	"git.sophuwu.com/manhttpd/tldr"
	"golang.org/x/term"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func init() {
	err := flags.NewFlag("conf", "c", "configuration file to use", "/etc/manhttpd/manhttpd.conf")
	logs.CheckFatal("creating conf flag", err)
	err = flags.NewFlag("passwd", "p", "open the program in password edit mode", false)
	logs.CheckFatal("creating passwd flag", err)
	err = flags.NewFlag("user", "u", "choose a username to set/change/delete password for", "")
	logs.CheckFatal("creating user flag", err)
	err = flags.ParseArgs()
	logs.CheckFatal("parsing flags", err)
	CFG.ParseConfig()
	embeds.OpenAndParse()
	tldr.Open()
}

func setPasswd() {
	u, err := flags.GetStringFlag("user")
	if err != nil {
		fmt.Println("getting user flag:", err)
		return
	}
	if u == "" {
		fmt.Println("no user specified, use -u <username> to set a password for a user")
		return
	}
	var in []byte
	fmt.Printf("Enter password for user %s (leave empty to delete user): \n", u)
	in, err = term.ReadPassword(0)
	if err != nil {
		fmt.Println("could not read password:", err)
		return
	}
	password := string(in)
	if password == "" {
		fmt.Printf("delete user %s? (y/N): ", u)
		in = make([]byte, 1)
		os.Stdin.Read(in)
		if in[0] != 'y' && in[0] != 'Y' {
			userpass.DeleteUser(u)
			fmt.Printf("User %s deleted.\n", u)
			return
		}
		fmt.Printf("exiting with no changes\n")
		return
	}
	fmt.Println("Enter password again: ")
	in, err = term.ReadPassword(0)
	if err != nil {
		fmt.Println("could not read password:", err)
		return
	}
	if string(in) != password {
		fmt.Println("Passwords do not match, please try again.")
		return
	}
	err = userpass.NewUser(u, password)
}

func main() {
	b, err := flags.GetBoolFlag("passwd")
	logs.CheckFatal("getting passwd flag", err)
	if b {
		err = authuwu.OpenDB(CFG.PasswdFile)
		if err != nil {
			if err.Error() == "timeout" {
				logs.Error("The database is currently opened by another process.")
				return
			}
			logs.Fatal("opening password database", err)
		}
		setPasswd()
		authuwu.CloseDB()
		return
	}
	if CFG.RequireAuth {
		err = authuwu.OpenDB(CFG.PasswdFile)
		logs.CheckFatal("opening password database", err)
		PageHandler = authuwu.NewAuthuwuHandler(PageHandler, time.Hour*24*3, embeds.LoginPage)
		defer authuwu.CloseDB()
	}
	CFG.ListenAndServe(Handler)
}

var RxWords = regexp.MustCompile(`("[^"]+")|([^ ]+)`).FindAllString
var RxWhatIs = regexp.MustCompile(`([a-zA-Z0-9_\-]+) [(]([0-9a-z]+)[)][\- ]+(.*)`).FindAllStringSubmatch

func SearchHandler(w http.ResponseWriter, r *http.Request, q string) {
	var args = RxWords("-lw "+q, -1)

	for i := range args {
		args[i] = strings.TrimSpace(args[i])
		args[i] = strings.TrimPrefix(args[i], `"`)
		args[i] = strings.TrimSuffix(args[i], `"`)
		if (args[i] == "-r" || args[i] == "-w") && args[0] != "-l" {
			args[0] = "-l"
		}
	}

	cmd := exec.Command(CFG.DbCmd, args...)
	b, e := cmd.Output()
	if len(b) < 1 || e != nil {
		embeds.WriteError(w, r, neterr.Err404, q)
		return
	}
	var output string
	for _, line := range RxWhatIs(string(b), -1) { // strings.Split(string(b), "\n") {
		if len(line) == 4 {
			output += fmt.Sprintf(`<p><a href="?%s.%s">%s (%s)</a> - %s</p>%c`, line[1], line[2], line[1], line[2], line[3], 10)
		}
	}
	embeds.WriteHtml(w, r, "Search", output, q, "")
}

var PageHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	name := r.URL.RawQuery
	name = strings.TrimSpace(name)
	if r.Method == "POST" {
		n := r.PostFormValue("q")
		n = strings.TrimSpace(n)
		if n != "" {
			name = n
			if strings.ContainsAny(name, `"*?^|`) {
				SearchHandler(w, r, name)
				return
			}
		}
	}
	if name == "" {
		embeds.WriteHtml(w, r, "Index", "", "", "")
		return
	}
	if embeds.Help(w, r, name) {
		return
	}
	if manpage.Http(w, r, name) {
		return
	}
	if CFG.TldrPages && tldr.Http(w, r, name) {
		return
	}
	SearchHandler(w, r, name)
})

var Handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("static") {
		StaticHandler(w, r)
		return
	}
	PageHandler.ServeHTTP(w, r)
})

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("static")
	if f, ok := embeds.StaticFile(q); ok {
		w.Header().Set("Content-Type", f.ContentType)
		w.Header().Set("Content-Length", f.Length)
		f.WriteTo(w)
		return
	}
}
