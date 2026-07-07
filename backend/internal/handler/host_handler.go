package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aiops/AiOpsHub/backend/internal/service"
	"github.com/gin-gonic/gin"
)

var hostService *service.HostService

func InitHostHandler() {
	hostService = service.NewHostService()
}

func GetGroupTree(c *gin.Context) {
	groups, err := hostService.GetGroupTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "获取分组树失败: " + err.Error(),
		})
		return
	}

	SuccessResponse(c, gin.H{
		"groups": groups,
	})
}

func GetGroupByID(c *gin.Context) {
	id := c.Param("id")

	group, err := hostService.GetGroupByID(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "分组不存在")
		return
	}

	SuccessResponse(c, group)
}

func CreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		ParentID    string `json:"parent_id"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	userID := getUserID(c)

	group, err := hostService.CreateGroup(req.Name, req.ParentID, req.Description, userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message": "分组创建成功",
		"group":   group,
	})
}

func UpdateGroup(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	userID := getUserID(c)

	group, err := hostService.UpdateGroup(id, req.Name, req.Description, userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message": "分组更新成功",
		"group":   group,
	})
}

func DeleteGroup(c *gin.Context) {
	id := c.Param("id")

	err := hostService.DeleteGroup(id)
	if err != nil {
		if strings.Contains(err.Error(), "存在子分组或主机") {
			ErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			ErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	SuccessResponse(c, gin.H{
		"message": "分组删除成功",
	})
}

func CheckGroupCascade(c *gin.Context) {
	id := c.Param("id")

	hasChildren, err := hostService.HasChildrenOrHosts(id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "检查级联失败")
		return
	}

	SuccessResponse(c, gin.H{
		"has_children": hasChildren,
	})
}

func ListHosts(c *gin.Context) {
	groupID := c.Query("group_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	hosts, total, err := hostService.ListHosts(groupID, page, pageSize)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "获取主机列表失败")
		return
	}

	SuccessResponse(c, gin.H{
		"hosts": hosts,
		"total": total,
	})
}

func GetHostByID(c *gin.Context) {
	id := c.Param("id")

	host, err := hostService.GetHostByID(id)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, "主机不存在")
		return
	}

	SuccessResponse(c, host)
}

func CreateHost(c *gin.Context) {
	var req struct {
		GroupID    string `json:"group_id" binding:"required"`
		Name       string `json:"name" binding:"required"`
		HostType   string `json:"host_type"`
		IP         string `json:"ip" binding:"required"`
		Port       int    `json:"port"`
		Username   string `json:"username" binding:"required"`
		AuthType   string `json:"auth_type"`
		Password   string `json:"password"`
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
		Remark     string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if req.HostType == "" {
		req.HostType = "linux"
	}
	if req.Port == 0 {
		req.Port = 22
	}
	if req.AuthType == "" {
		req.AuthType = "password"
	}

	userID := getUserID(c)

	host, err := hostService.CreateHost(
		req.GroupID,
		req.Name,
		req.HostType,
		req.IP,
		req.Port,
		req.Username,
		req.AuthType,
		req.Password,
		req.PrivateKey,
		req.PublicKey,
		req.Remark,
		userID,
	)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message": "主机创建成功",
		"host":    host,
	})
}

func UpdateHost(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	userID := getUserID(c)
	req["updated_by"] = userID

	host, err := hostService.UpdateHost(id, req)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message": "主机更新成功",
		"host":    host,
	})
}

func DeleteHost(c *gin.Context) {
	id := c.Param("id")

	err := hostService.DeleteHost(id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message": "主机删除成功",
	})
}

func BatchImportHosts(c *gin.Context) {
	groupID := c.PostForm("group_id")
	if groupID == "" {
		ErrorResponse(c, http.StatusBadRequest, "缺少分组ID")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "上传文件失败")
		return
	}

	fileReader, err := file.Open()
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "打开文件失败")
		return
	}
	defer fileReader.Close()

	userID := getUserID(c)

	hosts, importErrors, err := hostService.BatchImportHosts(groupID, fileReader, userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message":      "批量导入完成",
		"import_count": len(hosts),
		"error_count":  len(importErrors),
		"errors":       importErrors,
		"hosts":        hosts,
	})
}

func BatchDeleteHosts(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if len(req.IDs) == 0 {
		ErrorResponse(c, http.StatusBadRequest, "主机ID列表为空")
		return
	}

	err := hostService.BatchDeleteHosts(req.IDs)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message":      "批量删除成功",
		"delete_count": len(req.IDs),
	})
}

func TestHostConnection(c *gin.Context) {
	id := c.Param("id")

	err := hostService.TestConnection(id)
	if err != nil {
		SuccessResponse(c, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	SuccessResponse(c, gin.H{
		"success": true,
		"message": "连接测试成功",
	})
}

func getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return "system"
	}

	if id, ok := userID.(string); ok {
		return id
	}

	return "system"
}
