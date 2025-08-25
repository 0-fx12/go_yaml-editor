package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"vnf-config/internal/infra/db"
	"vnf-config/internal/model"
)

type UploadService struct {
	yamlParser    *YAMLParserService
	dualStorage   *DualStorageService
}

func NewUploadService() *UploadService { 
	return &UploadService{
		yamlParser:  NewYAMLParserService(),
		dualStorage: NewDualStorageService(),
	}
}

// UploadResult 上传结果
type UploadResult struct {
	VNFInstance  *model.VNFInstance
	Definitions  []model.VNFDefinition
	FormFields   map[string]interface{}
	YAMLConfig   *YAMLConfig
	StorageResult *StorageResult
	Errors       []string
}

func (s *UploadService) HandleZipUpload(c *gin.Context, fileHeader *multipart.FileHeader) (*UploadResult, error) {
	result := &UploadResult{}

	// 保存ZIP文件
	uploadDir := defaultString(os.Getenv("UPLOAD_DIR"), "./data/uploads")
	os.MkdirAll(uploadDir, 0755)
	filePath := filepath.Join(uploadDir, fileHeader.Filename)
	if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
		return nil, err
	}

	// 解压ZIP文件
	extractDir := defaultString(os.Getenv("EXTRACT_DIR"), "./data/extracts")
	os.MkdirAll(extractDir, 0755)
	dest := filepath.Join(extractDir, strings.TrimSuffix(fileHeader.Filename, ".zip"))
	if err := unzip(filePath, dest); err != nil {
		return nil, err
	}

	// 查找YAML文件
	yamlPath, err := s.findYAMLFile(dest)
	if err != nil {
		return nil, err
	}

	// 解析YAML文件
	yamlConfig, err := s.yamlParser.ParseYAMLFile(yamlPath)
	if err != nil {
		return nil, err
	}
	result.YAMLConfig = yamlConfig

	// 验证表单项
	validationErrors := s.yamlParser.ValidateFormFields(yamlConfig)
	if len(validationErrors) > 0 {
		result.Errors = validationErrors
	}

	// 提取表单项
	result.FormFields = yamlConfig.Fields

	// 创建VNF实例
	instance := &model.VNFInstance{Name: strings.TrimSuffix(fileHeader.Filename, ".zip")}
	
	// 使用双数据库存储
	storageResult := s.dualStorage.StoreVNFInstance(instance, yamlConfig)
	result.StorageResult = storageResult

	if !storageResult.MySQLSuccess {
		return nil, fmt.Errorf("MySQL存储失败: %v", storageResult.MySQLError)
	}

	// 将VNF实例转换为model.VNFInstance类型
	if vnfInstance, ok := storageResult.Data.(*model.VNFInstance); ok {
		result.VNFInstance = vnfInstance
	} else {
		return nil, errors.New("VNF实例创建失败")
	}

	// 生成VNF定义
	definitions := s.generateVNFDefinitions(yamlConfig, result.VNFInstance.ID)
	result.Definitions = definitions

	// 存储VNF定义到双数据库
	defStorageResult := s.dualStorage.StoreVNFDefinitions(definitions)
	if !defStorageResult.MySQLSuccess {
		result.Errors = append(result.Errors, fmt.Sprintf("VNF定义MySQL存储失败: %v", defStorageResult.MySQLError))
	}
	if !defStorageResult.MongoSuccess {
		result.Errors = append(result.Errors, fmt.Sprintf("VNF定义MongoDB存储失败: %v", defStorageResult.MongoError))
	}

	// 清理临时文件
	defer func() {
		os.Remove(filePath)
		os.RemoveAll(dest)
	}()

	return result, nil
}

// findYAMLFile 查找YAML文件
func (s *UploadService) findYAMLFile(dest string) (string, error) {
	var yamlPath string
	var firstYaml string
	
	err := filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() { 
			return nil 
		}
		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			if firstYaml == "" { 
				firstYaml = path 
			}
			// 优先选择包含特定关键词的YAML文件
			if strings.Contains(name, "config") || strings.Contains(name, "definition") || 
			   strings.Contains(name, "template") || strings.Contains(name, "schema") {
				yamlPath = path
				return filepath.SkipAll // 找到目标文件后停止搜索
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if yamlPath == "" { 
		yamlPath = firstYaml 
	}
	if yamlPath == "" {
		return "", errors.New("在压缩包中未找到YAML文件")
	}

	return yamlPath, nil
}

// generateVNFDefinitions 从YAML配置生成VNF定义
func (s *UploadService) generateVNFDefinitions(yamlConfig *YAMLConfig, vnfID uint) []model.VNFDefinition {
	var definitions []model.VNFDefinition

	for name, field := range yamlConfig.Fields {
		// 转换默认值
		defaultValue := ""
		if field.DefaultValue != nil {
			switch v := field.DefaultValue.(type) {
			case string:
				defaultValue = v
			case int, float64, bool:
				defaultValue = fmt.Sprintf("%v", v)
			default:
				defaultValue = fmt.Sprintf("%v", v)
			}
		}

		// 转换约束
		constraints := ""
		if len(field.Validation) > 0 {
			if constraintsBytes, err := yaml.Marshal(field.Validation); err == nil {
				constraints = string(constraintsBytes)
			}
		}

		// 转换可选性
		var optional *bool
		if field.Required {
			required := false
			optional = &required
		} else {
			optional = &field.Required
		}

		definition := model.VNFDefinition{
			VNFID:           vnfID,
			ParameterName:   name,
			DefaultValue:    defaultValue,
			DescriptionText: field.Description,
			Type:            field.Type,
			CanBeUpdated:    true, // 默认可更新
			HiddenCondition: field.HiddenCondition,
			Optional:        optional,
			Constraints:     constraints,
			CurrentValue:    defaultValue,
			Modified:        false,
		}

		definitions = append(definitions, definition)
	}

	return definitions
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil { 
		return err 
	}
	defer r.Close()
	
	os.MkdirAll(dest, 0755)
	for _, f := range r.File {
		fp := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fp, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", fp)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fp, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(fp), 0755)
		in, err := f.Open(); if err != nil { return err }
		out, err := os.Create(fp); if err != nil { in.Close(); return err }
		if _, err := io.Copy(out, in); err != nil { in.Close(); out.Close(); return err }
		in.Close(); out.Close()
	}
	return nil
}

// 保留原有的yamlParam结构以兼容旧代码
type yamlParam struct {
	Default         string `yaml:"default"`
	Description     string `yaml:"description"`
	Type            string `yaml:"type"`
	CanBeUpdate     *bool  `yaml:"can_be_update"`
	HiddenCondition string `yaml:"hiden_condition"`
	Optional        *bool  `yaml:"optional"`
	Constraints     string `yaml:"constraints"`
}

// parseDefinitionsFromYAML 保留原有的解析方法以兼容旧代码
func (s *UploadService) parseDefinitionsFromYAML(path string) ([]model.VNFDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil { 
		return nil, err 
	}
	var raw map[string]yamlParam
	if err := yaml.Unmarshal(data, &raw); err != nil { 
		return nil, err 
	}
	defs := make([]model.VNFDefinition, 0, len(raw))
	for name, p := range raw {
		def := model.VNFDefinition{
			ParameterName:   name,
			DefaultValue:    p.Default,
			DescriptionText: p.Description,
			Type:            p.Type,
			Constraints:     p.Constraints,
			HiddenCondition: p.HiddenCondition,
		}
		if p.CanBeUpdate != nil { 
			def.CanBeUpdated = *p.CanBeUpdate 
		}
		if p.Optional != nil { 
			def.Optional = p.Optional 
		}
		defs = append(defs, def)
	}
	return defs, nil
}

// to satisfy import warning for net/http (used by gin)
var _ = http.ErrAbortHandler

func defaultString(v, d string) string { 
	if v == "" { 
		return d 
	}; 
	return v 
}


