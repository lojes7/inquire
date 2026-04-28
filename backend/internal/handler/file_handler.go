package handler

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lojes7/inquire/internal/model"
	"github.com/lojes7/inquire/internal/service"
	"github.com/lojes7/inquire/pkg/infra"
	"github.com/lojes7/inquire/pkg/response"
	"github.com/lojes7/inquire/pkg/secure"
)

// UploadFile 上传文件到工作区
// @Summary      上传文件到工作区
// @Description  上传文件到个人工作区，不在任何会话中
// @Tags         file
// @Accept       multipart/form-data
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Param        file formData file true "文件"
// @Success      201  {object}  response.Response   "上传成功"
// @Failure      400  {object}  response.Response   "没有接收到文件"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/files/upload [post]
func UploadFile(c *gin.Context) {
	userID := c.GetUint64("id")

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 400, "没有接收到文件")
		return
	}

	resp, err := service.UploadFile(userID, file)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}
	response.Success(c, 201, "success", resp)
}

// WorkspaceFileList 获取工作区文件列表
// @Summary      工作区文件列表
// @Description  获取当前用户工作区中的所有文件
// @Tags         file
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer Token"
// @Success      200  {object}  response.Response   "查询成功"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/files [get]
func WorkspaceFileList(c *gin.Context) {
	userID := c.GetUint64("id")

	files, err := service.WorkspaceFileList(userID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}

	resp := model.WorkspaceFileListResp{Files: files}
	response.Success(c, 200, "success", resp)
}

// DownloadFile 下载文件
// @Summary      下载文件
// @Description  Download a file by message ID
// @Tags         message
// @Accept       json
// @Produce      application/octet-stream
// @Param        Authorization header string true "Bearer Token"
// @Param        message_id path string true "消息ID"
// @Success      200  {file}    file                "文件内容"
// @Failure      400  {object}  response.Response   "message_id错误"
// @Failure      403  {object}  response.Response   "非法文件路径"
// @Failure      404  {object}  response.Response   "文件不存在"
// @Failure      500  {object}  response.Response   "服务器错误"
// @Router       /auth/files/{message_id} [get]
func DownloadFile(c *gin.Context) {
	userID := c.GetUint64("id")
	messageIDStr := c.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		response.Fail(c, 400, "message_id 格式错误")
		return
	}

	fileURL, err := service.DownloadFile(userID, messageID)
	if err != nil {
		if myErr := secure.Unwrap(err); myErr != nil {
			response.Fail(c, myErr.Code, myErr.Message)
		} else {
			response.Fail(c, 500, "服务器错误")
		}
		return
	}

	// 安全检查，确保文件路径在允许的目录下
	allowedDir := infra.GetFilePath()
	if !strings.HasPrefix(fileURL, allowedDir) {
		response.Fail(c, 403, "非法文件路径")
		return
	}

	// 打开文件
	file, err := os.Open(fileURL)
	if err != nil {
		if os.IsNotExist(err) {
			response.Fail(c, 404, "文件不存在")
		} else {
			log.Printf("打开文件失败: %v\n", err)
			response.Fail(c, 500, "文件读取失败")
		}
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		response.Fail(c, 500, "无法获取文件信息")
		return
	}

	// 设置响应头
	fileName := filepath.Base(fileURL) // 例如 "report.pdf"

	// 自动推断 MIME 类型
	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Header("Content-Type", mimeType)
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.PathEscape(fileName)))

	// 流式传输文件（不加载全文件到内存）
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		// 客户端可能取消了下载，通常不用报错
		log.Printf("文件传输中断: %v", err)
		return
	}
}
