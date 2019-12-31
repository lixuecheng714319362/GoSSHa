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
	connectByPwd("192.168.100.128", "lixuecheng", "lixuecheng")
}

func TestExecuteCmdByPwd(t *testing.T) {
	executeCmdByPwd("uptime", "192.168.100.128", "lixuecheng", "lixuecheng")
}

func TestScp(t *testing.T) {
	executeCmdByPwd("touch tmp_file", "192.168.100.128", "lixuecheng", "lixuecheng")
	executeCatByPwd("192.168.100.128", "lixuecheng", "lixuecheng", "/home/lixuecheng/tmp_file", "D:\\tmpHome\\.ssh\\known_hosts")
	fmt.Println("=====upload tmp file success")
	executeCmdByPwd("rm -rf /home/lixuecheng/tmp_file", "192.168.100.128", "lixuecheng", "lixuecheng")

	fmt.Println("==================")

	executeCmdByPwd("touch liteidex36.2.linux64-qt5.5.1.tar.gz", "192.168.100.128", "lixuecheng", "lixuecheng")
	executeCatByPwd("192.168.100.128", "lixuecheng", "lixuecheng", "/home/lixuecheng/liteidex36.2.linux64-qt5.5.1.tar.gz", "D:\\tmpHome\\liteidex36.2.linux64-qt5.5.1.tar.gz")
	executeCmdByPwd("tar -zxvf liteidex36.2.linux64-qt5.5.1.tar.gz -C /home/lixuecheng", "192.168.100.128", "lixuecheng", "lixuecheng")
	fmt.Println("=====upload tar success")
	executeCmdByPwd("rm -rf /home/lixuecheng/liteidex36.2.linux64-qt5.5.1.tar.gz", "192.168.100.128", "lixuecheng", "lixuecheng")
	executeCmdByPwd("rm -rf /home/lixuecheng/liteide", "192.168.100.128", "lixuecheng", "lixuecheng")
}

func TestExecuteScpByPwd(t *testing.T) {
	executeCatByPwd("192.168.100.128", "lixuecheng", "lixuecheng", "/home/lixuecheng/tmp_file", "D:\\tmpHome\\.ssh\\known_hosts")
	executeCmdByPwd("rm -rf /home/lixuecheng/tmp_file", "192.168.100.128", "lixuecheng", "lixuecheng")
}

func buildCmdInfos() (cmdInfo1, cmdInfo2, cmdInfo3 []CmdInfo) {
	var ret1, ret2, ret3 []CmdInfo
	info1 := CmdInfo{
		Cmd:        "touch liteidex36.2.linux64-qt5.5.1.tar.gz",
		Action:     "ssh",
		TargetFile: "",
		SourceFile: "",
	}
	info2 := CmdInfo{
		Cmd:        "",
		Action:     "scp",
		TargetFile: "/home/lixuecheng/liteidex36.2.linux64-qt5.5.1.tar.gz",
		SourceFile: "D:\\tmpHome\\liteidex36.2.linux64-qt5.5.1.tar.gz",
	}
	info3 := CmdInfo{
		Cmd:        "tar -zxvf liteidex36.2.linux64-qt5.5.1.tar.gz -C /home/lixuecheng",
		Action:     "ssh",
		TargetFile: "",
		SourceFile: "",
	}
	info4 := CmdInfo{
		Cmd:        "rm -rf /home/lixuecheng/liteidex36.2.linux64-qt5.5.1.tar.gz",
		Action:     "ssh",
		TargetFile: "",
		SourceFile: "",
	}
	info5 := CmdInfo{
		Cmd:        "rm -rf /home/lixuecheng/liteide",
		Action:     "ssh",
		TargetFile: "",
		SourceFile: "",
	}

	ret1 = append(ret1, info1)
	ret2 = append(ret2, info2)
	ret3 = append(ret3, info3)
	ret3 = append(ret3, info4)
	ret3 = append(ret3, info5)
	return ret1, ret2, ret3
}

func TestSshAndScp(t *testing.T) {
	cmdInfos1, cmdInfos2, cmdInfos3:= buildCmdInfos()
	executeLinuxCmd(cmdInfos1, "192.168.100.128", "lixuecheng", "lixuecheng")
	executeLinuxCmd(cmdInfos2, "192.168.100.128", "lixuecheng", "lixuecheng")
	executeLinuxCmd(cmdInfos3, "192.168.100.128", "lixuecheng", "lixuecheng")
}