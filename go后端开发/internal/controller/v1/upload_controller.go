package v1

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"vnf-config/internal/service"
)

type UploadController struct {
	uploadService *service.UploadService
}

func NewUploadController() *UploadController {
	return &UploadController{uploadService: service.NewUploadService()}
}

func (u *UploadController) UploadZip(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件是必需的"})
		return
	}

	if !strings.EqualFold(filepath.Ext(file.Filename), ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只允许上传.zip文件"})
		return
	}

	result, err := u.uploadService.HandleZipUpload(c, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"message": "YAML文件上传并解析成功",
		"data": gin.H{
			"vnf": gin.H{
				"id":        result.VNFInstance.ID,
				"name":      result.VNFInstance.Name,
				"createdAt": result.VNFInstance.CreatedAt,
			},
			"definitions": result.Definitions,
			"formFields":  result.FormFields,
			"yamlConfig": gin.H{
				"fields":   result.YAMLConfig.Fields,
				"groups":   result.YAMLConfig.Groups,
				"metadata": result.YAMLConfig.Metadata,
				"version":  result.YAMLConfig.Version,
				"schema":   result.YAMLConfig.Schema,
			},
			"storage": gin.H{
				"mysql": gin.H{
					"success": result.StorageResult.MySQLSuccess,
					"error":   result.StorageResult.MySQLError,
				},
				"mongodb": gin.H{
					"success": result.StorageResult.MongoSuccess,
					"error":   result.StorageResult.MongoError,
				},
			},
		},
	}

	// 如果有错误，添加到响应中
	if len(result.Errors) > 0 {
		response["warnings"] = result.Errors
	}

	c.JSON(http.StatusOK, response)
}

// GetFormFields 获取表单项
func (u *UploadController) GetFormFields(c *gin.Context) {
	vnfID := c.Param("id")
	if vnfID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VNF ID是必需的"})
		return
	}

	// 这里可以调用服务获取表单项
	// 暂时返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取表单项成功",
		"data":    gin.H{},
	})
}

// GetYAMLConfig 获取YAML配置
func (u *UploadController) GetYAMLConfig(c *gin.Context) {
	vnfID := c.Param("id")
	if vnfID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VNF ID是必需的"})
		return
	}

	// 这里可以调用服务获取YAML配置
	// 暂时返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取YAML配置成功",
		"data":    gin.H{},
	})
}


