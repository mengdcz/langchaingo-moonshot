package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/tgis/baiduai"
	"os"
	"time"
)

func main() {
	apiKey := os.Getenv("BAIDU_TXTIMG_KEY")
	apiSecret := os.Getenv("BAIDU_TXTIMG_SECRET")

	client, err := baiduai.NewClient(apiKey, apiSecret)
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()
	r := &baiduai.TxtRequest{
		Prompt:   "海盗船，怪物攻击，暴风雨，夜晚，完美无瑕战斗，日系动漫风，双重曝光，超宽视角",
		Version:  "v2",
		Width:    1024,
		Height:   1024,
		ImageNum: 2,
	}
	resp, err := client.Txt2imgv2(ctx, r)
	if err != nil {
		fmt.Println("error===", err)
		return
	}
	fmt.Println(resp)
	taskId := resp.Data.TaskId
	fmt.Println("taskId:", taskId)

	result, err := client.GetImgv2(ctx, taskId)
	if err != nil {
		fmt.Println(err)
		return
	}
	b, err := json.Marshal(result)

	fmt.Printf("%#v", string(b))

	//r1 := &baiduai.Txtv1Request{
	//	Text:       "海盗船，怪物攻击，暴风雨，夜晚，完美无瑕战斗，日系动漫风，双重曝光，超宽视角",
	//	Style:      "水彩画",
	//	Resolution: "1024*1024",
	//	Num:        2,
	//}
	//
	//resp, err := client.Txt2img(ctx, r1)
	//if err != nil {
	//	fmt.Println("error === ", err)
	//	return
	//}
	//bb, _ := json.Marshal(resp)
	//fmt.Printf(string(bb))
	//taskIdInt := resp.Data.TaskId
	//taskId := strconv.FormatInt(taskIdInt, 10)

	////// 17722501
	//taskId := "17722616"
	fmt.Println("task id:", taskId)

	if taskId == "" {
		fmt.Println("task id is empty ")
		return
	}
	for i := 0; i < 10; i++ {
		resp1, err1 := client.GetImg(ctx, taskId)
		if err1 != nil {
			fmt.Println(err1)
			break
		}
		b, _ := json.Marshal(resp1)
		fmt.Printf("%\n", string(b))
		time.Sleep(time.Second * 3)
	}

	fmt.Printf("time over")

	//body := `{"data":{"style":"水彩画","taskId":"17722584","imgUrls":[],"text":"海盗船，怪物攻击，暴风雨，夜晚，完美无瑕战斗，日系动漫风，双重曝光，超宽视角","status":0,"createTime":"30s"},"log_id":1712409383202814549}`

}
