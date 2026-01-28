package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "https://ark.cn-beijing.volces.com/api/v3/contents/generations/tasks"
	// 使用用户提供的 Key
	apiKey := "c3f7f4b0-5072-47ee-b944-0a73f2377443"

	fmt.Printf("Testing VolcEngine API with Key: %s\n", apiKey)
	fmt.Printf("URL: %s\n", url)

	payload := map[string]interface{}{
		"model": "doubao-seedance-1-5-pro-251215",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "无人机以极快速度穿越复杂障碍或自然奇观，带来沉浸式飞行体验  --duration 5 --camerafixed false --watermark true",
			},
			{
				"type": "image_url",
				"image_url": map[string]string{
					"url": "https://ark-project.tos-cn-beijing.volces.com/doc_image/seepro_i2v.png",
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("JSON Marshal Error:", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("NewRequest Error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Client Do Error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ReadAll Error:", err)
		return
	}

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))
}
