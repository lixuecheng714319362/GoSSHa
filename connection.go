package main

import (
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
)

func connect2Server() {
	initialize(true)

}

func getConnByPwd() {
	check := func(err error, msg string) {
		if err != nil {
			log.Fatalf("%s error: %v", msg, err)
		}
	}

	client, err := ssh.Dial("tcp", "192.168.100.128:22", &ssh.ClientConfig{
		User: "lixuecheng",
		Auth: []ssh.AuthMethod{ssh.Password("lixuecheng")},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	check(err, "dial")

	session, err := client.NewSession()
	check(err, "new session")
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm", 25, 100, modes)
	check(err, "request pty")

	err = session.Shell()
	check(err, "start shell")

	err = session.Wait()
	check(err, "return")
}
//
//func getConnectionByPwd(hostname string) {
//	fmt.Println("=====enter get connection=====")
//	conn, ok := connectedHosts.Get(hostname)
//	if ok {
//		return
//	}
//
//	defer func() {
//		if msg := recover(); msg != nil {
//			err = errors.New("Panic: " + fmt.Sprint(msg))
//		}
//	}()
//
//	waitAgent()
//	conf, agentConn := makeConfig()
//	if agentConn != nil {
//		defer agentConn.Close()
//	}
//
//	defer releaseAgent()
//
//	port := "22"
//	str := strings.SplitN(hostname, ":", 2)
//	if len(str) == 2 {
//		hostname = str[0]
//		port = str[1]
//	}
//	fmt.Printf("the host is %v, and the port is %v\n", hostname, port)
//
//	conn, err = ssh.Dial("tcp", hostname+":"+port, conf)
//	if err != nil {
//		return
//	}
//
//	sendProxyReply(&ConnectionProgress{ConnectedHost: hostname})
//
//	if conn != nil {
//		fmt.Println("建立连接了？")
//	}
//	connectedHosts.Set(hostname, conn)
//	return
//}