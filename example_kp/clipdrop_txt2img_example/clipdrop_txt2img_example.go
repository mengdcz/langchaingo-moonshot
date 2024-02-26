package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/tgis/clipdropapi"
	clipdropapiParams "github.com/tmc/langchaingo/tgis/clipdropapi/params"
	"os"
	"strconv"
	"time"
)

const (
	textToImage       = "TextToImage"       // 文生图
	unCrop            = "UnCrop"            // 取消剪裁
	textInpainting    = "TextInpainting"    // 文本修复
	sketchToImage     = "SketchToImage"     // 草图到图
	replaceBackground = "ReplaceBackground" // 替换背景
	removeText        = "RemoveText"        // 删除文本
	removeBackground  = "RemoveBackground"  // 删除背景
	reimagine         = "Reimagine"         // 重新想象
	portraitSurface   = "PortraitSurface"   // 肖像表面法线
	portraitDepth     = "PortraitDepth"     // 人像深度估计
	imageUpscale      = "ImageUpscale"      // 图像放大
	cleanup           = "Cleanup"           // 清理
	saveFileDir       = "./images/"         // 保存目录 时间-图片类型
	imagesResource    = "./resource/"       // 原图片
)

var (
	clipdropApi *clipdropapi.ClipDropApi
	ctx         context.Context
)

func main() {
	apiKey := os.Getenv("CLIPDROP_KEY")
	fmt.Println(apiKey)

	clipdropLogic, err := clipdropapi.New(clipdropapi.WithAuthToken(apiKey))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	clipdropApi = clipdropLogic
	fmt.Println("clipdropApi", clipdropApi)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover ", r)
			msg := fmt.Sprintf("%v", r)
			err = errors.New(msg)
			return
		}
	}()

	ctx = context.Background()
	dType := unCrop // cleanup
	imagineResponse := &clipdropapiParams.ImagesResponse{}

	switch dType {
	case textToImage:
		imagineResponse, err = testImage()
		break
	case unCrop:
		imagineResponse, err = testUnCrop()

		break
	case textInpainting:
		imagineResponse, err = testTextInpainting()

		break
	case sketchToImage:
		imagineResponse, err = testSketchToImage()

		break
	case replaceBackground:
		imagineResponse, err = testReplaceBackground()

		break
	case removeText:
		imagineResponse, err = testRemoveText()

		break
	case removeBackground:
		imagineResponse, err = testRemoveBackground()

		break
	case reimagine:
		imagineResponse, err = testReimagine()

		break
	case portraitSurface:
		imagineResponse, err = testPortraitSurface()

		break
	case portraitDepth:
		imagineResponse, err = testPortraitDepth()

		break
	case imageUpscale:
		imagineResponse, err = testImageUpscale()

		break
	case cleanup:
		imagineResponse, err = testCleanup()

		break
	default:
		fmt.Println("please select the corresponding painting function")
		return
	}

	if err != nil {
		fmt.Printf("imagineResponse type %v exec , imagineResponse %v ,err %v \n ", dType, imagineResponse, err)
		return
	}

	err = saveFile(imagineResponse.ImgFile, imagineResponse.ImgExt, dType)
	if err != nil {
		fmt.Println("testExec saveFile err", err)
		return
	}

	fmt.Println("imagineResponse==imagineResponse.Success ", imagineResponse.Success)
	fmt.Println("imagineResponse==imagineResponse.XRemainingCredits ", imagineResponse.XRemainingCredits)
	fmt.Println("imagineResponse==imagineResponse.XReditsConsumed ", imagineResponse.XReditsConsumed)
	fmt.Println("imagineResponse==imagineResponse.ImgExt ", imagineResponse.ImgExt)
	fmt.Println("imagineResponse==imagineResponse.Error ", imagineResponse.Error)
}

func testCleanup() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.CleanupRequest{}
	imageRequest.ImageFile = imagesResource + "clean.jpeg"
	imageRequest.MaskFile = imagesResource + "clean-mask.png"
	imageRequest.Mode = "quality" //  可选字段，可以设置为fast或quality控制速度和质量之间的权衡, fast是默认模式，速度更快，但可能会在结果图像中产生伪影; quality速度较慢，但会产生更好的结果

	imagesResponse, err = clipdropApi.Cleanup(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.Cleanup err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testImageUpscale() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.ImageUpscaleRequest{}
	imageRequest.ImageFile = imagesResource + "image-upscaling.png"
	imageRequest.TargetWidth = 4096
	imageRequest.TargetHeight = 4096

	imagesResponse, err = clipdropApi.ImageUpscale(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.ImageUpscale err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testPortraitDepth() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.PortraitDepthEstimationRequest{}
	imageRequest.ImageFile = imagesResource + "reimagine_1024x1024.jpg"

	imagesResponse, err = clipdropApi.PortraitDepth(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.PortraitDepth err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testPortraitSurface() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.PortraitSurfaceNormalsRequest{}
	imageRequest.ImageFile = imagesResource + "reimagine_1024x1024.jpg"

	imagesResponse, err = clipdropApi.PortraitSurface(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.PortraitSurface err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testReimagine() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.ReimagineRequest{}
	imageRequest.ImageFile = imagesResource + "reimagine_1024x1024.jpg"

	imagesResponse, err = clipdropApi.Reimagine(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.Reimagine err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testRemoveBackground() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.RemoveBackgroundRequest{}
	imageRequest.ImageFile = imagesResource + "remove-background.jpeg"

	imagesResponse, err = clipdropApi.RemoveBackground(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.RemoveBackground err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testRemoveText() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.RemoveTextRequest{}
	imageRequest.ImageFile = imagesResource + "remove-text-2_923x693.png"

	imagesResponse, err = clipdropApi.RemoveText(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.RemoveText err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testReplaceBackground() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.ReplaceBackgroundRequest{}
	imageRequest.ImageFile = imagesResource + "replace-background.jpg"
	imageRequest.Prompt = "a cozy marble kitchen with wine glasses"

	imagesResponse, err = clipdropApi.ReplaceBackground(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.ReplaceBackground err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testSketchToImage() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.SketchToImageRequest{}
	imageRequest.SketchFile = imagesResource + "Sketch-to-image_1024x1024.png"
	imageRequest.Prompt = "an owl on a branch, cinematic"

	imagesResponse, err = clipdropApi.SketchToImage(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.SketchToImage err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testTextInpainting() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.TextInpaintingRequest{}
	imageRequest.ImageFile = imagesResource + "text-inpainting.jpeg"
	imageRequest.MaskFile = imagesResource + "TextInpainting-mask_file.png"
	imageRequest.TextPrompt = "A woman with a red scarf"

	imagesResponse, err = clipdropApi.TextInpainting(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.TextInpainting err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testUnCrop() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imageRequest := &clipdropapiParams.UnCropRequest{}
	imageRequest.ImageFile = imagesResource + "image-upscaling.png"
	imageRequest.ExtendLeft = 1000 // 可选 最大为 2k，默认为 0 【正负2k】
	imageRequest.ExtendRight = 1000
	imageRequest.ExtendUp = 0
	imageRequest.ExtendDown = 0

	imagesResponse, err = clipdropApi.UnCrop(ctx, imageRequest)
	if err != nil {
		fmt.Println("clipdropApi.UnCrop err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil
}

func testImage() (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	imagineRequest := &clipdropapiParams.ImagesRequest{}
	imagineRequest.Prompt = "shot of vaporwave fashion dog in miami"

	defer func() {
		if r := recover(); r != nil {
			//fmt.Println("recover ", r)
			msg := fmt.Sprintf("recover %v", r)
			err = errors.New(msg)
			return
		}
	}()

	imagesResponse, err = clipdropApi.Images(ctx, imagineRequest)
	if err != nil {
		fmt.Println("clipdropApi.Images err", err.Error())
		return imagesResponse, err
	}

	return imagesResponse, nil

}

func saveFile(b []byte, imgExt, dType string) (err error) {
	if len(imgExt) < 1 {
		err = errors.New("img ext empty")
		return
	}
	fileName := saveFileDir + strconv.Itoa(int(time.Now().Unix())) + "-" + dType + "." + imgExt

	err = os.WriteFile(fileName, b, os.ModePerm)
	if err != nil {
		fmt.Println("error2:", err.Error())
		return err
	}
	return nil
}
