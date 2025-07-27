package neterr

import (
	"fmt"
)

var (
	Err400 = HTCode(400, "Bad Request",
		"Your request cannot be understood by the server.",
		"Check that you are using a release version of manweb.",
		"Please check spelling and try again.",
	)
	Err404 = HTCode(404, "Not Found",
		"The requested does match any known page names. Please check your spelling and try again.",
		`If you cannot find the page using your system's man command, then you may need to update your manDB or apt-get <b>&lt;package&gt;-doc</b>.`,
		"If you can open a page using the cli but not in manweb, your service is misconfigured. For best results set user and group to your login user:",
		`You can edit &lt;<b>/etc/systemd/system/manweb.service</b>&gt; and set "<b>User</b>=&lt;<b>your-user</b>&gt;" and "<b>Group</b>=&lt;<b>your-group</b>&gt;".`,
		`Usually root user will work just fine, however root does not index user pages. If manuals are installed without superuser, they are saved to &lt;<b>$HOME/.local/share/man/</b>&gt;.`,
		`If you want user pages you have to run manweb as your login user. If you really want to run the service as root with user directories, at your own risk: adding users' homes into the global path &lt;<b>/etc/manpath.config</b>&gt; is usually safe but may cause catastrophic failure on some systems.`,
	)
	Err500 = HTCode(500, "Internal Server Error",
		"The server encountered an error and could not complete your request.",
		"Make sure you are using a release version of manweb.",
	)
)

func HTCode(code int, name string, desc ...string) NetErr {
	return HTErr{code, name, desc}
}

func Is(err error) bool {
	if err == nil {
		return false
	}
	return true
}

type HTErr struct {
	Code int
	Name string
	Desc []string
}
type NetErr interface {
	Error() HTErr
}

func (e HTErr) Error() HTErr {
	return e
}
func (e HTErr) Title() string {
	return fmt.Sprintf("%d %s", e.Code, e.Name)
}
func (e HTErr) Content() string {
	s := fmt.Sprintf("<h1>%3d</h1><h2>%s</h2><br>\n", e.Code, e.Name)
	for d := range e.Desc {
		s += fmt.Sprintf("<p>%s</p><br>\n", e.Desc[d])
	}

	s += `<script>SetRawQuery()</script>
`
	return s
}
