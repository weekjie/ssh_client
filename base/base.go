package base

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient 对ssh.client的封装
type SSHClient struct {
	*ssh.Client
	Addr string
}

// SSHSession 对ssh.Session的封装
type SSHSession struct {
	*ssh.Session
}

// GetSSHClient 用于获取SSHClinet
func GetSSHClient(user, password, host, rsaFile string, port int) (*SSHClient, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		err          error
	)

	// 设置登录方式
	auth = make([]ssh.AuthMethod, 0)
	if password != "" {
		auth = append(auth, ssh.Password(password))
	} else {
		rsaBytes, err := ioutil.ReadFile(rsaFile)
		if err != nil {
			return nil, err
		}
		var signer ssh.Signer
		signer, err = ssh.ParsePrivateKey(rsaBytes)
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	// 设置加密函数
	config = ssh.Config{
		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
	}

	// 设置客户端配置
	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 10 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// 获取地址
	addr = fmt.Sprintf("%s:%d", host, port)

	// 建立连接
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	log.Printf("连接建立成功: %s\n", addr)
	return &SSHClient{client, addr}, nil
}

// GetSession 从SSHClient中获取session
func (client *SSHClient) GetSession() (*SSHSession, error) {
	// 建立Session
	var (
		session *ssh.Session
		err     error
	)
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	// 设置Session终端, 直接拷贝官网代码
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err = session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}
	log.Printf("会话建立成功: %s", client.Addr)

	return &SSHSession{session}, nil
}

// SSHSessionError 内部的error封装
type SSHSessionError struct {
	msg string
}

// 内部Error实现
func (e SSHSessionError) Error() string {
	return e.msg
}

// ExceuteShell 在session中执行shell命令
func (session *SSHSession) ExceuteShell(cmd []string) error {
	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Shell()
	if err != nil {
		return err
	}
	for _, c := range cmd {
		c = c + "\n"
		// log.Print(c)
		stdin.Write([]byte(c))
	}
	stdin.Write([]byte("exit\n"))
	session.Wait()
	log.Print(stdout.String())
	if stderr.String() != "" {
		// session.Quit()
		return SSHSessionError{"Shell execute error: " + stderr.String()}
	}

	return nil
}

// Quit 关闭Client
func (client *SSHClient) Quit() {
	client.Close()
}

// Quit 关闭session
func (session *SSHSession) Quit() {
	session.Close()
}

// LocalFileSystem 本地终端, 用于在本地文件系统中执行操作
type LocalFileSystem struct{}

// Lstat 对默认Lstat方法的封装
func (l *LocalFileSystem) Lstat(p string) (os.FileInfo, error) {
	return os.Lstat(p)
}

// ReadDir 对默认ReadDir方法的封装
func (l *LocalFileSystem) ReadDir(p string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(p)
}

// Lister 封装列出目录的接口
type Lister interface {
	Lstat(p string) (os.FileInfo, error)
	ReadDir(p string) ([]os.FileInfo, error)
}

// List 通过Lister接口, 列出指定路径p中的文件列表
func List(l *Lister, p string) ([]os.FileInfo, error) {
	// log.Printf("%#v\n", *l)
	dirInfo, err := (*l).Lstat(p)
	if err != nil {
		return nil, err
	}
	if !dirInfo.IsDir() {
		msg := fmt.Sprintf("%s is not a directory", p)
		return nil, errors.New(msg)
	}
	// log.Printf("%v\n", dirInfo)
	return (*l).ReadDir(p)
}

// IsMatched 测试文件名和通配符是否匹配
func IsMatched(p, pattern string) bool {
	matched, err := filepath.Match(pattern, p)
	if err != nil {
		return false
	}
	return matched
}

// ListPattern 列出匹配文件
func ListPattern(l *Lister, pp string) ([]os.FileInfo, error) {
	dirName := path.Dir(pp)
	filePatten := path.Base(pp)
	if strings.Contains(dirName, "*") {
		return nil, errors.New("目录名中不可含有'*'")
	}
	files, err := List(l, dirName)
	if err != nil {
		return nil, err
	}
	resFiles := make([]os.FileInfo, 0, len(files))
	for _, f := range files {
		if IsMatched(f.Name(), filePatten) {
			resFiles = append(resFiles, f)
		}
	}
	return resFiles, nil
}
