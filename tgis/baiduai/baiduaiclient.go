package baiduai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	tokenUrl = "fdsf"
)

// Cache 公共缓存
type Cache interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Expire(key string, seconds int) error
}

// 文生图高级版
type BaiduAiClient struct {
	ClientId string
	Secret   string
	cache    Cache
	token    string
}
type TokenResponse struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	Scope         string `json:"scope"`
	SessionSecret string `json:"session_secret"`
}

func NewClient(clientId string, secret string) (*BaiduAiClient, error) {
	client := &BaiduAiClient{
		ClientId: clientId,
		Secret:   secret,
	}
	_, err := client.getAccessToken()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (b *BaiduAiClient) getAccessToken() (string, error) {
	if b.token != "" {
		return b.token, nil
	}
	key := "langchain:huihua:" + b.ClientId
	var token string
	var err error
	if b.cache != nil {
		token, _ = b.cache.Get(key)
		fmt.Println("image get token from cache:", token)
	}
	if token != "" {
		b.token = token
		return token, nil
	}

	url := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?client_id=%s&client_secret=%s&grant_type=client_credentials", b.ClientId, b.Secret)
	fmt.Println(url)
	payload := strings.NewReader(``)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	var resp TokenResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println(string(body))
	if resp.AccessToken != "" {
		b.token = resp.AccessToken
		if b.cache != nil {
			b.cache.Set(key, resp.AccessToken)
			b.cache.Expire(key, resp.ExpiresIn-3600)
		}
		return resp.AccessToken, nil
	}
	return "", errors.New("access token is empty")
}

type TxtRequest struct {
	Prompt       string `json:"prompt"`                  //生图的文本描述。仅支持中文、日常标点符号。不支持英文，特殊符号，限制 200 字
	Version      string `json:"version,omitempty"`       //模型版本，支持 v1、v2，默认为v2，v2 为最新模型，比 v1 在准确度、精细度上有比较明显的提升，且 v2 支持更多尺寸
	Width        int    `json:"width"`                   //图片宽度，v1 版本支持：1024x1024、1280x720、720x1280、2048x2048、2560x1440、1440x2560；v2 版本支持：512x512、640x360、360x640、1024x1024、1280x720、720x1280、2048x2048、2560x1440、1440x2560
	Height       int    `json:"height"`                  //图片高度，v1 版本支持：1024x1024、1280x720、720x1280、2048x2048、2560x1440、1440x2560；v2 版本支持：512x512、640x360、360x640、1024x1024、1280x720、720x1280、2048x2048、2560x1440、1440x2560
	ImageNum     int    `json:"image_num,omitempty"`     //生成图片数量，默认一张，支持生成 1-8 张
	Image        string `json:"image,omitempty"`         //和url/pdf_file 三选一    参考图，需 base64 编码，大小不超过 10M，最短边至少 15px，最长边最大 8192px，支持jpg/jpeg/png/bmp 格式。优先级：image > url > pdf_file，当image 字段存在时，url、pdf_file 字段失效
	Url          string `json:"url,omitempty"`           //参考图完整 url，url 长度不超过 1024 字节，url 对应的图片需 base64 编码，大小不超过 10M，最短边至少 15px，最长边最大8192px，支持 jpg/jpeg/png/bmp 格式。优先级：image > url > pdf_file，当image 字段存在时，url 字段失效请注意关闭 URL 防盗链
	PdfFile      string `json:"pdf_file,omitempty"`      //参考图 PDF 文件，base64 编码，大小不超过10M，最短边至少 15px，最长边最大 8192px 。优先级：image > url > pdf_file，当image 字段存在时，url、pdf_file 字段失效
	PdfFileNum   string `json:"pdf_file_num,omitempty"`  //需要识别的 PDF 文件的对应页码，当pdf_file 参数有效时，识别传入页码的对应页面内容，若不传入，则默认识别第 1 页
	ChangeDegree int    `json:"change_degree,omitempty"` //当 image、url或 pdf_file 字段存在时，为必需项   参考图影响因子，支持 1-10 内；数值越大参考图影响越大
}

type TxtResponse struct {
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	Data      struct {
		PrimaryTaskId int64  `json:"primary_task_id,omitempty"`
		TaskId        string `json:"task_id,omitempty"`
	} `json:"data,omitempty"`
	LogId int64 `json:"log_id,omitempty"`
}

// txt2imgv2  提交申请
func (b *BaiduAiClient) Txt2imgv2(ctx context.Context, r *TxtRequest) (*TxtResponse, error) {
	token, _ := b.getAccessToken()

	url := fmt.Sprintf("https://aip.baidubce.com/rpc/2.0/ernievilg/v1/txt2imgv2?access_token=%s", token)
	param, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(param))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(param))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	if res.StatusCode != http.StatusOK {
		err1 := getError(res.StatusCode)
		if err1 != nil {
			return nil, err1
		}
		return nil, errors.New(res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(string(body))
	var resp TxtResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.ErrorCode > 0 {
		err1 := getError(res.StatusCode)
		if err1 != nil {
			return nil, err1
		}
		return nil, errors.New(resp.ErrorMsg)
	}

	return &resp, nil
}

type Txtv1Response struct {
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	Data      struct {
		PrimaryTaskId int64 `json:"primary_task_id,omitempty"`
		TaskId        int64 `json:"taskId,omitempty"`
	} `json:"data,omitempty"`
	LogId int64 `json:"log_id,omitempty"`
}

type Txtv1Request struct {
	Text       string `json:"text"`                 //输入内容，长度不超过100个字（操作指南详见文档）
	Style      string `json:"style,omitempty"`      // 图片分辨率，可支持1024*1024、1024*1536、1536*1024
	Resolution string `json:"resolution,omitempty"` // 目前支持风格有：探索无限、古风、二次元、写实风格、浮世绘、low poly 、未来主义、像素风格、概念艺术、赛博朋克、洛丽塔风格、巴洛克风格、超现实主义、水彩画、蒸汽波艺术、油画、卡通画
	Num        int    `json:"num,omitempty"`        //图片生成数量，支持1-6张
}

// txt2imgv2  提交申请
func (b *BaiduAiClient) Txt2img(ctx context.Context, r *Txtv1Request) (*Txtv1Response, error) {
	token, _ := b.getAccessToken()

	url := fmt.Sprintf("https://aip.baidubce.com/rpc/2.0/ernievilg/v1/txt2img?access_token=%s", token)
	param, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(param))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader(param))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	if res.StatusCode != http.StatusOK {
		err1 := getError(res.StatusCode)
		if err1 != nil {
			return nil, err1
		}
		return nil, errors.New(res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(string(body))
	var resp Txtv1Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.ErrorCode > 0 {
		err1 := getError(res.StatusCode)
		if err1 != nil {
			return nil, err1
		}
		return nil, errors.New(resp.ErrorMsg)
	}

	return &resp, nil
}

type GetImgv1Response struct {
	Data struct {
		Style   string `json:"style,omitempty"`
		TaskId  any    `json:"taskId,omitempty"`
		ImgUrls []struct {
			Image string `json:"image,omitempty"`
		} `json:"imgUrls,omitempty"`
		Text       string `json:"text,omitempty"`
		Status     int    `json:"status,omitempty"`
		CreateTime string `json:"createTime,omitempty"`
		Img        string `json:"img,omitempty"`
		Waiting    string `json:"waiting,omitempty"`
	} `json:"data,omitempty"`
	LogId uint64 `json:"log_id"`
}

// GetImgv2 查询结果
func (b *BaiduAiClient) GetImg(ctx context.Context, taskId string) (*GetImgv1Response, error) {
	token, _ := b.getAccessToken()
	url := fmt.Sprintf("https://aip.baidubce.com/rpc/2.0/ernievilg/v1/getImg?access_token=%s", token)
	fmt.Println(url)
	payload := strings.NewReader("{\"taskId\":\"" + taskId + "\"}")
	fmt.Println("{\"task_id\":\"" + taskId + "\"}")
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(string(body))
	fmt.Println(res.StatusCode)
	if res.StatusCode != http.StatusOK {
		err1 := getError(res.StatusCode)
		if err1 != nil {
			return nil, err1
		}
		return nil, errors.New(res.Status)
	}
	var resp GetImgv1Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &resp, nil
}

// 响应示例
// {
// "data": {
// "task_status": "SUCCESS",
// "task_progress": 1,
// "sub_task_result_list": [
// {
// "final_image_list": [
// {
// "width": 1024,
// "img_url": "http://aigc-t2p.bj.bcebos.com/artist-long/112389890_0_final.png?authorization=bce-auth-v1%2F174bf5e9a7a84f55a8e85b1cc5d62b1d%2F2023-10-12T08%3A23%3A01Z%2F3600%2Fhost%2Fcd3c9216a25d7cf849cc696c5f879d8aabd93a054a9b1a6de455a2d7764ce50e",
// "img_approve_conclusion": "pass",
// "height": 1024
// }
// ],
// "sub_task_error_code": 0,
// "sub_task_status": "SUCCESS",
// "sub_task_progress": 1
// },
// {
// "final_image_list": [
// {
// "width": 1024,
// "img_url": "http://aigc-t2p.bj.bcebos.com/artist-long/112389891_0_final.png?authorization=bce-auth-v1%2F174bf5e9a7a84f55a8e85b1cc5d62b1d%2F2023-10-12T08%3A23%3A01Z%2F3600%2Fhost%2F75d9cfb5ac315abda54715bdf1a6119989b5ececd6b0fc1c1d2737e26563b0f0",
// "img_approve_conclusion": "pass",
// "height": 1024
// }
// ],
// "sub_task_error_code": 0,
// "sub_task_status": "SUCCESS",
// "sub_task_progress": 1
// }
// ],
// "task_id": 0
// },
// "log_id": "1712383433475585303"
// }
type GetImgResponse struct {
	Data struct {
		TaskStatus        string `json:"task_status"`   //计算总状态。有 INIT（初始化），WAIT（排队中）, RUNNING（生成中）, FAILED（失败）, SUCCESS（成功）四种状态，只有 SUCCESS 为成功状态
		TaskProgress      int    `json:"task_progress"` //图片生成总进度，进度包含2种，0为未处理完，1为处理完成
		SubTaskResultList []struct {
			FinalImageList []struct {
				Width                int    `json:"width"`
				ImgUrl               string `json:"img_url"`                // 图片所在 BOS http 地址，默认 1 小时失效
				ImgApproveConclusion string `json:"img_approve_conclusion"` //图片机审结果，"block"：输出图片违规；"review": 输出图片疑似违规；"pass": 输出图片未发现问题；
				Height               int    `json:"height"`
			} `json:"final_image_list"`
			SubTaskErrorCode int    `json:"sub_task_error_code"`
			SubTaskStatus    string `json:"sub_task_status"`   //单风格图片状态。有 INIT（初始化），WAIT（排队中）, RUNNING（生成中）, FAILED（失败）, SUCCESS（成功）四种状态，只有 SUCCESS 为成功状态
			SubTaskProgress  int    `json:"sub_task_progress"` //单任务图片生成进度，进度包含2种，0为未处理完，1为处理完成
		} `json:"sub_task_result_list"`
		TaskId int `json:"task_id"`
	} `json:"data"`
	LogId uint64 `json:"log_id"` //请求唯一标识码
}

// GetImgv2 查询结果
func (b *BaiduAiClient) GetImgv2(ctx context.Context, taskId string) (*GetImgResponse, error) {
	token, _ := b.getAccessToken()
	url := fmt.Sprintf("https://aip.baidubce.com/rpc/2.0/ernievilg/v1/getImgv2?access_token=%s", token)
	payload := strings.NewReader("{\"task_id\":\"" + taskId + "\"}")
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	if res.StatusCode != http.StatusOK {
		err1 := getError(res.StatusCode)
		if err1 != nil {
			return nil, err1
		}
		return nil, errors.New(res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var resp GetImgResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &resp, nil
}

func getError(code int) (err error) {
	switch code {
	case 282000:
		err = errors.New("internal error:服务器内部错误，请再次请求， 如果持续出现此类错误，请在控制台提交工单联系技术支持团队")
	case 282004:
		err = errors.New("文本黄反拦截/图片黄反拦截:请求中包含敏感词、非法参数、字数超限，或上传违规参考图，请检查后重新尝试")
	case 282003:
		err = errors.New("缺少必要参数")
	case 17:
		err = errors.New("日配额流量超限")
	case 18:
		err = errors.New("QPS 超限额")
	case 216630:
		err = errors.New("服务器内部错误，请再次请求，如果持续出现此类错误，请通过工单联系技术支持")
	case 201:
		err = errors.New("模型生图失败截")
	case 216100:
		err = errors.New("参数不满足格式要求")
	case 216201:
		err = errors.New("参考图不满足格式要求")
	case 4:
		err = errors.New("请求超限: 作画总数超限制")
	case 13:
		err = errors.New("QPS 超限")
	case 15:
		err = errors.New("并发超限")
	}

	return err
}
