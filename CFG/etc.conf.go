package CFG

import (
	"errors"
	"fmt"
	"git.sophuwu.com/manweb/logs"
	"os"
	"strings"
)

var ConfFile = "/etc/manweb/manweb.conf"

var NoConfError = errors.New("no configuration file found")

func setV(a any) func(j string) error {
	switch v := a.(type) {
	case *string:
		return func(j string) error {
			*v = j
			return nil
		}
	case *bool:
		return func(j string) error {
			j = strings.ToLower(j)
			*v = "yes" == j
			if (*v) || "no" == j {
				return nil
			}
			return errors.New("invalid boolean value: " + j)
		}
	}
	return nil
}

func rmComment(s *string) bool {
	i := strings.Index(*s, "#")
	if i >= 0 {
		*s = (*s)[:i]
	}
	*s = strings.TrimSpace(*s)
	return len(*s) == 0
}

func getKV(line, k, v *string) bool {
	i := strings.Index(*line, "=")
	if i < 0 {
		return false
	}
	*k = strings.TrimSpace((*line)[:i])
	*v = strings.TrimSpace((*line)[i+1:])
	return len(*k) == 0 || len(*v) == 0
}

var mp = map[string]any{
	"hostname":      &Hostname,
	"port":          &Port,
	"addr":          &Addr,
	"require_auth":  &RequireAuth,
	"passwd_file":   &PasswdFile,
	"tldr_pages":    &TldrPages,
	"tldr_dir":      &TldrDir,
	"tldr_git_src":  &TldrGitSrc,
	"enable_stats":  &EnableStats,
	"statistic_db":  &StatisticDB,
	"use_tls":       &UseTLS,
	"tls_cert_file": &TLSCertFile,
	"tls_key_file":  &TLSKeyFile,
}

func parse() error {
	b, err := os.ReadFile(ConfFile)
	if err != nil {
		return NoConfError
	}
	var line, k, v string
	var i int
	var a any
	var fn func(j string) error
	var ok bool
	errFmt := "invalid %s in " + ConfFile + " at line %d: %s"
	ErrPrint := func(e, s string) {
		logs.Logf(errFmt, e, i+1, s)
	}
	for i, line = range strings.Split(string(b), "\n") {
		if len(line) == 0 {
			continue
		}
		if rmComment(&line) {
			continue
		}
		if getKV(&line, &k, &v) {
			ErrPrint("format", line)
			continue
		}
		if a, ok = mp[k]; !ok {
			ErrPrint("key", k)
			continue
		}
		fn = setV(a)
		if fn == nil {
			continue
		}
		err = fn(v)
		if err != nil {
			ErrPrint(fmt.Sprintf("value for %s", k), v)
			continue
		}
	}
	return nil
}

func MakeConfig() string {
	var s, i string
	for k, a := range mp {
		switch v := a.(type) {
		case *string:
			i = *v
		case *bool:
			i = map[bool]string{true: "yes", false: "no"}[*v]
		default:
			continue
		}
		if len(i) == 0 {
			continue
		}
		s += fmt.Sprintf("%s = %s\n", k, i)
	}
	return s
}
