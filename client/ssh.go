package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/weekjie/ssh_client/base"
)

func printHelpMsg() {
	fmt.Println("本执行程序为golang编写的仿ssh客户端, 用于访问SSH服务, 有终端模式和命令模式")
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
	cmds := flag.String("c", "", "执行命令, 如果要执行多个命令, 需要用';'分隔; 如果未设置或为空, 则进入终端模式")
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
	sshClient, err := base.GetSSHClient(*user, *password, *host, *idFile, *port)
	if err != nil {
		log.Fatal(err)
	}
	defer sshClient.Quit()
	sshSession, err := sshClient.GetSession()
	if err != nil {
		log.Fatal(err)
	}
	defer sshSession.Quit()
	if *cmds == "" {
		err = sshSession.Terminal()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = sshSession.ExceuteCmd(strings.Split(*cmds, ";"))
		if err != nil {
			log.Fatal(err)
		}
	}

}
