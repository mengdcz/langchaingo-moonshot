package clipdropapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	clipdropapiParams "github.com/tmc/langchaingo/tgis/clipdropapi/params"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	baseUrl              = "https://clipdrop-api.co/"
	cleanupAPI           = "cleanup/v1"                         // 清理
	imageUpscaleAPI      = "image-upscaling/v1/upscale"         // 图像放大
	portraitDepthAPI     = "portrait-depth-estimation/v1"       // 人像深度估计
	portraitSurfaceAPI   = "portrait-surface-normals/v1"        // 肖像表面法线
	reimagineAPI         = "reimagine/v1/reimagine"             // 重新想象
	removeBackgroundAPI  = "remove-background/v1"               // 删除背景
	removeTextAPI        = "remove-text/v1"                     // 删除文本
	replaceBackgroundAPI = "replace-background/v1"              // 替换背景
	sketchToImageAPI     = "sketch-to-image/v1/sketch-to-image" // 草图到图
	textInpaintAPI       = "text-inpainting/v1"                 // 文本修复
	textToImageAPI       = "text-to-image/v1"                   // 文生图
	unCropAPI            = "uncrop/v1"                          // 取消剪裁 2k
)

type ClipDropApi struct {
	baseUrl    string // 请求url
	authToken  string
	httpClient Doer
	RespHeader map[string][]string
}

type Option func(*ClipDropApi)

func WithAuthToken(token string) Option {
	return func(leg *ClipDropApi) {
		leg.authToken = token
	}
}

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// New 创建TheNextLeg 客户端实例
func New(opts ...Option) (*ClipDropApi, error) {
	c := &ClipDropApi{}
	c.RespHeader = make(map[string][]string)

	for _, v := range opts {
		v(c)
	}
	if c.baseUrl == "" {
		c.baseUrl = baseUrl
	}
	if c.authToken == "" {
		return nil, errors.New("缺少token")
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c, nil
}

func (t *ClipDropApi) Cleanup(ctx context.Context, imageRequest *clipdropapiParams.CleanupRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", baseUrl, cleanupAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("Cleanup createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("Cleanup doHttp err", err.Error())
		return imagesResponse, err
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) ImageUpscale(ctx context.Context, imageRequest *clipdropapiParams.ImageUpscaleRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, imageUpscaleAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("ImageUpscale createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("ImageUpscale doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) PortraitDepth(ctx context.Context, imageRequest *clipdropapiParams.PortraitDepthEstimationRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, portraitDepthAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("PortraitDepth createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("PortraitDepth doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) PortraitSurface(ctx context.Context, imageRequest *clipdropapiParams.PortraitSurfaceNormalsRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, portraitSurfaceAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("PortraitSurface createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("PortraitSurface doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) Reimagine(ctx context.Context, imageRequest *clipdropapiParams.ReimagineRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, reimagineAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("Reimagine createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("Reimagine doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) RemoveBackground(ctx context.Context, imageRequest *clipdropapiParams.RemoveBackgroundRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, removeBackgroundAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("RemoveBackground createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("RemoveBackground doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) RemoveText(ctx context.Context, imageRequest *clipdropapiParams.RemoveTextRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, removeTextAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("RemoveText createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("RemoveText doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) ReplaceBackground(ctx context.Context, imageRequest *clipdropapiParams.ReplaceBackgroundRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, replaceBackgroundAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("ReplaceBackground createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) SketchToImage(ctx context.Context, imageRequest *clipdropapiParams.SketchToImageRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, sketchToImageAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("SketchToImage createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("SketchToImage doHttp err", err.Error())
		return
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) TextInpainting(ctx context.Context, imageRequest *clipdropapiParams.TextInpaintingRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, textInpaintAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("TextInpainting createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("TextInpainting doHttp err", err.Error())
		return imagesResponse, err
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) UnCrop(ctx context.Context, imageRequest *clipdropapiParams.UnCropRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, unCropAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imageRequest)
	if err != nil {
		fmt.Println("UnCrop SketchToImage createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("UnCrop doHttp err", err.Error())
		return imagesResponse, err
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err
}

func (t *ClipDropApi) Images(ctx context.Context, imagineRequest *clipdropapiParams.ImagesRequest) (imagesResponse *clipdropapiParams.ImagesResponse, err error) {
	url := fmt.Sprintf("%s/%s", t.baseUrl, textToImageAPI)
	bodyBuf, bodyWriter, err := t.createMultipartFormData(imagineRequest)
	if err != nil {
		fmt.Println("Images createMultipartFormData err", err.Error())
		return imagesResponse, err
	}

	contentType := bodyWriter.FormDataContentType()

	imagesResponse = &clipdropapiParams.ImagesResponse{}
	var b []byte
	if b, err = t.doHttp(ctx, url, http.MethodPost, bodyBuf, nil, imagesResponse, contentType); err != nil {
		fmt.Println("Images doHttp err", err.Error())
		return imagesResponse, err
	}

	imagesResponse.ImgFile = b
	imagesResponse.Success = true

	return imagesResponse, err

}

func (t *ClipDropApi) doHttp(ctx context.Context, url string, method string, body *bytes.Buffer, resp any, imagesResponse *clipdropapiParams.ImagesResponse, contentType string) (b []byte, err error) {
	fmt.Printf("ClipDropApi doHttp start, contentType %v \n", contentType)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		fmt.Println("ClipDropApi http.NewRequest err ", err.Error())
		return b, err
	}
	t.setHeader(req, contentType)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("ClipDropApi client.Do err ", err.Error())

		return b, err
	}
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			fmt.Println(err1)
		}
	}(response.Body)

	b, err = io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(" ClipDropApi io.ReadAll err ", err.Error())
		return b, err
	}

	if response.StatusCode != http.StatusOK {

		fmt.Printf("ClipDropApi API returned status code: %v, StatusCode %v, err %v \n", response.Status, response.StatusCode, string(b))
		errorResp := &clipdropapiParams.ErrorResponse{}
		if errJson := json.Unmarshal(b, errorResp); errJson != nil {
			fmt.Printf("ClipDropApi API returned err json.Unmarshal err %v, msg %v \n", errJson, string(b))
		}
		imagesResponse.Error = fmt.Sprintf("API returned %v; msg: %v", response.Status, errorResp.Error)
		err = errors.New(imagesResponse.Error)
		return b, err
	}

	t.setRespHeader(response, imagesResponse)

	fmt.Printf("ClipDropApi doHttp end \n")

	return b, err
}

func (t *ClipDropApi) setRespHeader(response *http.Response, ImagesResponse *clipdropapiParams.ImagesResponse) {
	if len(response.Header) < 1 {
		return
	}

	headers := make(map[string][]string)
	for key, values := range response.Header {
		headers[key] = append(values[:0], values...)
	}

	if XCreditsConsumed, ok := headers["X-Credits-Consumed"]; ok {
		fmt.Printf("XCreditsConsumed %v, type %T, string %s ,  value  %v \n", XCreditsConsumed, XCreditsConsumed, XCreditsConsumed, XCreditsConsumed[0])
		ImagesResponse.XReditsConsumed = XCreditsConsumed[0]
	}
	if XRemainingCredits, ok := headers["X-Remaining-Credits"]; ok {
		fmt.Printf("XRemainingCredits %v, type %T, string %s ,  value  %v \n", XRemainingCredits, XRemainingCredits, XRemainingCredits, XRemainingCredits[0])
		ImagesResponse.XRemainingCredits = XRemainingCredits[0]

	}
	if ContentType, ok := headers["Content-Type"]; ok {
		fmt.Printf("ContentType %v, type %T, string %s ,  value  %v \n", ContentType, ContentType, ContentType, ContentType[0])
		contentTypeStr := ContentType[0]
		if len(contentTypeStr) > 0 {
			checkContentType := strings.Contains(contentTypeStr, "/")
			if checkContentType {
				comma := strings.Index(contentTypeStr, "/")
				ImagesResponse.ImgExt = contentTypeStr[comma+1:]
			}
		}
	}
	return
}

func (t *ClipDropApi) createMultipartFormData(params any) (*bytes.Buffer, *multipart.Writer, error) {
	fmt.Println("createMultipartFormData start")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	defer func() {
		errWriterClose := writer.Close()
		if errWriterClose != nil {
			fmt.Printf("writer.Close err,err %v \n", errWriterClose.Error())
		}
	}()

	// 模拟
	//  reflect.ValueOf {[] 1000 0 0 0}
	//value.Type()  params.UncropRequest
	//Field = ImageFile, Value = [] tagValue = image_file, fieldForm = image_file, ftype = []uint8
	//Field = ExtendLeft, Value = 1000 tagValue = extend_left,omitempty, fieldForm = extend_left, ftype = int64
	//Field = ExtendRight, Value = 0 tagValue = extend_right,omitempty, fieldForm = extend_right, ftype = int64
	//Field = ExtendUp, Value = 0 tagValue = extend_up,omitempty, fieldForm = extend_up, ftype = int64
	//Field = ExtendDown, Value = 0 tagValue = extend_down,omitempty, fieldForm = extend_down, ftype = int64

	// 获取结构体的反射值
	value := reflect.ValueOf(params)
	// 获取结构体的反射类型
	//typ := value.Type()
	// 如果传入的是结构体指针，则获取其指向的结构体值
	if value.Kind() == reflect.Ptr {
		fmt.Println("val.Kind() == reflect.Ptr val.Elem()", value.Elem())
		value = value.Elem()
	}

	// 确保现在处理的是结构体值
	if value.Kind() != reflect.Struct {
		fmt.Println("Error: expected a struct")
		err := errors.New("Error: expected a struct")
		return body, writer, err
	}

	// 遍历结构体的字段
	for i := 0; i < value.NumField(); i++ {
		//field := typ.Field(i)
		field := value.Type().Field(i)
		fieldValue := value.Field(i)
		tagValue := field.Tag.Get("json")
		fieldForm := tagValue
		fType := fieldValue.Type()
		checkParamsEmpty := strings.Contains(tagValue, "omitempty")
		checkParamsFormat := strings.Contains(tagValue, ",")
		if len(tagValue) > 0 {
			if checkParamsFormat {
				tagVal := strings.Split(tagValue, ",")
				fieldForm = tagVal[0]
			}
		}

		//fmt.Printf("Field = %s, Value = %v tagValue = %v, fieldForm = %v, ftype = %v  \n", field.Name, fieldValue.Interface(), tagValue, fieldForm, fType)

		var fValue string
		checkFile := strings.Contains(tagValue, "file")
		if checkFile { // 文件处理方式
			fValue = fieldValue.Interface().(string)

			part, err := writer.CreateFormFile(fieldForm, fValue)
			if err != nil {
				fmt.Println("createMultipartFormData  writer.CreateFormFile ===", fieldForm, fValue, err.Error())
				return body, writer, err
			}
			fileContent, err := os.Open(fValue)
			if err != nil {
				fmt.Println("createMultipartFormData os.Open ===", fieldForm, fValue, err.Error())
				return body, writer, err
			}
			_, err = io.Copy(part, fileContent)
			_ = fileContent.Close()
			if err != nil {
				fmt.Println("createMultipartFormData io.Copy ===", fieldForm, fValue, err.Error())
				return body, writer, err
			}
		} else { // 非文件类型
			if fType.Kind() == reflect.Int64 {
				valueInt64 := fieldValue.Interface().(int64)
				fValue = strconv.Itoa(int(valueInt64))
				if checkParamsEmpty == false && len(fValue) < 1 { // 必填参数不能为空
					err := errors.New("params is empty")
					return body, writer, err
				}
				if valueInt64 == 0 { // 0 数据跳过
					continue
				}
			}
			if fType.Kind() == reflect.String {
				fValue = fieldValue.Interface().(string)
				if checkParamsEmpty == false && len(fValue) < 1 { // 必填参数不能为空
					err := errors.New("params is empty")
					return body, writer, err
				}
				if len(fValue) < 1 { // 空数据跳过
					continue
				}
			}

			err := writer.WriteField(fieldForm, fValue)
			if err != nil {
				fmt.Println("bodyWriter.WriteField ===", fieldForm, fValue, err.Error())
				return body, writer, err
			}
		}
	}

	fmt.Println("createMultipartFormData end")

	return body, writer, nil
}

func (t *ClipDropApi) setHeader(req *http.Request, contentType string) {
	req.Header.Set("x-api-key", t.authToken)
	req.Header.Set("Content-Type", contentType)
}
