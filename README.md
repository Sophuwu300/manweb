# Man Web
A Service to show man pages on the web.
It allows for easy viewing and advanced searching of man pages in a web browser.
It uses the current system's manpages, so you will always have the correct documentation
to all the commands and libraries installed on your system.

Written in GoLang, so the only dependencies are `man` and `mandoc` packages.
To use tldr, `git` will be required to download the tldr pages.

## Features

- Convert any man page into HTML, formatted for easy reading. With a dark theme.
- Search functionality to find specific man pages. Wildcards and regex are supported.
- Hyperlinked pages for easy navigation for any valid reference in the document, and the ability to open the page in a new tab.
- Will display all man pages in the manpath, including pages where `man2html` and `man -Thtml` fail.
- Filter by page function or section: 1=commands, 3=C/C++ Refs, 5,7=config/format, 8=sudo commands, etc.
- Able to correctly interpret and display incorrectly formatted man pages, to a degree.
- Auto updates man pages when new packages are installed or removed using standard installation methods.

# Installation Using Apt

Simply run the following commands to add my repository and install the package. This will install the latest release and automatically update the server when new versions are released.
This will also make the server available as a systemd service, and start it automatically. 

You can read [addrepo.sh.txt](https://cdn.sophuwu.com/deb/addrepo.sh.txt) in your browser.

```bash
curl https://cdn.sophuwu.com/deb/addrepo.sh | sudo sh
sudo apt install manweb
```

# Compiling From Source

## Dependencies
To compile the binary, the `go` compiler is required [go.dev](https://go.dev/doc/install), `make`
is recommended for compilation.

The following packages are required
* `mandoc`
* `make`
* `git` - Optional, for tldr support
```bash
sudo apt install mandoc make git
```
The go compiler is required to compile
*  [go.dev](https://go.dev/doc/install).


## Compiling The Binary

 ```bash
git clone "https://git.sophuwu.com/manweb"
cd manweb

# build binary
make build
make install_bin
```

# Using As Systemd Service:

If you used the installer script, the service will be automatically installed. You just need to enable it.
```bash
sudo systemctl daemon-reload
sudo systemctl enable manweb
sudo systemctl start manweb
```

# Configuring The Server
The server can be configured by editing the `/etc/manweb/manweb.conf` file.
Simply set the options how you like, and restart the service.
```bash
sudo nano /etc/manweb/manweb.conf
sudo systemctl restart manweb
```

# Setting A Password
To set a password for the web interface, you must enable the `require_auth` option in `/etc/manweb/manweb.conf`.
Generally, there is no need to change the `auth_file` option.

Once you have enabled `require_auth`, you can set a password `manweb-passwd` while the service is not running.
```bash
sudo systemctl stop manweb
sudo manweb-passwd username
sudo systemctl start manweb
```
The `manweb-passwd` command will prompt you for a password.
If you enter a password, it will be set for the specified user.
If you leave the password blank, it will remove the user from the authentication database.

If you set the `require_auth` option to `no`, the server will ignore the authentication database and allow access to the web interface without a password.
This is the default behavior.
'

# Accessing the Web Interface

If you have installled the service and are running with default settings, you should be able to access 
the web interface on [http://localhost:8082](http://localhost:8082).
By default, the web interface binds to `0.0.0.0:8082`.

### Accessibility Options
* For the best reading experience, the web interface has 3 themes: light, dark, and yellow filter.
* The text contrast can be adjusted to make it easier to read.
* The settings tab has an adjustable font scale, allowing easier uniform scaling on any device.

### Searching
Regex and wildcards are supported in the search bar.
You can also filter by section, or page function.

* `ls*`: List all pages that begin with `ls`, including `ls`, `lsblk`, `lsmod`, etc.
* `-r ^ls`: Same as above but with regex.
* `ls` or `ls.1`: Open the page for the `ls` user command. This is orignal man behavior.
* `-r ^ls -s1`: List all pages that begin with `ls` in section 1 (user/bin commands). Useful for finding commands that list any information without requiring sudo.
* `*config* -s8`: List pages for sudo commands containing keyword `config`. this can will show you commands that edit critical system files.  
* `vsftpd.5`: Open the manual page for vsftpd confuguration file if vsftpd is installed. This will show you how to configure the ftp server.
* `vsftpd.8`: Open the manual page for vsftpd executable if vsftpd is installed. This will show how to call the ftp server from the command line.

## License

MIT License

## Help and Support

I don't know how this git pull thing works. I will try if I see any issues. I've never collaborated on code before. If you have any suggestions, or questions about anything I've written, I would be happy to hear your thoughts.\
contant info: 
* discord: [@sophuwu](https://discord.com/users/sophuwu)
* email: [sophie@sophuwu.com](mailto:sophie@sophuwu.com)


## Gallery
[Video of Navigation](https://cdn.sophuwu.com/img/manhttpd-demo/manhttpd-v2025-07.mp4)

<img src="https://cdn.sophuwu.com/img/manhttpd-demo/page.png" width="31%">
<img src="https://cdn.sophuwu.com/img/manhttpd-demo/search.png" width="33%">
<img src="https://cdn.sophuwu.com/img/manhttpd-demo/stats.png" width="33%">
<img src="https://cdn.sophuwu.com/img/manhttpd-demo/dark.png" width="31%"> 
<img src="https://cdn.sophuwu.com/img/manhttpd-demo/light.png" width="33%"> 
<img src="https://cdn.sophuwu.com/img/manhttpd-demo/yellow.png" width="33%">
