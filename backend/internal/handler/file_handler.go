package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// FileInfo 文件信息
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// createSSHClient 创建SSH客户端（复用host_handler中的模式）
func createSSHClient(hostID string) (*ssh.Client, error) {
	host, err := hostService.GetHostByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	if host.AuthType == "password" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(host.Password))
	} else if host.AuthType == "key" && host.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(host.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("私钥解析失败: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	sshAddr := fmt.Sprintf("%s:%d", host.IP, host.Port)
	client, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	return client, nil
}

// createSFTPClient 创建SFTP客户端
func createSFTPClient(hostID string) (*sftp.Client, error) {
	sshClient, err := createSSHClient(hostID)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("SFTP客户端创建失败: %w", err)
	}

	return sftpClient, nil
}

// ListFiles 列出远程目录文件
func ListFiles(c *gin.Context) {
	hostID := c.Param("id")
	if hostID == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少主机ID")
		return
	}

	// 获取路径参数，默认为用户主目录
	remotePath := c.Query("path")
	if remotePath == "" {
		remotePath = "."
	}

	sftpClient, err := createSFTPClient(hostID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer sftpClient.Close()

	// 获取远程工作目录（如果路径为相对路径）
	if remotePath == "." {
		wd, err := sftpClient.Getwd()
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "获取工作目录失败: "+err.Error())
			return
		}
		remotePath = wd
	}

	// 列出目录内容
	entries, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "读取目录失败: "+err.Error())
		return
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		fullPath := filepath.Join(remotePath, entry.Name())
		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    fullPath,
			Size:    entry.Size(),
			Mode:    entry.Mode().String(),
			ModTime: entry.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   entry.IsDir(),
		})
	}

	SuccessResponse(c, gin.H{
		"path":  remotePath,
		"files": files,
	})
}

// UploadFile 上传文件到远程主机
func UploadFile(c *gin.Context) {
	hostID := c.Param("id")
	if hostID == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少主机ID")
		return
	}

	remotePath := c.PostForm("path")
	if remotePath == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少目标路径")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "获取上传文件失败")
		return
	}

	srcFile, err := file.Open()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "打开上传文件失败")
		return
	}
	defer srcFile.Close()

	sftpClient, err := createSFTPClient(hostID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer sftpClient.Close()

	// 创建远程文件
	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "创建远程文件失败: "+err.Error())
		return
	}
	defer dstFile.Close()

	// 写入内容
	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "写入文件失败: "+err.Error())
		return
	}

	logger.Info(fmt.Sprintf("文件上传成功: 主机=%s, 路径=%s, 大小=%d", hostID, remotePath, written))

	SuccessResponse(c, gin.H{
		"message": "文件上传成功",
		"path":    remotePath,
		"size":    written,
	})
}

// DownloadFile 从远程主机下载文件
func DownloadFile(c *gin.Context) {
	hostID := c.Param("id")
	if hostID == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少主机ID")
		return
	}

	remotePath := c.Query("path")
	if remotePath == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少文件路径")
		return
	}

	sftpClient, err := createSFTPClient(hostID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer sftpClient.Close()

	// 检查文件是否存在
	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "文件不存在: "+err.Error())
		return
	}

	if stat.IsDir() {
		ErrorResponse(c, http.StatusBadRequest, "不能下载目录")
		return
	}

	// 打开远程文件
	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "打开远程文件失败: "+err.Error())
		return
	}
	defer srcFile.Close()

	// 设置响应头
	fileName := filepath.Base(remotePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", stat.Size()))

	// 流式传输文件内容
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, srcFile)
	if err != nil {
		logger.Error(fmt.Sprintf("文件下载失败: %s, 错误: %v", remotePath, err))
	}
}

// GetFileInfo 获取远程文件信息
func GetFileInfo(c *gin.Context) {
	hostID := c.Param("id")
	if hostID == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少主机ID")
		return
	}

	remotePath := c.Query("path")
	if remotePath == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少文件路径")
		return
	}

	sftpClient, err := createSFTPClient(hostID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer sftpClient.Close()

	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "文件不存在: "+err.Error())
		return
	}

	SuccessResponse(c, FileInfo{
		Name:    stat.Name(),
		Path:    remotePath,
		Size:    stat.Size(),
		Mode:    stat.Mode().String(),
		ModTime: stat.ModTime().Format("2006-01-02 15:04:05"),
		IsDir:   stat.IsDir(),
	})
}
