# go-piping-sshd
SSH server from anywhere with [Piping Server](https://github.com/nwtgck/piping-server)

## Thanks

SSH server implementation is highly based on <https://github.com/jpillora/sshd-lite/blob/master/server/server.go>. Thanks!

## CAUTION

Running this command exposes SSH server to public Piping Server by default. Someone who knows paths on the Piping Server can run any command on the machine. 

Here are possible use cases for using securely.

- The paths should be long or complex.
- Use self-host Piping Server in private network
- Run this command on a machine which everyone may access to.
- Use Piping Server with some authentication methods such as [basic auth](https://github.com/nwtgck/piping-server-basic-auth-docker-compose) or [JWT](https://github.com/nwtgck/jwt-piping-server).
