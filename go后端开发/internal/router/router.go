package router

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"vnf-config/internal/controller/v1"
)

func New() *gin.Engine {
	if os.Getenv("APP_ENV") != "production" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	staticDir := defaultString(os.Getenv("STATIC_DIR"), "./web")
	r.Static("/", staticDir)

	api := r.Group("/api/v1")
	{
		uploadCtl := v1.NewUploadController()
		vnfCtl := v1.NewVNFController()
		defCtl := v1.NewDefinitionController()

		// 上传相关
		api.POST("/uploads", uploadCtl.UploadZip)
		api.GET("/vnfs/:id/form-fields", uploadCtl.GetFormFields)
		api.GET("/vnfs/:id/yaml-config", uploadCtl.GetYAMLConfig)

		// VNF实例管理
		api.GET("/vnfs", vnfCtl.ListVNFInstances)
		api.GET("/vnfs/:id", vnfCtl.GetVNFInstance)
		api.DELETE("/vnfs/:id", vnfCtl.DeleteVNFInstance)

		// VNF定义管理
		api.GET("/vnfs/:id/definitions", defCtl.ListDefinitions)
		api.POST("/vnfs/:id/definitions", defCtl.CreateDefinition)
		api.PUT("/vnfs/:id/definitions/:defId", defCtl.UpdateDefinition)
		api.DELETE("/vnfs/:id/definitions/:defId", defCtl.DeleteDefinition)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "路由未找到"})
	})

	return r
}

func defaultString(v string, d string) string {
	if v == "" {
		return d
	}
	return v
}


