package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const (
	defaultTimeout             = 30000  // default timeout for operations (in milliseconds)
	chunkSize                  = 655360 // chunk size in bytes for scp
	throughputSleepInterval    = 100    // how many milliseconds to sleep between writing "tickets" to channel in maxThroughputThread
	minChunks                  = 20     // minimum allowed count of chunks to be sent per sleep interval
	minThroughput              = chunkSize * minChunks * (1000 / throughputSleepInterval)
	maxOpensshAgentConnections = 128 // default connection backlog for openssh
)

var (
	user         string
	signers      []ssh.Signer
	keys         []string
	repliesChan  = make(chan interface{})
	requestsChan = make(chan *ProxyRequest)

	maxThroughputChan = make(chan bool, minChunks) // channel that is used for throughput limiting in scp

	maxThroughput uint64 // max throughput (for scp) in bytes per second

	agentConnChan      = make(chan chan bool) // channel for getting "ticket" for new agent connection
	agentConnFreeChan  = make(chan bool, 50)  // channel for freeing connections
	sshAuthSock        string
	maxConnections     uint64 // max concurrent ssh connections
	disconnectAfterUse bool   // close connection after each action

	connectedHosts = connHostsMap{v: make(map[string]*ssh.Client)}
)

type connHostsMap struct {
	mu sync.Mutex
	v  map[string]*ssh.Client
}

func (c *connHostsMap) Get(hostname string) (v *ssh.Client, ok bool) {
	c.mu.Lock()
	v, ok = c.v[hostname]
	c.mu.Unlock()
	return v, ok
}

func (c *connHostsMap) Set(hostname string, v *ssh.Client) {
	c.mu.Lock()
	c.v[hostname] = v
	c.mu.Unlock()
}

func (c *connHostsMap) Close(hostname string) error {
	c.mu.Lock()
	v, ok := c.v[hostname]
	delete(c.v, hostname)
	c.mu.Unlock()
	if !ok {
		return nil
	}

	return v.Close()
}

type (
	SshResult struct {
		hostname string
		stdout   string
		stderr   string
		err      error
	}

	ScpResult struct {
		hostname string
		err      error
	}

	ProxyRequest struct {
		Action        string
		Password      string // password for private key (only for Action == "password")
		Cmd           string // command to execute (only for Action == "ssh")
		Source        string // source file to copy (only for Action == "scp")
		Target        string // target file (only for Action == "scp")
		Hosts         []string
		Timeout       uint64 // timeout (in milliseconds), default is defaultTimeout
		MaxThroughput uint64 // max throughput (for scp) in bytes per second, default is no limit
	}

	Reply struct {
		Hostname string
		Stdout   string
		Stderr   string
		Success  bool
		ErrMsg   string
	}

	PasswordRequest struct {
		PasswordFor string
	}

	FinalReply struct {
		TotalTime     float64
		TimedOutHosts map[string]bool
	}

	ConnectionProgress struct {
		ConnectedHost string
	}

	UserError struct {
		IsCritical bool
		ErrorMsg   string
	}

	InitializeComplete struct {
		InitializeComplete bool
	}

	// executionRequest is used to send requests to the execution pool.
	executionRequest struct {
		Func func(string) *SshResult
		Host string
	}

	DisableReportConnectedHosts bool
	EnableReportConnectedHosts  bool
)

func reportErrorToUser(msg string) {
	repliesChan <- &UserError{ErrorMsg: msg}
}

func reportCriticalErrorToUser(msg string) {
	repliesChan <- &UserError{IsCritical: true, ErrorMsg: msg}
}

func waitAgent() {
	if sshAuthSock != "" {
		respChan := make(chan bool)
		agentConnChan <- respChan
		<-respChan
	}
}

func releaseAgent() {
	if sshAuthSock != "" {
		agentConnFreeChan <- true
	}
}

func makeConfig() (config *ssh.ClientConfig, agentUnixSock net.Conn) {
	fmt.Println("=============make config=============")
	clientAuth := []ssh.AuthMethod{}

	var err error

	if sshAuthSock != "" {
		fmt.Println("sshAuthSock is not empty!!!!!!")
		for {
			agentUnixSock, err = net.Dial("unix", sshAuthSock)

			if err != nil {
				netErr := err.(net.Error)
				if netErr.Temporary() {
					time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
					continue
				}

				reportErrorToUser("Cannot open connection to SSH agent: " + netErr.Error())
			} else {
				authAgent := ssh.PublicKeysCallback(agent.NewClient(agentUnixSock).Signers)
				clientAuth = append(clientAuth, authAgent)
			}

			break
		}
	}

	fmt.Printf("the length of signers in makeConfig is %v\n", len(signers))
	if len(signers) > 0 {
		clientAuth = append(clientAuth, ssh.PublicKeys(signers...))
	}

	config = &ssh.ClientConfig{
		User:            user,
		Auth:            clientAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return
}

func makeSigner(keyname string) (signer ssh.Signer, err error) {
	fp, err := os.Open(keyname)
	if err != nil {
		if !os.IsNotExist(err) {
			reportErrorToUser("Could not parse " + keyname + ": " + err.Error())
		}
		return
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		reportErrorToUser("Could not read " + keyname + ": " + err.Error())
		return
	}

	if bytes.Contains(buf, []byte("ENCRYPTED")) {
		var (
			tmpfp *os.File
			out   []byte
		)

		tmpfp, err = ioutil.TempFile("", "key")
		if err != nil {
			reportErrorToUser("Could not create temporary file: " + err.Error())
			return
		}

		tmpName := tmpfp.Name()

		defer func() { tmpfp.Close(); os.Remove(tmpName) }()

		_, err = tmpfp.Write(buf)

		if err != nil {
			reportErrorToUser("Could not write encrypted key contents to temporary file: " + err.Error())
			return
		}

		err = tmpfp.Close()
		if err != nil {
			reportErrorToUser("Could not close temporary file: " + err.Error())
			return
		}

		repliesChan <- &PasswordRequest{PasswordFor: keyname}
		response := <-requestsChan

		if response.Password == "" {
			reportErrorToUser("No passphrase supplied in request for " + keyname)
			err = errors.New("No passphrase supplied")
			return
		}

		cmd := exec.Command("ssh-keygen", "-f", tmpName, "-N", "", "-P", response.Password, "-p")
		out, err = cmd.CombinedOutput()
		if err != nil {
			reportErrorToUser(strings.TrimSpace(string(out)))
			return
		}

		tmpfp, err = os.Open(tmpName)
		if err != nil {
			reportErrorToUser("Cannot open back " + tmpName)
			return
		}

		buf, err = ioutil.ReadAll(tmpfp)
		if err != nil {
			return
		}

		tmpfp.Close()
		os.Remove(tmpName)
	}

	signer, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		reportErrorToUser("Could not parse " + keyname + ": " + err.Error())
		return
	}

	return
}

func makeSigners() {
	signers = []ssh.Signer{}

	for _, keyname := range keys {
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}
	fmt.Printf("the length of signer in makeSingers is %v\n", len(signers))
}

func getConnection(hostname string) (conn *ssh.Client, err error) {
	fmt.Println("=====enter get connection=====")
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
	conf, agentConn := makeConfig()
	if agentConn != nil {
		defer agentConn.Close()
	}

	defer releaseAgent()

	port := "22"
	str := strings.SplitN(hostname, ":", 2)
	if len(str) == 2 {
		hostname = str[0]
		port = str[1]
	}
	fmt.Printf("the host is %v, and the port is %v\n", hostname, port)

	conn, err = ssh.Dial("tcp", hostname+":"+port, conf)
	if err != nil {
		return
	}

	sendProxyReply(&ConnectionProgress{ConnectedHost: hostname})

	if conn != nil {
		fmt.Println("建立连接了？")
	}
	connectedHosts.Set(hostname, conn)
	return
}

func uploadFile(target string, contents []byte, hostname string) (stdout, stderr string, err error) {
	conn, err := getConnection(hostname)
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
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Start(cmd)
	if err != nil {
		return
	}

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

	return
}

func executeCmd(cmd string, hostname string) (stdout, stderr string, err error) {
	conn, err := getConnection(hostname)
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

	return
}

// do not allow more than maxConn simultaneous ssh-agent connections
func agentConnectionManagerThread(maxConn uint64) {
	freeConn := maxConn // free connections count

	for {
		reqCh := agentConnChan
		freeCh := agentConnFreeChan

		if freeConn <= 0 {
			reqCh = nil
		}

		select {
		case respChan := <-reqCh:
			freeConn--
			respChan <- true
		case <-freeCh:
			freeConn++
		}
	}
}

func initialize(internalInput bool) {
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

	keysPath := os.Getenv("TMP_HOME")
	keys = []string{keysPath + "/.ssh/id_rsa", keysPath + "/.ssh/id_dsa", keysPath + "/.ssh/id_ecdsa"}

	if pubKey != "" {
		if strings.HasSuffix(pubKey, ".pub") {
			pubKey = strings.TrimSuffix(pubKey, ".pub")
		}

		keys = append(keys, pubKey)
	}

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

func jsonReplierThread() {
	fmt.Println("======enter jsonReplierThread=======")
	connectionReporting := true

	for {
		reply := <-repliesChan

		switch reply.(type) {
		case DisableReportConnectedHosts:
			connectionReporting = false
			continue

		case EnableReportConnectedHosts:
			connectionReporting = true
			continue

		case *ConnectionProgress:
			if !connectionReporting {
				continue
			}
		}

		buf, err := json.Marshal(reply)
		if err != nil {
			panic("Could not marshal json reply: " + err.Error())
		}

		if buf[0] == '{' {
			typeStr := strings.TrimPrefix(fmt.Sprintf("%T", reply), "*main.")
			fmt.Printf("{\"Type\":\"%s\",%s}\n", typeStr, buf[1:len(buf)-1])
		} else {
			fmt.Println(string(buf))
		}
	}
}

func sendProxyReply(response interface{}) {
	repliesChan <- response
}

func maxThroughputThread() {
	for {
		throughput := atomic.LoadUint64(&maxThroughput)

		// how many chunks can be sent in specified time interval
		chunks := throughput / chunkSize * throughputSleepInterval / 1000

		if chunks < minChunks {
			chunks = minChunks
		}

		for i := uint64(0); i < chunks; i++ {
			maxThroughputChan <- true
		}

		if throughput > 0 {
			time.Sleep(throughputSleepInterval * time.Millisecond)
		}
	}
}

func getExecFunc(msg *ProxyRequest) func(string) *SshResult {
	if msg.Action == "ssh" {
		if msg.Cmd == "" {
			reportCriticalErrorToUser("Empty 'Cmd'")
			return nil
		}

		return func(hostname string) *SshResult {
			stdout, stderr, err := executeCmd(msg.Cmd, hostname)
			return &SshResult{hostname: hostname, stdout: stdout, stderr: stderr, err: err}
		}
	} else if msg.Action == "scp" {
		if msg.Source == "" {
			reportCriticalErrorToUser("Empty 'Source'")
			return nil
		}

		if msg.Target == "" {
			reportCriticalErrorToUser("Empty 'Target'")
			return nil
		}

		if msg.MaxThroughput > 0 && msg.MaxThroughput < minThroughput {
			reportErrorToUser(fmt.Sprint("Minimal supported throughput is ", minThroughput, " Bps"))
		}

		atomic.StoreUint64(&msg.MaxThroughput, maxThroughput)

		fp, err := os.Open(msg.Source)
		if err != nil {
			reportCriticalErrorToUser(err.Error())
			return nil
		}

		defer fp.Close()

		contents, err := ioutil.ReadAll(fp)
		if err != nil {
			reportCriticalErrorToUser("Cannot read " + msg.Source + " contents: " + err.Error())
			return nil
		}

		return func(hostname string) *SshResult {
			stdout, stderr, err := uploadFile(msg.Target, contents, hostname)
			return &SshResult{hostname: hostname, stdout: stdout, stderr: stderr, err: err}
		}
	}

	reportCriticalErrorToUser(fmt.Sprintf("Unsupported action: %s", msg.Action))
	return nil
}

func runAction(msg *ProxyRequest) {
	fmt.Println("=======enter run action =========")
	execFunc := getExecFunc(msg)
	if execFunc == nil {
		return
	}

	timeout := uint64(defaultTimeout)

	if msg.Timeout > 0 {
		timeout = msg.Timeout
	}

	startTime := time.Now().UnixNano()

	responseChannel := make(chan *SshResult, len(msg.Hosts))
	timeoutChannel := time.After(time.Millisecond * time.Duration(timeout))

	timedOutHosts := make(map[string]bool)
	sendProxyReply(EnableReportConnectedHosts(true))

	maxConcurrency := uint64(len(msg.Hosts))
	if maxConnections > 0 {
		maxConcurrency = maxConnections
	}
	maxConcurrencyCh := make(chan struct{}, maxConcurrency)

	for _, h := range msg.Hosts {
		fmt.Println("go func exec cmd")
		go func(h string) {
			maxConcurrencyCh <- struct{}{}
			defer func() { <-maxConcurrencyCh }()
			responseChannel <- execFunc(h)
		}(h)
	}

	for i := 0; i < len(msg.Hosts); i++ {
		select {
		case <-timeoutChannel:
			goto finish
		case msg := <-responseChannel:
			delete(timedOutHosts, msg.hostname)
			success := true
			errMsg := ""
			if msg.err != nil {
				errMsg = msg.err.Error()
				success = false
			}
			sendProxyReply(&Reply{Hostname: msg.hostname, Stdout: msg.stdout, Stderr: msg.stderr, ErrMsg: errMsg, Success: success})
		}
	}

finish:
	for hostname := range timedOutHosts {
		connectedHosts.Close(hostname)
	}

	sendProxyReply(DisableReportConnectedHosts(true))

	sendProxyReply(&FinalReply{TotalTime: float64(time.Now().UnixNano()-startTime) / 1e9, TimedOutHosts: timedOutHosts})
}

func inputDecoder() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := new(ProxyRequest)

		line := scanner.Bytes()
		err := json.Unmarshal(line, msg)
		if err != nil {
			reportCriticalErrorToUser("Cannot parse JSON: " + err.Error())
			continue
		}
		fmt.Printf("this time, msg is %v\n", msg)

		requestsChan <- msg
	}

	if err := scanner.Err(); err != nil {
		reportCriticalErrorToUser("Error reading stdin: " + err.Error())
	}

	close(requestsChan)
}

func runProxy() {
	for msg := range requestsChan {
		switch {
		case msg.Action == "ssh" || msg.Action == "scp":
			fmt.Println("runProxy, action is ssh or scp")
			runAction(msg)
		default:
			reportCriticalErrorToUser("Unsupported action: " + msg.Action)
		}
	}
}

func main() {
	initialize(false)
	sendProxyReply(&InitializeComplete{InitializeComplete: true})
	runProxy()
}
