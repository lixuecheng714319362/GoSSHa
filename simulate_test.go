package main

import (
	"fmt"
	"testing"
	"time"
)

var (
	hostname    = "192.168.100.132"
	user_test   = "lixuecheng"
	passwd_test = "lixuecheng"
)

func TestScpPack(t *testing.T) {
	scpPack(hostname, user_test, passwd_test)
}

func TestInstallDocker(t *testing.T) {
	installDocker(hostname, user_test, passwd_test)
}

func TestInstallDockerCompose(t *testing.T) {
	installDockerCompose(hostname, user_test, passwd_test)
}

func TestSendFabricImages(t *testing.T) {
	sendFabricImages(hostname, user_test, passwd_test)
}

func TestTarFabricImages(t *testing.T) {
	tarFabricImages(hostname, user_test, passwd_test)
}

func TestLoadDockerImages(t *testing.T) {
	loadDockerImages(hostname, user_test, passwd_test)
}

func TestGenCrypto(t *testing.T) {
	genCrypto(hostname, user_test, passwd_test)
}

func TestGenGenesisBlock(t *testing.T) {
	genGenesisBlock(hostname, user_test, passwd_test)
}

func TestStartFabricNetwork(t *testing.T) {
	startFabricNetwork(hostname, user_test, passwd_test)
}

func TestMakeChannel(t *testing.T) {
	makeChannel(hostname, user_test, passwd_test)
}

func TestInstallChaincode(t *testing.T) {
	installChaincode(hostname, user_test, passwd_test)
}

func TestInstantiateChaincode(t *testing.T) {
	instantiateChaincode(hostname, user_test, passwd_test)
}

func TestScpImagesViaInternet(t *testing.T) {
	start := time.Now().Unix()
	scpImagesViaInternet("192.168.4.174", "root", "shuqinkeji")
	end := time.Now().Unix()
	fmt.Printf("use time %v seconds\n", end-start)
}

func TestSshByKey(t *testing.T) {
	stdout, stderr, err := sshByKey(hostname, "lixuecheng", "/home/lixuecheng/.ssh", "uptime")
	if err != nil {
		fmt.Printf("error is %v\n", err)
	}
	fmt.Printf("stdout is %v\nstderr is %v\n", stdout, stderr)
}

func TestExecuteCatByKey(t *testing.T) {
	executeCatByKey(hostname, "lixuecheng", "/home/lixuecheng/.ssh", "/home/lixuecheng/fabric/images/tools.tar", "/home/lixuecheng/fabric/fabric_images/tools.tar")
}