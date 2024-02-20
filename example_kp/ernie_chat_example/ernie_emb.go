package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ernieembedding "github.com/tmc/langchaingo/embeddings/ernie"
	"github.com/tmc/langchaingo/llms/ernie"
	"golang.org/x/sys/windows"
	"io"
	"net/http"
	"os"
)

func main() {
	// 大文本向量化
	//filePath := "./test.txt"
	//b, err := os.ReadFile(filePath)
	//if err != nil {
	//	fmt.Print(err)
	//	return
	//}
	//content := string(b)

	content := []string{"曹云金被污蔑三件事",
		"相声比赛退赛",
		"短信证据",
		"学费问题",
		"发票证据",
		"离开后的短信问题",
		"郭德纲言论：冤枉你的人，比你知道有多冤枉",
		"对师父的不满",
		"反感劝大度的人",
		"曹云金有证据",
		"曹云金被污蔑，夫妻俩阴阳怪气",
		"曹云金口碑反转，被广大网友冠名为'反职场霸凌第一人'"}
	// 向量合并
	k := os.Getenv("ERNIE_API_KEY")
	v := os.Getenv("ERNIE_SECRET_KEY")
	c, err := ernie.New(ernie.WithAKSK(k, v))
	if err != nil {
		fmt.Println(err)
		return
	}
	emb, err := ernieembedding.NewErnie(ernieembedding.WithClient(*c))
	if err != nil {
		fmt.Println(err)
		return
	}
	//res, err := emb.EmbedQuery(context.Background(), content)
	res, err := emb.EmbedCombine(context.Background(), content)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 向量入库 , 创建集合， 添加
	payload := &payloadPoint{Points: []point{
		{
			Id:     103,
			Vector: res,
			Payload: map[string]interface{}{
				"content": "",
			},
		},
	}}

	bb, err := json.Marshal(payload)
	os.WriteFile("emb.txt", bb, windows.FILE_ACTION_ADDED)

	os.WriteFile("emb.txt", []byte("\n"), windows.FILE_ACTION_ADDED)
	addPoint(payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 向量查询
	fmt.Println("todo 向量查询")
}

//func createCollection(collectionName string) {
//	url := fmt.Sprintf("http://192.168.5.89:6333/collections/ernie_examples")
//	payload := `{"name":"ernie_examples","vectors":{ "size": 384, "distance": "Cosine" }}`
//
//	res, err := doHttp(url, http.MethodPut, []byte(payload))
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println("创建集合成功：", string(res))
//}

type payloadPoint struct {
	Points []point `json:"points"`
}

type point struct {
	Id      int                    `json:"id"`
	Vector  []float64              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

func addPoint(payload *payloadPoint) {
	url := fmt.Sprintf("http://192.168.5.89:6333/collections/ernie_examples/points")

	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := doHttp(url, http.MethodPut, b)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("添加point成功：", string(res))
}

func doHttp(url, method string, payloadBytes []byte) (b []byte, err error) {
	// Build request
	var body io.Reader
	if payloadBytes != nil {
		body = bytes.NewReader(payloadBytes)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return b, err
	}
	//req.Header.Set("Authorization", "")
	req.Header.Set("Content-Type", "application/json")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return b, err
	}
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			fmt.Println(err1)
		}
	}(r.Body)
	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)
		return b, errors.New(msg)
	}
	b, err = io.ReadAll(r.Body)
	if err != nil {
		return b, err
	}
	fmt.Println(string(b))
	return b, err
}
