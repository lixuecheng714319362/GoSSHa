package main

import "testing"

var (
	hostname    = "192.168.100.129"
	user_test   = "lixuecheng"
	passwd_test = "lixuecheng"
)

func TestInstallEnvProcess(t *testing.T) {

}

func TestScpPack(t *testing.T) {
	scpPack(hostname, user_test, passwd_test)
}

func TestInstallDocker(t *testing.T) {
	installDocker(hostname, user_test, passwd_test)
}

func TestInstallDockerCompose(t *testing.T) {
	installDockerCompose(hostname, user_test, passwd_test)
}
