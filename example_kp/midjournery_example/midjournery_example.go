package main

import (
	"context"
	"example_kp/midjournery_example/internal/helpers"
	"fmt"
	"github.com/tmc/langchaingo/tgis/thenextleg"
	"os"
)

func main() {

	//	var ssp thenextleg.MessageResponse
	//	s := `{"progress":100,
	//"response":{"accountId":"2HNhG5DksW4OLPtjG3ZP","createdAt":"2023-10-24T11:34:09.667Z","originatingMessageId":"Hk1aJbDvdV2MrW9qolRN","buttons":["U1","U2","U3","U4","🔄","V1","V2","V3","V4"],"imageUrl":"https://cdn.discordapp.com/attachments/1152602889924124682/1166338842316509254/freedomlink._Carp_in_the_Lotus_Pond_f76c6db0-a743-49f3-9347-0593a3091b23.png?ex=654a20b1&is=6537abb1&hm=efda793c5ba9188c185725ca7cb8a52d972a73f0eada15a476f6e8f266b3ee5a&","imageUrls":["https://cdn.midjourney.com/f76c6db0-a743-49f3-9347-0593a3091b23/0_0.png","https://cdn.midjourney.com/f76c6db0-a743-49f3-9347-0593a3091b23/0_1.png","https://cdn.midjourney.com/f76c6db0-a743-49f3-9347-0593a3091b23/0_2.png","https://cdn.midjourney.com/f76c6db0-a743-49f3-9347-0593a3091b23/0_3.png"],"responseAt":"2023-10-24T11:34:10.190Z","description":"","type":"imagine","content":"Carp in the Lotus Pond","buttonMessageId":"INNNWP6Nkfo2cixjkWx5"}}`
	//	err := json.Unmarshal([]byte(s), &ssp)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	fmt.Println(ssp)
	//	return

	token := os.Getenv("MID_TOKEN")
	// 生成图片任务和查询进度  ----start
	c, err := thenextleg.New(thenextleg.WithAuthToken(token))
	if err != nil {
		fmt.Println(err.Error())
	}
	payload := &thenextleg.ImagineRequest{
		Msg: "Carp in the Lotus Pond",
		//Ref:             "",
		//WebhookOverride: "",
		//IgnorePrefilter: "",
	}
	resp, err := c.Imagine(context.Background(), payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n", resp)

	// get message

	msgId := resp.MessageId

	if msgId == "" {
		fmt.Println("message id is empty")
		return
	}
	respp, err := helpers.GetMessage(c, msgId)
	if err != nil {
		fmt.Println(err)
		return
	}

	if respp.Response.ButtonMessageId == "" {
		fmt.Println("fail")
		return
	}
	// 生成图片任务和查询进度  -------end

	// 对生成的图片微调任务，和微调任务进度查询   -----start
	// buttons  INNNWP6Nkfo2cixjkWx5
	btnMessageid := respp.Response.ButtonMessageId
	//btnMessageid := "INNNWP6Nkfo2cixjkWx5"

	butPayload := &thenextleg.ButtonRequest{
		ButtonMessageId: btnMessageid,
		Button:          "U1",
	}
	btnResp, err := c.Button(context.Background(), butPayload)

	bgtMsgId := btnResp.MessageId
	helpers.GetMessage(c, bgtMsgId)

	// 对生成的图片微调任务，和微调任务进度查询   -----end
}
