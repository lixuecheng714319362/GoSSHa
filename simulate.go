package main

import (
	"fmt"
)

//安装环境
func installEnvProcess(hostname, user, passwd string) {
	scpPack(hostname, user, passwd)
}

func scpPack(hostname, user, passwd string) {
	//make docker dir
	cmds := "mkdir -p /home/" + user + "/fabric;"
	cmds += "cd /home/" + user + "/fabric;mkdir -p install;cd install;"
	cmds += "touch docker_pack_19.03.5.deb;touch containerd.io_1.2.6-3_amd64.deb; touch docker-ce-cli_19.03.5~3-0~ubuntu-xenial_amd64.deb"
	executeBatchSshCmd(cmds, hostname, user, passwd)

	fmt.Println("=====start scp docker=====")
	//传输docker安装包
	target := "/home/" + user + "/fabric/install/docker_pack_19.03.5.deb"
	source := "/home/lixuecheng/fabric/dockerPack/docker-ce_19.03.5_3-0_ubuntu-xenial_amd64.deb"
	executeCatByPwd(hostname, user, passwd, target, source)
	target = "/home/" + user + "/fabric/install/containerd.io_1.2.6-3_amd64.deb"
	source = "/home/lixuecheng/fabric/dockerPack/containerd.io_1.2.6-3_amd64.deb"
	executeCatByPwd(hostname, user, passwd, target, source)
	target = "/home/" + user + "/fabric/install/docker-ce-cli_19.03.5~3-0~ubuntu-xenial_amd64.deb"
	source = "/home/lixuecheng/fabric/dockerPack/docker-ce-cli_19.03.5~3-0~ubuntu-xenial_amd64.deb"
	executeCatByPwd(hostname, user, passwd, target, source)

	fmt.Println("=====start scp docker-compose=====")
	//make docker-compose file
	cmds1 := "cd /home/" + user + "/fabric/install;touch docker-compose;"
	executeBatchSshCmd(cmds1, hostname, user, passwd)
	//send docker-compose package
	target1 := "/home/" + user + "/fabric/install/docker-compose"
	source1 := "/home/lixuecheng/fabric/docker-compose_pack/docker-compose"
	executeCatByPwd(hostname, user, passwd, target1, source1)
}

func installDocker(hostname, user, passwd string) {
	dockercmds := "cd /home/" + user + "/fabric/install;sudo dpkg -i docker-ce-cli_19.03.5~3-0~ubuntu-xenial_amd64.deb;"
	dockercmds += "sudo dpkg -i containerd.io_1.2.6-3_amd64.deb;"
	dockercmds += "sudo dpkg -i docker_pack_19.03.5.deb;"
	executeBatchSshCmd(dockercmds, hostname, user, passwd)
}

func installDockerCompose(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/install;"
	cmds += "sudo mv docker-compose /usr/local/bin;"
	cmds += "sudo chmod +x /usr/local/bin/docker-compose"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}
