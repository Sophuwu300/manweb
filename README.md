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

## Extra Information
### Performance:
A query for all user or sudo  commands that list or organise information, shows 119 commands in 53ms.
Searching for all C libraries for parsing, shows 38 libraries in 48ms.

If I wish to find a command that configures a service, I would use:
network interfaces,

# Installation Using Apt

Simply run the following commands to add my repository and install the package. This will install the latest release and automatically update the server when new versions are released.
This will also make the server available as a systemd service, and start it automatically. You may still want to configure a user for the service as some manuals may be in user home directories.
This isn't common on most systems, so the default configuration should work out of the box in most cases.

```bash
curl https://cdn.sophuwu.com/deb/addrepo.sh | sudo sh
sudo apt install manweb
```

# Compiling From Source

## Dependencies
If you are not installing from apt, you will need to install the mandoc package.
* Ubuntu/Debian dependency installation: `sudo apt-get install mandoc -y`

If you wish to compile from source, you will need Go installed.\
* Golang installation instructions at [go.dev](https://go.dev/doc/install).


## Compiling The Binary

 ```bash
git clone "https://git.sophuwu.com/manweb"
cd manweb

# build binary
make build
make install_bin

# build a debian package and install it
make build_deb
make install_deb
```

# Using As Systemd Service:

The provided service file should work on most systems, but you may need to edit it to fit your needs.\
It will open a http server on port 8082 available through all network interfaces.\
You should change the `ListenAddr` variable to `127.0.0.1` and use a secure reverse proxy if you are on a public network.\
TLS and authentication are not implemented in this server.

### Variables in the service file:

Environment Variables:\
`HOSTNAME`: Used for displaying the hostname on the webpage.\
`ListenPort`: If unset, the server will default to 8082.\
`ListenAddr`: This should be changed if you are on a public network.\

### System Variables:

`User`: Reccomended to use your login user so the service can access your ~/.local man pages. But not required.\
`ExecStart`: The path to the manweb binary. If you installed it to /usr/local/bin, you can leave it as is.\
`WorkingDirectory`: This should be /tmp since the server doesn't need to write to disk.\

```sh
# to edit paths, users, and environment variables if needed
nano manweb.service 

# install the service file to systemd and load it
sudo install manweb.service /etc/systemd/system/manweb.service
sudo systemctl daemon-reload

# start the service and check its status
sudo systemctl start manweb.service
sudo systemctl status manweb.service

# to keep the service running after a reboot
sudo systemctl enable manweb.service

# to stop the service and disable it from restarting
sudo systemctl stop manweb.service
sudo systemctl disable manweb.service

# to edit the server configuration after installation
sudo systemctl edit manweb.service
sudo systemctl daemon-reload 
sudo systemctl reload-or-restart manweb.service
sudo systemctl status manweb.service
```

# Accessing the Web Interface

Open your web browser and navigate to `http://localhost:8082` if you are running the server locally or the remote server's IP address or hostname.\
To search with regex, you can use the search bar at the top of the page with `-r` at the beginning of the search term.\
To look into a specific section, you can add `-sN` to the search term where N is the section number.\
If no section is specified, the server will display with the same priority as the defualt `man` command.\
Glob patterns are also supported in the search bar if regex not enabled.\

## Example Usage:

- `ls*`: List all pages that begin with `ls`, including `ls`, `lsblk`, `lsmod`, etc.
- `-r ^ls`: Same as above but with regex. Usually more useful for with multiple queries and logical operators. Like finding any C++ reference to `std::string` and `std::vector`.
- `ls` or `ls.1`: Open the page for the `ls` user command. This is orignal man behavior.
- `-r ^ls -s1`: List all pages that begin with `ls` in section 1 (user/bin commands). Useful for finding commands that list any information without requiring sudo.
- `*config* -s8`: List pages for sudo commands containing keyword `config`. this can will show you commands that edit critical system files.  
- `vsftpd.5`: Open the manual page for vsftpd confuguration file if vsftpd is installed. This will show you how to configure the ftp server.
- `vsftpd.8`: Open the manual page for vsftpd executable if vsftpd is installed. This will show how to call the ftp server from the command line.

## License

MIT License

## Help and Support

I don't know how this git pull thing works. I will try if I see any issues. I've never collaborated on code before. If you have any suggestions, or questions about anything I've written, I would be happy to hear your thoughts.\
contant info: 
* discord: [@sophuwu](https://discord.com/users/sophuwu)
* email: [sophie@sophuwu.com](mailto:sophie@sophuwu.com)

