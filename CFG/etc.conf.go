package CFG

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

var ConfFile = "/etc/manhttpd/manhttpd.conf"

type EtcConf struct {
	Hostname    string
	Port        string
	Addr        string
	RequireAuth bool
	PasswdFile  string
	TldrPages   bool
	EnableStats bool
	StatisticDB string
}

var DefaultConf = EtcConf{
	Hostname:    "",
	Port:        "8082",
	Addr:        "0.0.0.0",
	RequireAuth: false,
	PasswdFile:  "/etc/manhttpd/passwd",
	TldrPages:   false,
	EnableStats: false,
	StatisticDB: "/var/lib/manhttpd/stats.db",
}

func (c *EtcConf) Parse() error {
	var mp = map[string]any{
		"hostname":     &(c.Hostname),
		"port":         &(c.Port),
		"addr":         &(c.Addr),
		"require_auth": &(c.RequireAuth),
		"passwd_file":  &(c.PasswdFile),
		"tldr_pages":   &(c.TldrPages),
		"enable_stats": &(c.EnableStats),
		"statistic_db": &(c.StatisticDB),
	}
	b, err := os.ReadFile(ConfFile)
	if err != nil {
		return err
	}
	var j, k string
	var kv []string
	for _, v := range strings.Split(string(b), "\n") {
		if len(v) == 0 || v[0] == '#' {
			continue // skip empty lines and comments
		}
		kv = strings.SplitN(v, " ", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid line in %s: %s", ConfFile, v)
		}
		k = strings.TrimSpace(kv[0])
		j = strings.TrimSpace(kv[1])
		if val, ok := mp[k]; ok {
			switch v := val.(type) {
			case *string:
				*v = j
			case *bool:
				*v = j == "yes"
			default:
				return fmt.Errorf("unsupported type for key %s in %s", k, ConfFile)
			}
		} else {
			return fmt.Errorf("unknown key %s in %s", k, ConfFile)
		}
	}
	return nil
}

func ParseEtcConf() (EtcConf, error) {
	c := DefaultConf
	if err := c.Parse(); err != nil {
		return c, fmt.Errorf("failed to parse %s: %w", ConfFile, err)
	}
	if c.Port == "" {
		c.Port = "8082"
	}
	if c.Addr == "" {
		c.Addr = "0.0.0.0"
	}
	if c.Hostname != "" {
		Hostname = c.Hostname
	}
	return c, nil
}

func (c *EtcConf) Server(h http.Handler) *http.Server {
	return &http.Server{
		Addr:    c.Addr + ":" + c.Port,
		Handler: h,
	}
}
