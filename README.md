# SSH Tunnel

Implements SSH Tunneling via Local Port forwarding. It forwards a connection
from the client to the SSH server host, which then forwards to the destination
host port.

## Usage

```
> ssh_tunnel  --help
Usage of ./ssh_tunnel:
  -i, --identity_file string   key file (default "/home/bnm/.ssh/id_ed25519")
  -l, --local_port uint16      local port. Defaults to 0 ie random port is picked
  -r, --remote_port uint16     remote port
  -s, --ssh string             the ssh endpoint (user and address) with the format user@host:port. Defaults is 22
```

Suppose you have an SSH server running at address `1.2.3.4` as user `admin`. At
the server, you've got a service that's running at `localhost:8000` that you
want to connect to via SSH port forwarding:

For example, at the server, run the following

```
nc -lvk 8000
```

At your client, start `ssh_tunnel`. It will listen for local connections at
`localhost:9000`:

```
ssh_tunnel -s admin@1.2.3.4 -i ~/.ssh/id_ed25519 -l 9000 -r 8000
```

You can then connect to the remote service:

```
nc -N localhost 9000
```

## Build

```
make build
```

## Install

```
make install
```

## TODO

- Validate server's host key

## Credits

- ["How to Set up SSH Tunneling (Port Forwarding)" - Linuxize](https://linuxize.com/post/how-to-setup-ssh-tunneling/)
- ["SSH port forwarding with Go" - Eli Bendersky](https://eli.thegreenplace.net/2022/ssh-port-forwarding-with-go/)
