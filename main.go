package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
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

func SetupRouter() *gin.Engine {
	qrPath, err := FindQRExecutable()
	if err != nil {
		fmt.Printf("Error in %s: %v\n", "Echo Test", err)
		return nil
	}

	commandTpl := strings.Replace("echo 'https://mp.weixin.qq.com/s/CLCBKrDANQsuFhGYAujzaQ' | %qrPath% -b 4", "%qrPath%", qrPath, -1)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		result := ""

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
