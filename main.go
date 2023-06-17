package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
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

func createSSHConfig(user, keyFilePath string) (*ssh.ClientConfig, error) {
	key, err := os.ReadFile(keyFilePath)

	if err != nil {
		return nil, err
	}
	// create signer for this private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func main() {
	// get home dir
	currUser, err := user.Current()
	if err != nil {
		log.Fatal("On get current user: ", err)
	}
	homeDir := currUser.HomeDir

	// args + defaults where necessary
	sshArg := ""
	var localPort uint16 = 0
	var remotePort uint16
	var keyFilePath string = filepath.Join(homeDir, ".ssh", "id_ed25519")

	// get cli args
	flag.StringVarP(&sshArg, "ssh", "s", "", "the ssh endpoint (user and address) with the format user@host:port. If port is omitted it defaults to 22")
	flag.StringVarP(&keyFilePath, "identity_file", "i", keyFilePath, "key file")
	flag.Uint16VarP(&localPort, "local_port", "l", localPort, "local port. Defaults to 0 ie random port is picked")
	flag.Uint16VarP(&remotePort, "remote_port", "r", 0, "remote port")
	flag.Parse()

	// check args
	user, host, port, err := parseSSHArg(sshArg)
	if err != nil {
		log.Fatal("On parse SSH arg: ", err)
	}
	if port == "" {
		port = "22"
	}

	if remotePort == 0 {
		log.Fatal("Remote port missing. Provide remote port")
	}

	// [localhost:<localPort>] <-> [<user>@<host>:<port>'] <-> [:<remotePort>]

	// get ssh config
	config, err := createSSHConfig(user, keyFilePath)
	if err != nil {
		log.Fatal("On create SSH config: ", err)
	}

	// create ssh sshClient
	sshAddr := host + ":" + port
	sshClient, err := ssh.Dial("tcp", sshAddr, config)
	if err != nil {
		log.Fatal("On dial to SSH endpoint: ", err)
	}
	defer sshClient.Close()
	log.Printf("Connected to SSH: %s@%s\n", user, sshAddr)

	// create session. Each client conn can support multiple interactive
	// sessions
	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatal("On create session: ", err)
	}
	defer session.Close()
	log.Println("Created SSH session")

	// listen on local port
	localAddr := fmt.Sprintf("localhost:%d", localPort)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("On create listener at '%d': %v", localPort, err)
	}
	defer listener.Close()
	log.Printf("Listening for local connections at: %s\n", localAddr)

	// ctx
	ctx, cancel := context.WithCancel(context.Background())

	// handle close signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("received signal: %s\n", sig.String())
		cancel()
		listener.Close()
	}()

	remoteAddr := fmt.Sprintf("localhost:%d", remotePort)
loop:
	for {
		localConn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// listener closed
				break loop
			}
			log.Fatal("On accept: ", err)
		}
		remoteConn, err := sshClient.Dial("tcp", remoteAddr)
		if err != nil {
			log.Fatalf("On conn to remote address (%s) via tunnel: %v", remoteAddr, err)
		}

		log.Printf("Tunnel established: local [%s] <-> remote [%s]\n",
			localConn.LocalAddr(), remoteAddr)

		go createTunnel(ctx, localConn, remoteConn)
	}
	log.Println("Exiting")
}

func createTunnel(ctx context.Context, local net.Conn, remote net.Conn) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		io.Copy(local, remote)
		cancel()
	}()

	go func() {
		io.Copy(remote, local)
		cancel()
	}()

	<-ctx.Done()
	remote.Close()
	local.Close()
	log.Printf("Tunnel closed for: %s\n", local.RemoteAddr())
}
