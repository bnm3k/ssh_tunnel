package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"

	flag "github.com/spf13/pflag"
)

func parseSSHArg(arg string) (user string, host string, port string, err error) {
	parts := strings.SplitN(arg, "@", 2)
	if len(parts) != 2 {
		err = errors.New("invalid ssh arg")
		return
	}
	user = parts[0]
	if user == "" {
		err = errors.New("empty username")
		return
	}

	addr := parts[1]
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = addr
			err = nil
		}
	}

	return
}

func main() {
	// args
	ssh := ""
	var localPort uint16
	var remotePort uint16
	var keyFile string

	// get cli args
	flag.StringVarP(&ssh, "ssh", "s", "", "the ssh endpoint (user and address) with the format user@host:port. If port is omitted it defaults to 22")
	flag.StringVarP(&keyFile, "identity_file", "i", "~/.ssh/id_ed25519", "key file")
	flag.Uint16VarP(&localPort, "local_port", "l", 0, "local port. Defaults to 0 i.e. random port is picked")
	flag.Uint16VarP(&remotePort, "remote_port", "r", 0, "remote port")
	flag.Parse()

	// check args
	user, host, port, err := parseSSHArg(ssh)
	if err != nil {
		log.Fatal(err)
	}
	if port == "" {
		port = "22"
	}

	if remotePort == 0 {
		log.Fatal("provide remote port")
	}

	fmt.Printf("ssh tunnel: [localhost:%d] <-> [%s@%s:%s'] <-> [:%d]\n",
		localPort, user, host, port, remotePort)
}
