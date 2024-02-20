/*
 * Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hunyuan

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms/hunyuan/internal/hysdk"

	"time"
)

func synchronizeChat(client *hysdk.TencentHyChat) {
	resp, err := client.Chat(context.Background(), hysdk.NewRequest(hysdk.Synchronize, []hysdk.Message{
		{
			Role:    "user",
			Content: "给我推荐一首歌",
		},
	}))
	if err != nil {
		fmt.Printf("tencent hunyuan chat err:%+v\n", err)
		return
	}
	fmt.Println("同步访问结果: ")
	fmt.Println(resp.Choices[0].Messages.Content)

	//fmt.Println("同步访问结果: ")
	//for res := range resp {
	//	if res.Error.Code != 0 {
	//		fmt.Printf("tencent hunyuan chat err:%+v\n", res.Error)
	//		break
	//	}
	//	//synchronize 同步打印message
	//	fmt.Println(res.Choices[0].Messages.Content)
	//}
}

func streamChat(client *hysdk.TencentHyChat) {
	messages := []hysdk.Message{
		{
			Role:    "user",
			Content: "给我推荐一首歌",
		},
		{
			Role:    "assistant",
			Content: "消愁",
		},
		{
			Role:    "user",
			Content: "给我推荐消愁作者的其它歌曲",
		},
	}
	queryID := uuid.NewString()
	req := hysdk.Request{
		Timestamp:   int(time.Now().Unix()),
		Expired:     int(time.Now().Unix()) + 24*60*60,
		Temperature: 0,
		TopP:        0.8,
		Messages:    messages,
		QueryID:     queryID,
		Stream:      0,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			fmt.Println(string(chunk))
			return nil
		},
	}
	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		fmt.Printf("tencent hunyuan chat err:%+v\n", err)
		return
	}

	fmt.Println("所有访问结果: ")
	//stream  流式打印Delta
	fmt.Print(resp.Choices[0].Delta.Content)

}

func main() {
	//登陆控制台获取appID和密钥信息 替换下面的值
	var SecretID = "xxxx"
	var SecretKey = "xxxx"
	var appID int64 = 1122

	credential := hysdk.NewCredential(SecretID, SecretKey)
	client := hysdk.NewTencentHyChat(appID, credential)
	synchronizeChat(client) //同步访问
	streamChat(client)      //流式访问
}
