package chatglm_client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type JWT struct {
	header    string
	Payload   string
	signature string
}

func encodeBase64(data string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(data))
}
func generateSignature(key []byte, data []byte) (string, error) {
	// 创建一个哈希对象
	hash := hmac.New(sha256.New, key)
	// 将要签名的信息写入哈希对象中 hash.Write(data)
	_, err := hash.Write(data)
	if err != nil {
		return "", err
	}
	// hash.Sum()计算签名，在这里会返回签名内容
	// 将签名经过base64编码生成字符串形式返回。
	return encodeBase64(string(hash.Sum(nil))), nil
}
func CreateToken(key []byte, payloadData any) (string, error) {
	// 标准头部
	header := `{"alg":"HS256","sign_type":"SIGN"}`
	// 将负载的数据转换为json
	payload, jsonErr := json.Marshal(payloadData)
	if jsonErr != nil {
		return "", fmt.Errorf("负载json解析错误")
	}

	// 将头部和负载通过base64编码，并使用.作为分隔进行连接
	encodedHeader := encodeBase64(header)
	encodedPayload := encodeBase64(string(payload))
	HeaderAndPayload := encodedHeader + "." + encodedPayload

	// 使用签名使用的key将传入的头部和负载连接所得的数据进行签名
	signature, err := generateSignature(key, []byte(HeaderAndPayload))
	if err != nil {
		return "", err
	}

	// 将token的三个部分使用.进行连接并返回
	return HeaderAndPayload + "." + signature, nil
}
func ParseJwt(token string, key []byte) (*JWT, error) {
	// 分解规定，我们使用.进行分隔，所以我们通过.进行分隔成三个字符串的数组
	jwtParts := strings.Split(token, ".")
	// 数据数组长度不是3就说明token在格式上就不合法
	if len(jwtParts) != 3 {
		return nil, fmt.Errorf("非法token")
	}

	// 分别拿出
	encodedHeader := jwtParts[0]
	encodedPayload := jwtParts[1]
	signature := jwtParts[2]

	// 使用key将token中的头部和负载用.连接后进行签名
	// 这个签名应该个token中第三部分的签名一致
	confirmSignature, err := generateSignature(key, []byte(encodedHeader+"."+encodedPayload))
	if err != nil {
		return nil, fmt.Errorf("生成签名错误")
	}
	// 如果不一致
	if signature != confirmSignature {
		return nil, fmt.Errorf("token验证失败")
	}

	// 将payload解base64编码
	dstPayload, _ := base64.RawURLEncoding.DecodeString(encodedPayload)
	// 返回我们的JWT对象以供后续使用
	return &JWT{encodedHeader, string(dstPayload), signature}, nil
}
