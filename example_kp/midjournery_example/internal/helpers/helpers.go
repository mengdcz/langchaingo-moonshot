package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/tgis/thenextleg"
	"html/template"
	"reflect"
	"strconv"
	"time"
)

func GetMessage(c *thenextleg.TheNextLeg, messageId string) (*thenextleg.MessageResponse, error) {
	var resppp *thenextleg.MessageResponse
	var err error
	for i := 0; i < 30; i++ {
		resppp, err = c.Message(context.Background(), messageId)
		if err != nil {
			fmt.Println(err)
			return resppp, err
		}
		a, _ := json.Marshal(resppp)

		fmt.Printf("%#v\n", string(a))
		str, err := ToStringE(resppp.Progress)
		if err != nil {
			fmt.Println(err)
			return resppp, err
		}
		// 完成
		if str == "100" {
			break
		}
		// 失败
		if str == "incomplete" {
			break
		}
		time.Sleep(time.Second * 2)
	}
	return resppp, nil
}

func ToStringE(i any) (string, error) {
	i = indirectToStringerOrError(i)

	switch s := i.(type) {
	case string:
		return s, nil
	case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(s), nil
	case int64:
		return strconv.FormatInt(s, 10), nil
	case int32:
		return strconv.Itoa(int(s)), nil
	case int16:
		return strconv.FormatInt(int64(s), 10), nil
	case int8:
		return strconv.FormatInt(int64(s), 10), nil
	case uint:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(s), 10), nil
	case json.Number:
		return s.String(), nil
	case []byte:
		return string(s), nil
	case template.HTML:
		return string(s), nil
	case template.URL:
		return string(s), nil
	case template.JS:
		return string(s), nil
	case template.CSS:
		return string(s), nil
	case template.HTMLAttr:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", i, i)
	}
}

var (
	errorType       = reflect.TypeOf((*error)(nil)).Elem()
	fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
)

func indirectToStringerOrError(a any) any {
	if a == nil {
		return nil
	}
	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
