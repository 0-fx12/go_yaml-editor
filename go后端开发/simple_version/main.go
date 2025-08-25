package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
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

// parseYAMLFile 解析YAML文件（保留文件中原始键顺序）
func parseYAMLFile(filePath string) (*YAMLData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 原始内容（用于 /yaml/raw）
	var rawData interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, err
	}

	// 使用 yaml.Node 保留顺序
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	var fields []Field
	if len(root.Content) > 0 {
		fields = extractFieldsNode("", root.Content[0])
	}

	return &YAMLData{
		Content: rawData,
		Fields:  fields,
	}, nil
}

// extractFieldsNode 使用 yaml.Node 递归提取字段，按文件顺序遍历
func extractFieldsNode(path string, node *yaml.Node) []Field {
	var fields []Field

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			newPath := buildPath(path, keyNode.Value)
			fields = append(fields, extractFieldsNode(newPath, valNode)...)
		}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			newPath := buildPath(path, "[]")
			fields = append(fields, extractFieldsNode(newPath, item)...)
		}
	case yaml.ScalarNode:
		if path != "" {
			val := valueFromNode(node)
			fields = append(fields, Field{Path: path, Value: val, Type: getType(val), Description: ""})
		}
	}

	return fields
}

// valueFromNode 将标量节点解码为对应Go类型
func valueFromNode(n *yaml.Node) interface{} {
	var v interface{}
	if err := n.Decode(&v); err == nil {
		return v
	}
	return n.Value
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

// findWritableYAMLFile 返回可写的yaml文件路径（存在的第一个）
func findWritableYAMLFile() (string, error) {
	candidates := []string{"config.yaml", "sample_config.yaml", "test.yaml"}
	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}
	return "", errors.New("未找到可写的YAML文件")
}

// setNodeValueByPath 在 yaml.Node 中按点路径设置标量值（不支持数组路径）
func setNodeValueByPath(root *yaml.Node, path string, value interface{}) error {
	if strings.Contains(path, "[]") {
		return errors.New("不支持修改数组路径: " + path)
	}
	parts := strings.Split(path, ".")
	cur := root
	// 根应是 MappingNode
	if cur.Kind == yaml.DocumentNode && len(cur.Content) > 0 {
		cur = cur.Content[0]
	}
	for i := 0; i < len(parts); i++ {
		if cur.Kind != yaml.MappingNode {
			return errors.New("路径非对象节点: " + parts[i])
		}
		key := parts[i]
		found := false
		for j := 0; j+1 < len(cur.Content); j += 2 {
			k := cur.Content[j]
			v := cur.Content[j+1]
			if k.Value == key {
				if i == len(parts)-1 {
					// 设置标量值
					setScalar(v, value)
					return nil
				}
				// 继续深入
				if v.Kind != yaml.MappingNode {
					// 若不是对象，则替换为对象节点
					v.Kind = yaml.MappingNode
					v.Tag = "!!map"
					v.Content = []*yaml.Node{}
				}
				cur = v
				found = true
				break
			}
		}
		if !found {
			// 创建缺失的 key 与空对象/标量
			k := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
			var v *yaml.Node
			if i == len(parts)-1 {
				v = &yaml.Node{}
				setScalar(v, value)
			} else {
				v = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			}
			cur.Content = append(cur.Content, k, v)
			cur = v
		}
	}
	return nil
}

// setScalar 将 node 设置为对应类型的标量
func setScalar(n *yaml.Node, v interface{}) {
	switch x := v.(type) {
	case bool:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!bool"
		if x {
			n.Value = "true"
		} else {
			n.Value = "false"
		}
	case int, int8, int16, int32, int64:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!int"
		n.Value = strconv.FormatInt(toInt64(x), 10)
	case uint, uint8, uint16, uint32, uint64:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!int"
		n.Value = strconv.FormatUint(toUint64(x), 10)
	case float32, float64:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!float"
		n.Value = strconv.FormatFloat(toFloat64(x), 'f', -1, 64)
	default:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!str"
		n.Value = toString(v)
	}
}

func toInt64(v interface{}) int64 { switch t := v.(type) { case int: return int64(t); case int8: return int64(t); case int16: return int64(t); case int32: return int64(t); case int64: return t; default: return 0 } }
func toUint64(v interface{}) uint64 { switch t := v.(type) { case uint: return uint64(t); case uint8: return uint64(t); case uint16: return uint64(t); case uint32: return uint64(t); case uint64: return t; default: return 0 } }
func toFloat64(v interface{}) float64 { switch t := v.(type) { case float32: return float64(t); case float64: return t; default: return 0 } }
func toString(v interface{}) string { if v == nil { return "" }; s, ok := v.(string); if ok { return s }; return stringify(v) }
func stringify(v interface{}) string { b, _ := yaml.Marshal(v); return strings.TrimSpace(string(b)) }

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

		// 保存YAML修改（基于 yaml.Node 原位更新，保留原始顺序）
		api.POST("/yaml", func(c *gin.Context) {
			var req struct { Updates map[string]interface{} `json:"updates"` }
			if err := c.ShouldBindJSON(&req); err != nil || req.Updates == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
				return
			}

			filePath, err := findWritableYAMLFile()
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			// 读取为 yaml.Node
			b, err := os.ReadFile(filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
				return
			}
			var root yaml.Node
			if err := yaml.Unmarshal(b, &root); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "解析YAML失败"})
				return
			}

			// 应用更新到节点
			for p, v := range req.Updates {
				_ = setNodeValueByPath(&root, p, v)
			}

			// 备份原文件
			_ = os.WriteFile(filePath+".bak", b, 0644)

			// 写回，保留原始键顺序
			out, err := yaml.Marshal(&root)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "生成YAML失败"})
				return
			}
			if err := os.WriteFile(filePath, out, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "saved", "file": filepath.Base(filePath)})
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
	srv := &http.Server{ Addr: ":" + port, Handler: r }

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
