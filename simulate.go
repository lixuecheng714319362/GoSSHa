package main

//安装环境
func installEnvProcess(hostname, user, passwd string) {
	scpPack(hostname, user, passwd)
}

func scpPack(hostname, user, passwd string) {
	cmds := "mkdir -p /home/" + user + "/fabric;"
	cmds += "cd /home/" + user + "/fabric;mkdir -p install;cd install;touch docker_pack_19.03.5.deb"
	executeBatchSshCmd(cmds, hostname, user, passwd)
	//传输安装包
	target := "/home/" + user + "fabric/install/docker_pack_19.03.5.deb"
	executeCatByPwd(hostname, user, passwd, target, "D:/tmpHome/liteidex36.2.linux64-qt5.5.1.tar.gz")
}


