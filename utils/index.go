package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func IndexFileInfo(targetPath, filename, database string) error {
	// 将文件信息索引到gofound中
	url := "http://localhost:8080/api/index?database=" + database

	// 准备 JSON 数据
	filepath := targetPath + "/" + filename
	docData := map[string]interface{}{
		"filepath": filepath,
	}
	jsonData := map[string]interface{}{
		"id":       time.Now().Unix(),
		"text":     filename,
		"document": docData,
	}

	// 将 JSON 数据编码
	requestBody, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 发送 POST 请求
	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal(err)
		return err
	}

	defer response.Body.Close()
	return nil
}

type DocumentData struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	Document struct {
		Filepath string `json:"filepath"`
	} `json:"document"`
	OriginalText string   `json:"originalText"`
	Score        int      `json:"score"`
	Keys         []string `json:"keys"`
}

// 定义结构体来映射返回结果
type ResponseData struct {
	State   bool   `json:"state"`
	Message string `json:"message"`
	Data    struct {
		Time      float64        `json:"time"`
		Total     int            `json:"total"`
		PageCount int            `json:"pageCount"`
		Page      int            `json:"page"`
		Limit     int            `json:"limit"`
		Documents []DocumentData `json:"documents"`
		Words     []string       `json:"words"`
	} `json:"data"`
}

func QueryIndex(query string, database string) ([]DocumentData, error) {
	url := "http://localhost:8080/api/query?database=" + database

	// 准备请求数据
	requestBody := map[string]interface{}{
		"query": query,
		"page":  1,
		"limit": 10,
		"order": "desc",
		"highlight": map[string]string{
			"preTag":  "<span style='color:red'>",
			"postTag": "</span>",
		},
	}

	// 将请求数据编码为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 发送 POST 请求
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer response.Body.Close()

	// 读取响应数据
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 解析响应数据
	var responseData ResponseData
	err = json.Unmarshal(responseBody, &responseData)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 获取 documents 字段
	documents := responseData.Data.Documents
	if documents == nil {
		documents = []DocumentData{}
	}
	return documents, nil

}
