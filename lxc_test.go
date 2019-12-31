package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

func myInit(internalInput bool) {
	fmt.Println("========enter myinit========")
	var (
		pubKey              string
		maxAgentConnections uint64
	)

	flag.StringVar(&pubKey, "i", "", "Optional path to public key to use")
	flag.StringVar(&user, "l", os.Getenv("LOGNAME"), "Optional login name")
	flag.Uint64Var(&maxAgentConnections, "c", maxOpensshAgentConnections, "Maximum simultaneous ssh-agent connections")
	flag.BoolVar(&disconnectAfterUse, "d", false, "Disconnect after each action")
	flag.Uint64Var(&maxConnections, "m", 0, "Maximum simultaneous connections")
	flag.Parse()

	//keys = []string{os.Getenv("HOME") + "/.ssh/id_rsa", os.Getenv("HOME") + "/.ssh/id_dsa", os.Getenv("HOME") + "/.ssh/id_ecdsa"}
	//
	//if pubKey != "" {
	//	if strings.HasSuffix(pubKey, ".pub") {
	//		pubKey = strings.TrimSuffix(pubKey, ".pub")
	//	}
	//
	//	keys = append(keys, pubKey)
	//}
	keys = []string{"C:/Users/71431/.ssh/id_rsa"}

	fmt.Println(keys)
	sshAuthSock = os.Getenv("SSH_AUTH_SOCK")

	if sshAuthSock != "" {
		go agentConnectionManagerThread(maxAgentConnections)
	}

	if !internalInput {
		fmt.Println("!internalInput")
		go inputDecoder()
		go jsonReplierThread()
	}

	go maxThroughputThread()

	makeSigners()
}

func TestGetConnByPwd(t *testing.T) {
	getConnByPwd()
}