package main

import (
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
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
	cmds += "touch javaenv.tar;"
	cmds += "touch tools.tar;"
	executeBatchSshCmd(cmds, hostname, user, passwd)

	var wg sync.WaitGroup
	fmt.Println("=====start scp images=====")
	targetPrefix := "/home/" + user + "/fabric/images/"
	sourcePrefix := "/home/lixuecheng/fabric/fabric_images/"

	// num := runtime.NumCPU()
	// fmt.Printf("cpu num is %v\n", num)
	// //GOMAXPROCS 设置可同时执行的最大CPU数
	// runtime.GOMAXPROCS(num)
	var files1 = []string{"baseos.tar", "ca.tar", "ccenv.tar", "couchdb.tar", "kafka.tar", "orderer.tar", "peer.tar", "zookeeper.tar"}
	for i := 0; i < len(files1); i++ {
		wg.Add(1)
		go func(tmp string) {
			defer wg.Done()
			executeCatByPwd(hostname, user, passwd, targetPrefix+tmp, sourcePrefix+tmp)
		}(files1[i])
		// executeCatByPwd(hostname, user, passwd, targetPrefix+files1[i], sourcePrefix+files1[i])
	}
	// time.Sleep(2000)
	var files2 = []string{"javaenv.tar.0", "javaenv.tar.1", "javaenv.tar.2", "tools.tar.0", "tools.tar.1"}
	for i := 0; i < len(files2); i++ {
		//executeCatByPwd(hostname, user, passwd, targetPrefix+files2[i], sourcePrefix+files2[i])
		wg.Add(1)
		go func(tmp string) {
			defer wg.Done()
			executeCatByPwd(hostname, user, passwd, targetPrefix+tmp, sourcePrefix+tmp)
		}(files2[i])
	}

	wg.Wait()
}

func scpImagesViaInternet(hostname, user, passwd string) {
	cmds := "cd /data/ssh_scp_test/images;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
	fmt.Println("=====start scp images=====")
	targetPrefix := "/data/ssh_scp_test/images/"
	sourcePrefix := "/home/lixuecheng/fabric/fabric_images/"
	var files1 = []string{"baseos.tar", "ca.tar", "ccenv.tar", "couchdb.tar", "kafka.tar", "orderer.tar", "peer.tar", "zookeeper.tar", "javaenv.tar", "tools.tar"}
	for i := 0; i < len(files1); i++ {
		executeCatByPwd(hostname, user, passwd, targetPrefix+files1[i], sourcePrefix+files1[i])
	}
}

//tar -zcvf farbric images
func tarFabricImages(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/images;mkdir -p images_untar;cd images_untar;"
	cmds += "tar -xzvf ../baseos.tar;"
	cmds += "tar -xzvf ../ca.tar;"
	cmds += "tar -xzvf ../ccenv.tar;"
	cmds += "tar -xzvf ../couchdb.tar;"
	cmds += "tar -xzvf ../kafka.tar;"
	cmds += "tar -xzvf ../orderer.tar;"
	cmds += "tar -xzvf ../peer.tar;"
	cmds += "tar -xzvf ../zookeeper.tar;"
	cmds += "cat ../tools.tar.* | tar -zxv;"
	cmds += "cat ../javaenv.tar.* | tar -zxv;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

//load docker images
func loadDockerImages(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/images/images_untar;"
	var files = []string{"baseos", "ca", "ccenv", "couchdb", "javaenv", "kafka", "orderer", "peer", "tools", "zookeeper"}
	for i := 0; i < len(files); i++ {
		cmds += "sudo docker load < " + files[i] + ";"
	}
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

//generate crypto
func genCrypto(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/config_file;mkdir -p crypto-config;mkdir -p channel-artifacts;"
	cmds += "cd config;export FABRIC_CFG_PATH=$PWD;cd ..;"
	cmds += "./bin/cryptogen generate --config=./config/crypto-config.yaml;"
	cmds += "mv crypto-config/ config/;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

func genGenesisBlock(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/config_file;"
	cmds += "cd config;export FABRIC_CFG_PATH=$PWD;cd ..;"
	cmds += "./bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock ./channel-artifacts/genesis.block;"
	cmds += "./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/mychannel.tx -channelID mychannel;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

func startFabricNetwork(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/config_file;"
	cmds += "sudo docker-compose -f docker-compose-cli.yaml up -d;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

func makeChannel(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/config_file;"
	cmds += "sudo docker exec -i cli bash;"
	cmds += "peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/mychannel.tx;"
	cmds += "peer channel join -b mychannel.block;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

func installChaincode(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/config_file;"
	//install chaincode
	cmds += "sudo docker exec -i cli bash;"
	cmds += "peer chaincode install -n mycc -p github.com/chaincode/go/ -v 1.0;"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

func instantiateChaincode(hostname, user, passwd string) {
	cmds := "cd /home/" + user + "/fabric/config_file;"
	// instantiate chaincode
	cmds += "sudo docker exec -i cli bash;"
	//cmds += "peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mycc -v 1.0 -c"
	cmds += "peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n mycc -v 1.0 -c '{\"Args\":[\"init\",\"a\",\"100\",\"b\",\"200\"]}' -P \"AND ('Org1MSP.peer')\";"
	executeBatchSshCmd(cmds, hostname, user, passwd)
}

func sshByKey(hostname, username, keysPath, cmd string) (stdout, stderr string, err error) {
	keys = []string{keysPath + "/id_rsa", keysPath + "/id_dsa", keysPath + "/id_ecdsa"}
	signers = []ssh.Signer{}

	for _, keyname := range keys {
		fmt.Printf("key name is %v\n", keyname)
		signer, err := makeSigner(keyname)
		if err == nil {
			signers = append(signers, signer)
		}
	}
	sshAuthSock = os.Getenv("SSH_AUTH_SOCK")
	fmt.Printf("sshauthsock is %v\n", sshAuthSock)
	conn, err := getConnectionByKey(hostname, username)
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

