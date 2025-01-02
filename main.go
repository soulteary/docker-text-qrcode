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
	if opts.Version > 0 && opts.Version <= 40 {
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
	if opts.Border > 0 && opts.Border <= 4 {
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

	argsStr := strings.Join(args, " ")
	return fmt.Sprintf("echo '%s' | %s %s", text, qrPath, argsStr)
}

func IsTriggerOn(q string) bool {
	input := strings.ToLower(q)
	return input == "true" || input == "on" || input == "yes" || input == "1"
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

		opts := QROptions{
			Mode:      "8",   // -m 默认是 '8' (8-bit模式)
			Version:   0,     //-v 默认是 0 (自动选择版本)
			ECLevel:   "1",   // -e 默认是 '1' (L级别纠错)
			Large:     false, // -l 默认是 false (不使用巨型模式)
			Compact:   false, // -c 默认是 false (不使用紧凑模式)
			Border:    1,     // -b 默认是 1 (边框宽度为1)
			Invert:    false, // -i 默认是 false (不反转颜色)
			Colorless: false, // -p 默认是 false (有颜色输出,除非不是终端)
			UTF8BOM:   false, // -u 默认是 false (不输出UTF-8 BOM)
			Animated:  false, // 默认不启用动画
		}

		if mode := strings.ToLower(c.Query("mode")); mode != "" {
			if mode == "n" || mode == "a" || mode == "8" || mode == "k" {
				opts.Mode = mode
			}
		}

		if version := c.Query("version"); version != "" {
			if v, err := strconv.Atoi(version); err == nil && v >= 1 && v <= 40 {
				opts.Version = v
			}
		}

		if ecLevel := strings.ToLower(c.Query("ecLevel")); ecLevel != "" {
			if ecLevel == "l" || ecLevel == "m" || ecLevel == "q" || ecLevel == "h" {
				opts.ECLevel = ecLevel
			}
			if ecLevel == "1" || ecLevel == "2" || ecLevel == "3" || ecLevel == "4" {
				opts.ECLevel = ecLevel
			}
		}

		if large := c.Query("large"); IsTriggerOn(large) {
			opts.Large = true
		}

		if compact := c.Query("compact"); IsTriggerOn(compact) {
			opts.Compact = true
		}

		if border := c.Query("border"); border != "" {
			if b, err := strconv.Atoi(border); err == nil && b >= 1 && b <= 4 {
				opts.Border = b
			}
		}

		if invert := c.Query("invert"); IsTriggerOn(invert) {
			opts.Invert = true
		}

		if colorless := c.Query("colorless"); IsTriggerOn(colorless) {
			opts.Colorless = true
		}

		if utf8bom := c.Query("utf8bom"); IsTriggerOn(utf8bom) {
			opts.UTF8BOM = true
		}

		if animated := c.Query("animated"); IsTriggerOn(animated) {
			opts.Animated = true
		}

		text := c.Query("text")
		if text == "" {
			text = "https://github.com/soulteary/docker-text-qrcode"
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
