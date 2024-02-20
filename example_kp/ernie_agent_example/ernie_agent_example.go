package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/tools/gaokao"
	"os"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	llm, err := openai.New()
	if err != nil {
		return err
	}
	search, err := gaokao.New()
	if err != nil {
		return err
	}
	agentTools := []tools.Tool{
		tools.Calculator{},
		search,
	}

	//	prefixStr := `今天是 {{.today}}，您可以使用工具来获取新信息。
	//使用以下工具尽可能回答以下问题:
	//
	//{{.tool_descriptions}}`
	//	instruStr := `使用以下格式：
	//
	//问题：您必须回答的输入问题
	//想法：你应该时刻思考该做什么
	//动作：要采取的操作，应该是 [ {{.tool_names}} ] 之一
	//动作输入：动作的输入
	//观察：动作的结果
	//...（这个想法/动作/动作输入/观察可以重复N次）
	//想法：我现在知道了最终答案
	//最终答案：原始输入问题的最终答案`
	//	suffixStr := `开始!
	//
	//问题: {{.input}}
	//想法:{{.agent_scratchpad}}`

	//prefix := agents.WithPromptPrefix(prefixStr)
	//instru := agents.WithPromptFormatInstructions(instruStr)
	//suffix := agents.WithPromptSuffix(suffixStr)

	executor, err := agents.Initialize(
		llm,
		agentTools,
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(3),
		//prefix,
		//instru,
		//suffix,
	)
	if err != nil {
		return err
	}
	question := "今天是什么日子"
	answer, err := chains.Run(context.Background(), executor, question)
	fmt.Println(answer)
	return err
}
