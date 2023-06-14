package main

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func createSshConfig(username, keyFile string) *ssh.ClientConfig {
	// knownHosts := sshConfigPath("known_hosts")
	// knownHostsCallback, err := knownhosts.New(knownHosts)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	key, err := os.ReadFile(keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// create the signer for this private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("unable to parse private key: %v", err)
	}

	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoED25519,
		},
	}
}

func sshConfigPath(filename string) string {
	return filepath.Join(os.Getenv("HOME"), ".ssh", filename)
}

func main() {
	username := "bnm"
	keyFile := sshConfigPath("id_ed25519")
	addr := "localhost:2222"

	config := createSshConfig(username, keyFile)

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// each client conn can support multiple interactive sessions
	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	cmd := "uname -a"
	if err := session.Run(cmd); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
}
