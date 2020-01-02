package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

//安装环境
func installEnvProcess(hostname, user, passwd string) {
	scpPack(hostname, user, passwd)
}

//send docker deb pack and docker-compose
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

//send Fabric Images
func sendFabricImages(hostname, user, passwd string) {
	//make images folder
	cmds := "cd /home/" + user + "/fabric;mkdir -p images;cd images;"
	cmds += "touch baseos.tar;touch ca.tar;touch ccenv.tar;touch couchdb.tar;touch kafka.tar;touch orderer.tar;touch peer.tar;touch zookeeper.tar;"
	cmds += "touch javaenv.tar.0;touch javaenv.tar.2;touch javaenv.tar.1;"
	cmds += "touch tools.tar.0;touch tools.tar.1;"
	executeBatchSshCmd(cmds, hostname, user, passwd)

	var wg sync.WaitGroup
	fmt.Println("=====start scp images=====")
	targetPrefix := "/home/" + user + "/fabric/images/"
	sourcePrefix := "/home/lixuecheng/fabric/fabric_images/"

	num := runtime.NumCPU()
	fmt.Printf("cpu num is %v\n", num)
	//GOMAXPROCS 设置可同时执行的最大CPU数
	runtime.GOMAXPROCS(num)
	var files1 = []string{"baseos.tar", "ca.tar", "ccenv.tar", "couchdb.tar", "kafka.tar", "orderer.tar", "peer.tar", "zookeeper.tar"}
	for i := 0; i < len(files1); i++ {
		wg.Add(1)
		go func(tmp string) {
			defer wg.Done()
			executeCatByPwd(hostname, user, passwd, targetPrefix+tmp, sourcePrefix+tmp)
		}(files1[i])
	}
	var files2 = []string{"javaenv.tar.0", "javaenv.tar.1", "javaenv.tar.2", "tools.tar.0", "tools.tar.1"}
	for i := 0; i < len(files2); i++ {
		wg.Add(1)
		go func(tmp string) {
			defer wg.Done()
			executeCatByPwd(hostname, user, passwd, targetPrefix+tmp, sourcePrefix+tmp)
		}(files2[i])
	}

	fmt.Printf("the length of map is %v\n", len(connectedHosts.v))
	wg.Wait()
}

var synWait sync.WaitGroup

func testR() {

	start := time.Now()
	for i := 1; i <= 20; i++ {
		synWait.Add(1)
		go testNum(1)
	}
	synWait.Wait()
	end := time.Now()
	fmt.Println(end.Sub(start).Seconds())
}

func testNum(num int) {
	fmt.Println("=====execute test num")
	for i := 1; i <= 10000000; i++ {
		num = num + i
		num = num - i
		num = num * i
		num = num / i
	}
	synWait.Done() // 相当于 synWait.Add(-1)
}
