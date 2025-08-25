package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// YAMLData 存储解析后的YAML数据
type YAMLData struct {
	Content interface{} `json:"content"`
	Fields  []Field     `json:"fields"`
}

// Field 字段结构
type Field struct {
	Path        string      `json:"path"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
}

// parseYAMLFile 解析YAML文件
func parseYAMLFile(filePath string) (*YAMLData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rawData interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	// 提取所有字段
	fields := extractFields("", rawData)

	return &YAMLData{
		Content: rawData,
		Fields:  fields,
	}, nil
}

// extractFields 递归提取字段
func extractFields(path string, node interface{}) []Field {
	var fields []Field

	switch v := node.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPath := buildPath(path, key)
			fields = append(fields, extractFields(newPath, value)...)
		}
	case []interface{}:
		for _, item := range v {
			newPath := buildPath(path, "[]")
			fields = append(fields, extractFields(newPath, item)...)
		}
	default:
		// 叶子节点
		if path != "" {
			field := Field{
				Path:        path,
				Value:       v,
				Type:        getType(v),
				Description: "",
			}
			fields = append(fields, field)
		}
	}

	return fields
}

// buildPath 构建路径
func buildPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

// getType 获取类型
func getType(value interface{}) string {
	switch value.(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	case float32, float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "string"
	}
}

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	// API路由（先注册API，避免与静态资源通配符冲突）
	api := r.Group("/api/v1")
	{
		// 获取YAML数据（支持分页）
		api.GET("/yaml", func(c *gin.Context) {
			// 尝试读取当前目录下的yaml文件
			yamlFiles := []string{"config.yaml", "sample_config.yaml", "test.yaml"}
			var yamlData *YAMLData

			for _, filename := range yamlFiles {
				if _, err := os.Stat(filename); err == nil {
					if data, err := parseYAMLFile(filename); err == nil {
						yamlData = data
						break
					}
				}
			}

			if yamlData == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "未找到可用的YAML文件"})
				return
			}

			// 分页参数
			page := 1
			pageSize := 20
			if pageStr := c.Query("page"); pageStr != "" {
				if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
					page = p
				}
			}
			if sizeStr := c.Query("size"); sizeStr != "" {
				if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
					pageSize = s
				}
			}

			// 计算分页
			total := len(yamlData.Fields)
			start := (page - 1) * pageSize
			end := start + pageSize
			if start >= total {
				start = total
			}
			if end > total {
				end = total
			}

			var pageFields []Field
			if start < total {
				pageFields = yamlData.Fields[start:end]
			}

			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"fields":    pageFields,
					"total":     total,
					"page":      page,
					"pageSize":  pageSize,
					"totalPage": (total + pageSize - 1) / pageSize,
				},
				"message": "success",
			})
		})

		// 获取YAML原始内容
		api.GET("/yaml/raw", func(c *gin.Context) {
			yamlFiles := []string{"config.yaml", "sample_config.yaml", "test.yaml"}
			var yamlData *YAMLData

			for _, filename := range yamlFiles {
				if _, err := os.Stat(filename); err == nil {
					if data, err := parseYAMLFile(filename); err == nil {
						yamlData = data
						break
					}
				}
			}

			if yamlData == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "未找到可用的YAML文件"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"data":    yamlData.Content,
				"message": "success",
			})
		})
	}

	// 静态资源（放在最后，避免与 /api 路由冲突）
	r.Static("/static", "./web")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "路由未找到"})
	})

	// 获取端口配置
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在端口 %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("服务器已关闭")
}
