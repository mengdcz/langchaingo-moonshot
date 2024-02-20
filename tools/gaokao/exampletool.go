package gaokao

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"regexp"
	"strconv"
)

type Tool struct {
	CallbacksHandler callbacks.Handler
}

var _ tools.Tool = Tool{}

func New() (*Tool, error) {
	return &Tool{
		CallbacksHandler: callbacks.LogHandler{},
	}, nil
}

func (t Tool) Name() string {
	return "中国大学查询"
}
func (t Tool) Description() string {
	return `
	"这是一个查询大学的工具"
	"当你需要获取大学信息是很有用"
	"当你需要在互联网上查询大学信息时，这是你的首选"
	"输入应该是高考分数"`
}
func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	fmt.Println("aaaaa", input)
	re := regexp.MustCompile(`\d+`)
	score := re.FindString(input)
	if score == "" {
		return "", nil
	}
	var scoreNu int
	var err error
	if scoreNu, err = strconv.Atoi(score); err != nil {
		return "", nil
	}
	var result string = ""
	if scoreNu > 560 {
		result += "\n\n" + `
	学校名称：江苏财经大学
	专业： 计算机科学与技术
	2023年最低录取分数： 560
	学校描述： 985、211工程
	专业描述： 本校最强学科是计算机科学与技术"`
	}
	if scoreNu > 540 && scoreNu < 560 {
		result += "\n\n" + `"学校名称：湖北财经大学
	专业： 法学
	2023年最低录取分数： 540
	学校描述： 在长江边上
	专业描述： 可调剂"`
	}
	if scoreNu > 520 && scoreNu < 540 {
		result += "\n\n" + `"学校名称：济南财经大学
	专业： 汉语
	2023年最低录取分数： 520
	学校描述： 校园优美，生活方便
	专业描述： 汉语专业是本校不重视的专业"`
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}
	return result, nil
}
