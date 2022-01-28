# go-piping-sshd
SSH server from anywhere with [Piping Server](https://github.com/nwtgck/piping-server)

## Thanks

SSH server implementation for shell is highly based on <https://github.com/jpillora/sshd-lite/blob/master/server/server.go>. Thanks!

## Install for macOS
```bash
brew install nwtgck/piping-sshd/piping-sshd
```

## Install for Ubuntu
```bash
wget https://github.com/nwtgck/go-piping-sshd/releases/download/v0.7.0/piping-sshd-0.7.0-linux-amd64.deb
dpkg -i piping-sshd-0.7.0-linux-amd64.deb 
```

Get more executables in the [releases](https://github.com/nwtgck/go-piping-sshd/releases).

## Usage

Run the command below in machine A which will be controlled.

```bash
piping-sshd mypath --password=changeme
```

Run the command below in another machine B which will control the machine A.

```bash
curl -NsS https://ppng.io/mypath/sc | socat TCP-LISTEN:2022,reuseaddr - | curl -NsST - https://ppng.io/mypath/cs
```

Run the command below in the machine B to ssh. The password is `changeme`.

```bash
ssh-keygen -R [localhost]:2022; ssh -p 2022 dummy@localhost
```

## Help

```txt
SSH server from anywhere with Piping Server

Usage:
  piping-sshd [flags]

Flags:
      --allow-empty-password      Allows to run SSH server with empty password
      --dns-server string         DNS server (e.g. 1.1.1.1:53)
  -H, --header stringArray        HTTP header
  -h, --help                      help for piping-sshd
      --http-read-buf-size int    HTTP read-buffer size in bytes (default 4096)
      --http-write-buf-size int   HTTP write-buffer size in bytes (default 4096)
  -k, --insecure                  Allow insecure server connections when using SSL
  -p, --password string           SSH user password
  -s, --server string             Piping Server URL (default "https://ppng.io")
      --shell string              Shell
  -u, --user string               SSH user name
  -v, --version                   show version
      --yamux                     Multiplex connection by yamux
```

## CAUTION

Running this command exposes SSH server to public Piping Server. Someone who knows paths on the Piping Server can run any command on the machine. 

Here are ways for using securely.

- Specify `--password` or `-p` in long or complex password.
- The paths should be long or complex.
- Use self-host Piping Server in private network
- Run this command on a machine which everyone may access to.
- Use Piping Server with some authentication methods such as [basic auth](https://github.com/nwtgck/piping-server-basic-auth-docker-compose) or [JWT](https://github.com/nwtgck/jwt-piping-server).
