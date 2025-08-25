package service

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLParserService YAML解析服务
type YAMLParserService struct{}

func NewYAMLParserService() *YAMLParserService {
	return &YAMLParserService{}
}

// FormField 表单项结构
type FormField struct {
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	DefaultValue    interface{}            `json:"defaultValue"`
	Description     string                 `json:"description"`
	Required        bool                   `json:"required"`
	Hidden          bool                   `json:"hidden"`
	HiddenCondition string                 `json:"hiddenCondition"`
	Validation      map[string]interface{} `json:"validation"`
	Options         []interface{}          `json:"options,omitempty"`
	Group           string                 `json:"group,omitempty"`
	Order           int                    `json:"order"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// YAMLConfig YAML配置结构
type YAMLConfig struct {
	Fields    map[string]FormField `json:"fields"`
	Groups    map[string]string    `json:"groups"`
	Metadata  map[string]interface{} `json:"metadata"`
	Version   string               `json:"version"`
	Schema    string               `json:"schema"`
}

// ParseYAMLFile 解析YAML文件并提取表单项
func (s *YAMLParserService) ParseYAMLFile(filePath string) (*YAMLConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取YAML文件失败: %v", err)
	}

	var rawData interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("YAML解析失败: %v", err)
	}

	config := &YAMLConfig{
		Fields:   make(map[string]FormField),
		Groups:   make(map[string]string),
		Metadata: make(map[string]interface{}),
	}

	// 递归解析YAML结构
	s.parseNode("", rawData, config, 0)

	return config, nil
}

// parseNode 递归解析YAML节点
func (s *YAMLParserService) parseNode(path string, node interface{}, config *YAMLConfig, order int) {
	switch v := node.(type) {
	case map[string]interface{}:
		// 检查是否是特殊配置节点
		if s.isSpecialConfigNode(path, v) {
			s.parseSpecialConfig(path, v, config)
			return
		}

		// 检查是否包含表单项属性
		if s.hasFormFieldProperties(v) {
			field := s.extractFormField(path, v, order)
			if field.Name != "" {
				config.Fields[field.Name] = field
			}
		} else {
			// 递归解析子节点
			for key, value := range v {
				newPath := s.buildPath(path, key)
				s.parseNode(newPath, value, config, order+1)
			}
		}

	case []interface{}:
		// 处理数组类型
		for i, item := range v {
			newPath := s.buildPath(path, fmt.Sprintf("[%d]", i))
			s.parseNode(newPath, item, config, order+1)
		}

	case string, int, float64, bool:
		// 叶子节点，可能是简单的配置值
		if path != "" {
			field := s.createSimpleField(path, v, order)
			config.Fields[field.Name] = field
		}
	}
}

// isSpecialConfigNode 检查是否是特殊配置节点
func (s *YAMLParserService) isSpecialConfigNode(path string, node map[string]interface{}) bool {
	specialKeys := []string{"metadata", "groups", "schema", "version", "config", "settings"}
	for _, key := range specialKeys {
		if strings.Contains(strings.ToLower(path), key) {
			return true
		}
	}
	return false
}

// parseSpecialConfig 解析特殊配置节点
func (s *YAMLParserService) parseSpecialConfig(path string, node map[string]interface{}, config *YAMLConfig) {
	for key, value := range node {
		switch key {
		case "metadata":
			if metadata, ok := value.(map[string]interface{}); ok {
				config.Metadata = metadata
			}
		case "groups":
			if groups, ok := value.(map[string]interface{}); ok {
				for groupKey, groupValue := range groups {
					if groupName, ok := groupValue.(string); ok {
						config.Groups[groupKey] = groupName
					}
				}
			}
		case "schema", "version":
			if strValue, ok := value.(string); ok {
				if key == "schema" {
					config.Schema = strValue
				} else if key == "version" {
					config.Version = strValue
				}
			}
		}
	}
}

// hasFormFieldProperties 检查是否包含表单项属性
func (s *YAMLParserService) hasFormFieldProperties(node map[string]interface{}) bool {
	formFieldKeys := []string{
		"type", "default", "description", "required", "hidden", "validation",
		"options", "group", "order", "constraints", "can_be_update", "optional",
	}
	
	for _, key := range formFieldKeys {
		if _, exists := node[key]; exists {
			return true
		}
	}
	return false
}

// extractFormField 提取表单项
func (s *YAMLParserService) extractFormField(path string, node map[string]interface{}, order int) FormField {
	field := FormField{
		Name:            path,
		Type:            "string",
		DefaultValue:    nil,
		Description:     "",
		Required:        false,
		Hidden:          false,
		HiddenCondition: "",
		Validation:      make(map[string]interface{}),
		Options:         []interface{}{},
		Group:           "",
		Order:           order,
		Metadata:        make(map[string]interface{}),
	}

	// 提取基本属性
	for key, value := range node {
		switch strings.ToLower(key) {
		case "type":
			if strValue, ok := value.(string); ok {
				field.Type = strValue
			}
		case "default", "default_value":
			field.DefaultValue = value
		case "description", "desc", "help":
			if strValue, ok := value.(string); ok {
				field.Description = strValue
			}
		case "required", "mandatory":
			if boolValue, ok := value.(bool); ok {
				field.Required = boolValue
			}
		case "hidden", "visible":
			if boolValue, ok := value.(bool); ok {
				field.Hidden = boolValue
			}
		case "hidden_condition", "hiden_condition", "visibility":
			if strValue, ok := value.(string); ok {
				field.HiddenCondition = strValue
			}
		case "validation", "constraints", "rules":
			if validationMap, ok := value.(map[string]interface{}); ok {
				field.Validation = validationMap
			}
		case "options", "choices", "enum":
			if optionsSlice, ok := value.([]interface{}); ok {
				field.Options = optionsSlice
			}
		case "group", "category":
			if strValue, ok := value.(string); ok {
				field.Group = strValue
			}
		case "order", "sort":
			if intValue, ok := value.(int); ok {
				field.Order = intValue
			}
		default:
			// 其他属性作为元数据
			field.Metadata[key] = value
		}
	}

	// 智能类型推断
	if field.Type == "string" && field.DefaultValue != nil {
		field.Type = s.inferType(field.DefaultValue)
	}

	return field
}

// createSimpleField 创建简单字段
func (s *YAMLParserService) createSimpleField(path string, value interface{}, order int) FormField {
	return FormField{
		Name:            path,
		Type:            s.inferType(value),
		DefaultValue:    value,
		Description:     "",
		Required:        false,
		Hidden:          false,
		HiddenCondition: "",
		Validation:      make(map[string]interface{}),
		Options:         []interface{}{},
		Group:           "",
		Order:           order,
		Metadata:        make(map[string]interface{}),
	}
}

// inferType 推断字段类型
func (s *YAMLParserService) inferType(value interface{}) string {
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

// buildPath 构建路径
func (s *YAMLParserService) buildPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

// GetFormFieldsByGroup 按组获取表单项
func (s *YAMLParserService) GetFormFieldsByGroup(config *YAMLConfig) map[string][]FormField {
	groupedFields := make(map[string][]FormField)
	
	for _, field := range config.Fields {
		group := field.Group
		if group == "" {
			group = "default"
		}
		groupedFields[group] = append(groupedFields[group], field)
	}
	
	return groupedFields
}

// ValidateFormFields 验证表单项
func (s *YAMLParserService) ValidateFormFields(config *YAMLConfig) []string {
	var errors []string
	
	for name, field := range config.Fields {
		if field.Required && field.DefaultValue == nil {
			errors = append(errors, fmt.Sprintf("字段 '%s' 是必需的但没有默认值", name))
		}
		
		if field.Type == "array" && len(field.Options) == 0 {
			errors = append(errors, fmt.Sprintf("字段 '%s' 是数组类型但没有定义选项", name))
		}
	}
	
	return errors
}
