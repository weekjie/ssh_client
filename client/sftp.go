package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/weekjie/ssh_client/sftp"

	"github.com/weekjie/ssh_client/base"
)

func printHelpMsg() {
	fmt.Println("本执行程序为golang编写的仿sftp客户端, 用于上传或下载文件, 默认工作模式为下载模式.")
	fmt.Println("帮助说明中, 有 * 号是必须设置的参数")
	flag.Usage()
}

func main() {
	host := flag.String("h", "localhost", "SSH服务器主机名或IP")
	port := flag.Int("p", 22, "SSH服务器端口")
	idFile := flag.String("i", "", "使用密钥文件认证, 密码文件的路径")
	password := flag.String("r", "", "使用密码认证, 密码串")
	user := flag.String("u", "", "* 登录账户")
	isHelp := flag.Bool("H", false, "获取帮助信息")
	putType := flag.Bool("P", false, "设置工作模块式为上传模式, 默认为下载模式")
	workHome := flag.String("D", ".", "设置工作目录, 默认当前目录")
	sourcePath := flag.String("source", "", "* 设置源文件路径, 可以是文件/ 目录/ 路径通配符")
	destPath := flag.String("dest", "", "* 设置目的文件路径, 只能是目录, 如不存在会自动创建")
	flag.Parse()
	if *isHelp {
		printHelpMsg()
		os.Exit(0)
	}
	if *user == "" {
		fmt.Println("登录账户必须设置")
		printHelpMsg()
		os.Exit(1)
	}
	if *idFile == "" && *password == "" {
		fmt.Println("必须使用一种认证方式")
		printHelpMsg()
		os.Exit(1)
	}
	if *sourcePath == "" || *destPath == "" {
		fmt.Println("源路劲和目标路径必须都设置")
		printHelpMsg()
		os.Exit(1)
	}
	sshClient, err := base.GetSSHClient(*user, *password, *host, *idFile, *port)
	if err != nil {
		log.Fatal(err)
	}
	sftpClient, err := sftp.GetSFTPClient(sshClient)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Chdir(*workHome)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*sourcePath, *destPath)
	if *putType {
		err = sftpClient.MPut(*sourcePath, *destPath)
		if err != nil {
			log.Fatal(err)
		}
		sftpClient.PutFromRecord()
	} else {
		err = sftpClient.MGet(*sourcePath, *destPath)
		if err != nil {
			log.Fatal(err)
		}
		sftpClient.GetFromRecord()
	}
}
