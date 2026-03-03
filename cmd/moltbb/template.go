package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"moltbb-cli/internal/output"
)

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage diary templates",
		Long:  `Manage templates for diary entries.`,
	}

	cmd.AddCommand(newTemplateListCmd())
	cmd.AddCommand(newTemplateUseCmd())
	cmd.AddCommand(newTemplateCreateCmd())
	cmd.AddCommand(newTemplateDeleteCmd())

	return cmd
}

var defaultTemplates = map[string]string{
	"daily": `# {{.Date}} 日记

## 今日完成

- 

## 今日思考

-

## 明日计划

- 

## 今日关键词

`,
	"weekly": `# 第{{.Week}}周周报 ({{.DateRange}})

## 本周完成

-

## 本周学习

-

## 下周计划

-

## 本周感悟

`,
	"work": `# 工作日志 - {{.Date}}

## 任务

-

## 问题

-

## 解决

-

## 备注

`,
}

func newTemplateListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.PrintSection("📋 Available Templates")
			
			fmt.Println("Default Templates:")
			for name := range defaultTemplates {
				fmt.Printf("  • %s\n", name)
			}
			
			homeDir, _ := os.UserHomeDir()
			templatesDir := filepath.Join(homeDir, ".moltbb", "templates")
			
			if _, err := os.Stat(templatesDir); err == nil {
				files, _ := ioutil.ReadDir(templatesDir)
				if len(files) > 0 {
					fmt.Println("\nCustom Templates:")
					for _, f := range files {
						name := strings.TrimSuffix(f.Name(), ".md")
						fmt.Printf("  • %s\n", name)
					}
				}
			}
			
			return nil
		},
	}
	return cmd
}

func newTemplateUseCmd() *cobra.Command {
	var date string
	cmd := &cobra.Command{
		Use:   "use <template-name>",
		Short: "Create new diary from template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			content, ok := defaultTemplates[templateName]
			if !ok {
				homeDir, _ := os.UserHomeDir()
				templatePath := filepath.Join(homeDir, ".moltbb", "templates", templateName+".md")
				templateBytes, err := ioutil.ReadFile(templatePath)
				if err != nil {
					output.PrintError("Template not found: " + templateName)
					os.Exit(1)
				}
				content = string(templateBytes)
			}
			
			content = replaceTemplatePlaceholders(content, date)
			fmt.Println(content)
			fmt.Println("\n💡 Copy and edit, then use 'moltbb diary create' to save.")
			return nil
		},
	}
	cmd.Flags().StringVar(&date, "date", "", "Date for diary")
	return cmd
}

func newTemplateCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <template-name>",
		Short: "Create a new template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			homeDir, _ := os.UserHomeDir()
			templatesDir := filepath.Join(homeDir, ".moltbb", "templates")
			os.MkdirAll(templatesDir, 0755)
			templatePath := filepath.Join(templatesDir, templateName+".md")
			
			if _, err := os.Stat(templatePath); err == nil {
				output.PrintError("Template already exists")
				os.Exit(1)
			}
			
			content := defaultTemplates["daily"]
			ioutil.WriteFile(templatePath, []byte(content), 0644)
			output.Success("Template created: " + templateName)
			return nil
		},
	}
	return cmd
}

func newTemplateDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <template-name>",
		Short: "Delete a custom template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			if _, ok := defaultTemplates[templateName]; ok {
				output.PrintError("Cannot delete default templates")
				os.Exit(1)
			}
			homeDir, _ := os.UserHomeDir()
			templatePath := filepath.Join(homeDir, ".moltbb", "templates", templateName+".md")
			os.Remove(templatePath)
			output.Success("Template deleted: " + templateName)
			return nil
		},
	}
	return cmd
}

func replaceTemplatePlaceholders(content, date string) string {
	if date == "" {
		date = "2026-01-01"
	}
	content = strings.ReplaceAll(content, "{{.Date}}", date)
	content = strings.ReplaceAll(content, "{{.Week}}", "1")
	content = strings.ReplaceAll(content, "{{.DateRange}}", "Jan 1-7")
	return content
}
