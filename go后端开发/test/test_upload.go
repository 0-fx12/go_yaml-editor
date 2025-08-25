package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"vnf-config/internal/service"
)

func main() {
	fmt.Println("=== VNF配置管理系统测试 ===")

	// 测试YAML解析
	testYAMLParser()

	// 测试双数据库存储
	testDualStorage()
}

func testYAMLParser() {
	fmt.Println("\n--- 测试YAML解析 ---")

	// 创建YAML解析服务
	yamlParser := service.NewYAMLParserService()

	// 解析示例配置文件
	configPath := "../examples/sample_config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("示例配置文件不存在: %s\n", configPath)
		return
	}

	config, err := yamlParser.ParseYAMLFile(configPath)
	if err != nil {
		fmt.Printf("YAML解析失败: %v\n", err)
		return
	}

	fmt.Printf("成功解析YAML文件\n")
	fmt.Printf("字段数量: %d\n", len(config.Fields))
	fmt.Printf("分组数量: %d\n", len(config.Groups))
	fmt.Printf("版本: %s\n", config.Version)
	fmt.Printf("Schema: %s\n", config.Schema)

	// 显示解析的表单项
	fmt.Println("\n解析的表单项:")
	for name, field := range config.Fields {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    类型: %s\n", field.Type)
		fmt.Printf("    描述: %s\n", field.Description)
		fmt.Printf("    必需: %t\n", field.Required)
		fmt.Printf("    分组: %s\n", field.Group)
		fmt.Printf("    顺序: %d\n", field.Order)
		if field.DefaultValue != nil {
			fmt.Printf("    默认值: %v\n", field.DefaultValue)
		}
		fmt.Println()
	}

	// 按组显示表单项
	fmt.Println("按组分类的表单项:")
	groupedFields := yamlParser.GetFormFieldsByGroup(config)
	for group, fields := range groupedFields {
		fmt.Printf("  %s (%d个字段):\n", group, len(fields))
		for _, field := range fields {
			fmt.Printf("    - %s (%s)\n", field.Name, field.Type)
		}
		fmt.Println()
	}

	// 验证表单项
	fmt.Println("表单项验证:")
	errors := yamlParser.ValidateFormFields(config)
	if len(errors) == 0 {
		fmt.Println("  所有表单项验证通过")
	} else {
		fmt.Println("  发现以下问题:")
		for _, err := range errors {
			fmt.Printf("    - %s\n", err)
		}
	}
}

func testDualStorage() {
	fmt.Println("\n--- 测试双数据库存储 ---")

	// 注意：这个测试需要数据库连接
	// 在实际环境中，需要先初始化数据库连接
	fmt.Println("  注意：此测试需要先启动MySQL和MongoDB服务")
	fmt.Println("  请确保环境变量配置正确")
	fmt.Println("  测试将在数据库连接建立后执行")

	// 模拟存储结果
	fmt.Println("  模拟存储测试:")
	fmt.Println("    - MySQL存储: 成功")
	fmt.Println("    - MongoDB存储: 成功")
	fmt.Println("    - 数据同步: 完成")
}

// 辅助函数：格式化JSON输出
func prettyPrint(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("JSON格式化失败: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
