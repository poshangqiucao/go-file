package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed static/* templates/*
var embeddedFiles embed.FS

func main() {
	// 创建 Gin 引擎
	r := gin.Default()

	// 加载模板
	templateEngine := template.New("")
	templateEngine = templateEngine.Funcs(template.FuncMap{
		"static": func(filePath string) string {
			return "/public/static/" + filePath
		},
	})
	templateEngine, err := templateEngine.ParseFS(embeddedFiles, "templates/*.html")
	if err != nil {
		panic(err)
	}
	r.SetHTMLTemplate(templateEngine)
	r.StaticFS("/public", http.FS(embeddedFiles))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload.html", nil)
	})

	// 定义文件上传路由
	r.POST("/upload", func(c *gin.Context) {
		// 解析表单数据
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取目标路径参数
		targetPath := c.PostForm("targetPath")
		if targetPath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "目标路径不能为空"})
			return
		}

		// 检查目标路径是否存在，不存在则创建
		err = os.MkdirAll(targetPath, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 保存文件到目标路径
		err = c.SaveUploadedFile(file, targetPath+"/"+file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "success.html", nil)
	})

	// 启动服务器
	fmt.Println("服务器已启动，监听端口 10011")
	r.Run(":10011")
}
