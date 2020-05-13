package sftp

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/pkg/sftp"
	"week.per/ssh_client/base"
)

// Client 对sftp.Client的封装
type Client struct {
	*sftp.Client
}

// Error 内部错误封装
type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}

// GetSFTPClient 获取sftp客户端, 用于之后的使用
func GetSFTPClient(sshClient *base.SSHClient) (*Client, error) {
	client, err := sftp.NewClient(sshClient.Client)
	if err != nil {
		return nil, err
	}
	log.Printf("SFTP会话建立成功: %s\n", sshClient.Addr)
	return &Client{client}, nil
}

// Quit 关闭sftp会话
func (c *Client) Quit() {
	c.Close()
}

// Get 通过sftp获取单个文件
func (c *Client) Get(fileName string) error {
	baseName := path.Base(fileName)
	fr, err := c.Open(fileName)
	if err != nil {
		return err
	}
	log.Println(fr)
	defer fr.Close()
	fw, err := os.OpenFile(baseName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fw.Close()
	log.Println(fw)
	_, err = fr.WriteTo(fw)
	if err != nil {
		return err
	}
	return nil
}

// MGet 通过sftp获取多个匹配给出匹配符的多个文件
func (c *Client) MGet(fileNamePatten string) error {
	dirName := path.Dir(fileNamePatten)
	var lister base.Lister
	lister = c
	files, err := base.ListPattern(&lister, fileNamePatten)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		msg := fmt.Sprintf("没有知道匹配%s的文件", fileNamePatten)
		return Error{msg}
	}
	var wg sync.WaitGroup
	log.Println("开始下载文件")
	for _, f := range files {
		if f.IsDir() {
			log.Printf("%s是目录, 无法下载\n", f.Name())
			continue
		}
		wg.Add(1)
		go func(filenName string) {
			c.Get(filenName)
			wg.Done()
		}(dirName + "/" + f.Name())
	}
	wg.Wait()
	return nil
}

// Put 通过sftp上传单个文件
func (c *Client) Put(fileName string) error {
	return nil
}

// MPut 通过sftp上传多个匹配给出匹配符的多个文件
func (c *Client) MPut(fileNameReg string) error {
	return nil
}
