package main

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

func connect2Server() {
	initialize(true)

}

//buildMsg build host info etc.
func buildMsg(hostname, action, cmd string) {

}

//执行cat命令将源文件内容输入到目标文件中
func executeCatByPwd(hostname, user, password, target, source string) (stdout, stderr string, err error) {
	conn, err := getConnectionByPwd(hostname, user, password)
	if err != nil {
		return
	}

	session, err := conn.NewSession()
	if err != nil {
		return
	}
	if disconnectAfterUse {
		defer connectedHosts.Close(hostname)
	}
	defer session.Close()

	cmd := "cat >'" + strings.Replace(target, "'", "'\\''", -1) + "'"
	fmt.Printf("cmd is %v\n", cmd)
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	fmt.Println("start run cmd")
	err = session.Start(cmd)
	if err != nil {
		return
	}

	contents := readSourceFile(source)
	for start, maxEnd := 0, len(contents); start < maxEnd; start += chunkSize {
		<-maxThroughputChan

		end := start + chunkSize
		if end > maxEnd {
			end = maxEnd
		}
		_, err = stdinPipe.Write(contents[start:end])
		if err != nil {
			return
		}
	}

	err = stdinPipe.Close()
	if err != nil {
		return
	}

	err = session.Wait()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	fmt.Printf("stdout is %v\n", stdout)
	fmt.Printf("stderr is %v\n", stderr)
	return
}

//read source file to []byte
func readSourceFile(source string) []byte {
	fp, err := os.Open(source)
	if err != nil {
		reportCriticalErrorToUser(err.Error())
		return nil
	}

	defer fp.Close()

	contents, err := ioutil.ReadAll(fp)
	if err != nil {
		reportCriticalErrorToUser("Cannot read " + source + " contents: " + err.Error())
		return nil
	}
	return contents
}

//执行ssh命令
func executeCmdByPwd(cmd string, hostname, user, password string) (stdout, stderr string, err error) {
	conn, err := getConnectionByPwd(hostname, user, password)
	if err != nil {
		return
	}

	fmt.Println("start new session")
	session, err := conn.NewSession()
	if err != nil {
		return
	}
	if disconnectAfterUse {
		defer connectedHosts.Close(hostname)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	fmt.Println("start run cmd")
	err = session.Run(cmd)

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	fmt.Printf("stdout is %v\n", stdout)
	fmt.Printf("stderr is %v\n", stderr)
	return
}

//通过密码登录建立ssh连接
func getConnectionByPwd(hostname, user, password string) (conn *ssh.Client, err error) {
	conn, ok := connectedHosts.Get(hostname)
	if ok {
		return
	}

	defer func() {
		if msg := recover(); msg != nil {
			err = errors.New("Panic: " + fmt.Sprint(msg))
		}
	}()

	waitAgent()
	defer releaseAgent()

	port := "22"
	str := strings.SplitN(hostname, ":", 2)
	if len(str) == 2 {
		hostname = str[0]
		port = str[1]
	}
	fmt.Printf("the host is %v, and the port is %v\n", hostname, port)

	conn, err = ssh.Dial("tcp", hostname + ":" + port, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	if err != nil {
		return
	}

	//sendProxyReply(&ConnectionProgress{ConnectedHost: hostname})

	connectedHosts.Set(hostname, conn)
	return
}

func connectByPwd(hostName, user, password string) {
	check := func(err error, msg string) {
		if err != nil {
			log.Fatalf("%s error: %v", msg, err)
		}
	}

	client, err := ssh.Dial("tcp", hostName + ":22", &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
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