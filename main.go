package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func FindQRExecutable() (string, error) {
	path, err := exec.LookPath("qr")
	if err != nil {
		return "", fmt.Errorf("未找到 qr 命令: %v", err)
	}
	return path, nil
}

// https://github.com/soulteary/certs-maker/blob/main/internal/fn/execute.go
func Execute(args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("sh", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

const DefaultTemplate = `
<!DOCTYPE html>
<html lang="en"%s>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>QR Code</title>
</head>
<body>
    <pre style="line-height: 0.92;user-select: none;text-shadow: 0 0 black;">
%s
    </pre>
</body>
</html>`

const DarkTemplate = ` style="filter: invert(1);background: #fff;"`

type QROptions struct {
	Animated  bool   `json:"animated"`
	Mode      string `json:"mode"`
	Version   int    `json:"version"`
	ECLevel   string `json:"ecLevel"`
	Large     bool   `json:"large"`
	Compact   bool   `json:"compact"`
	Border    int    `json:"border"`
	Invert    bool   `json:"invert"`
	Colorless bool   `json:"colorless"`
	UTF8BOM   bool   `json:"utf8bom"`
}

func BuildQRCommand(qrPath string, text string, opts QROptions) string {
	args := []string{}

	if opts.Animated {
		args = append(args, "-a")
	}
	if opts.Mode != "" {
		args = append(args, "-m", opts.Mode)
	}
	if opts.Version > 0 {
		args = append(args, "-v", strconv.Itoa(opts.Version))
	}
	if opts.ECLevel != "" {
		args = append(args, "-e", opts.ECLevel)
	}
	if opts.Large {
		args = append(args, "-l")
	}
	if opts.Compact {
		args = append(args, "-c")
	}
	if opts.Border > 0 {
		args = append(args, "-b", strconv.Itoa(opts.Border))
	}
	if opts.Invert {
		args = append(args, "-i")
	}
	if opts.Colorless {
		args = append(args, "-p")
	}
	if opts.UTF8BOM {
		args = append(args, "-u")
	}

	// 拼接完整命令
	argsStr := strings.Join(args, " ")
	return fmt.Sprintf("echo '%s' | %s %s", text, qrPath, argsStr)
}
func SetupRouter() *gin.Engine {
	qrPath, err := FindQRExecutable()
	if err != nil {
		fmt.Printf("Error in %s: %v\n", "Echo Test", err)
		return nil
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		result := ""

		// 从查询参数中获取选项
		opts := QROptions{
			Border: 1, // 默认边框宽度
		}

		// 处理 bool 类型参数
		if animated := c.Query("animated"); animated == "true" {
			opts.Animated = true
		}
		if large := c.Query("large"); large == "true" {
			opts.Large = true
		}
		if compact := c.Query("compact"); compact == "true" {
			opts.Compact = true
		}
		if invert := c.Query("invert"); invert == "true" {
			opts.Invert = true
		}
		if colorless := c.Query("colorless"); colorless == "true" {
			opts.Colorless = true
		}
		if utf8bom := c.Query("utf8bom"); utf8bom == "true" {
			opts.UTF8BOM = true
		}

		// 处理字符串类型参数
		if mode := c.Query("mode"); mode != "" {
			if strings.Contains("na8k", mode) { // 验证模式是否有效
				opts.Mode = mode
			}
		}
		if ecLevel := c.Query("ecLevel"); ecLevel != "" {
			if strings.Contains("lmqh1234", ecLevel) { // 验证EC级别是否有效
				opts.ECLevel = ecLevel
			}
		}

		// 处理整数类型参数
		if version := c.Query("version"); version != "" {
			if v, err := strconv.Atoi(version); err == nil && v >= 1 && v <= 40 {
				opts.Version = v
			}
		}
		if border := c.Query("border"); border != "" {
			if b, err := strconv.Atoi(border); err == nil && b >= 1 && b <= 4 {
				opts.Border = b
			}
		}

		// 获取文本内容，如果没有提供则使用默认值
		text := c.Query("text")
		if text == "" {
			text = "https://mp.weixin.qq.com/s/CLCBKrDANQsuFhGYAujzaQ"
		}

		commandTpl := BuildQRCommand(qrPath, text, opts)
		command := []string{"-c", commandTpl}

		output, err := Execute(command...)
		if err != nil {
			fmt.Printf("Error in %s: %v\n", "Echo Test", err)
		} else {
			if len(output) > 0 {
				fmt.Printf("Output:\n%s\n", output)
				dark := DarkTemplate
				result += fmt.Sprintf(DefaultTemplate, dark, output)
			}
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(result))
	})

	return r
}

func main() {
	r := SetupRouter()
	if r != nil {
		r.Run(":8083")
	}
}
