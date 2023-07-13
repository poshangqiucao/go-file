package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/poshangqiucao/go-file/utils"
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

	r.GET("/search", func(c *gin.Context) {
		c.HTML(http.StatusOK, "search.html", nil)
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

		// 索引文件信息到gofound
		err = utils.IndexFileInfo(targetPath, file.Filename, "test")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.HTML(http.StatusOK, "success.html", nil)
	})

	r.GET("/s", func(c *gin.Context) {
		q := c.Query("q")
		result, err := utils.QueryIndex(q, "test")
		if err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, result)

	})

	r.GET("/download", func(c *gin.Context) {
		filePath := c.Query("filepath") // 获取文件名参数

		// 打开文件
		file, err := os.Open(filePath)
		if err != nil {
			log.Println("打开文件出错:", err)
			c.String(http.StatusInternalServerError, "文件打开失败")
			return
		}
		defer file.Close()

		filename := filepath.Base(filePath)
		// 设置响应头
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Length", "")

		// 将文件内容复制到响应中
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			log.Println("写入响应出错:", err)
			return
		}
	})

	// 启动服务器
	fmt.Println("服务器已启动，监听端口 10011")
	r.Run(":10011")
}
