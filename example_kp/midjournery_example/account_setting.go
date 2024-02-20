package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/tgis/thenextleg"
	"os"
)

func main() {
	token := os.Getenv("MID_TOKEN")
	fmt.Println(token)
	//// 生成图片任务和查询进度  ----start
	c, err := thenextleg.New(thenextleg.WithAuthToken(token))
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//resp, err := c.AccountSettings(context.Background())
	////setAccountSettingsPayload := &thenextleg.SettingsRequest{
	////	SettingsToggle: "MJ version 5",
	////	//Ref:             "",
	////	//WebhookOverride: "",
	////}
	////resp, err := c.SetAccountSettings(context.Background(), setAccountSettingsPayload)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Printf("%#v\n", resp)
	//
	//// get message
	//
	//msgId := resp.MessageId
	//
	//if msgId == "" {
	//	fmt.Println("message id is empty")
	//	return
	//}
	//respp, err := helpers.GetMessage(c, msgId)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Printf("%#v\n", respp)

	// face swap
	faceReq := &thenextleg.FaceSwapRequest{
		SourceImg: "https://hips.hearstapps.com/hmg-prod/images/ariana_grande_photo_jon_kopaloff_getty_images_465687098.jpg",
		TargetImg: "https://imageio.forbes.com/specials-images/imageserve/6474d985fece284f32569957/0x0.jpg?format=jpg&width=1200",
	}
	faceResp, err := c.FaceSwap(context.Background(), faceReq)
	if err != nil {
		fmt.Println()
		fmt.Println("error:", err.Error())
		return
	}
	err = os.WriteFile("./aaa.jpg", faceResp, os.ModePerm)
	if err != nil {
		fmt.Println()
		fmt.Println("error2:", err.Error())
	}
}
