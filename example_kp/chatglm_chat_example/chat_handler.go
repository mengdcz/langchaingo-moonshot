//nolint:forbidigo
package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/callbacks"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LogHandler is a callback handler that prints to the standard output.
type ChatHandler struct {
	DetailModel string
}

var _ callbacks.Handler = ChatHandler{}

func (l ChatHandler) HandleText(_ context.Context, text string) {
	fmt.Println(text)
}

func (l ChatHandler) HandleLLMStart(_ context.Context, prompts []string) {
	fmt.Println("Entering LLM with prompts:", prompts)
}

func (l ChatHandler) HandleLLMEnd(_ context.Context, output llms.LLMResult) {
	fmt.Println("Exiting LLM with results:", formatLLMResult(output))
}

func (l ChatHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	fmt.Println("Entering chain with inputs:", formatChainValues(inputs))
}

func (l ChatHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	fmt.Println("Exiting chain with outputs:", formatChainValues(outputs))
}

func (l ChatHandler) HandleToolStart(_ context.Context, input string) {
	fmt.Println("Entering tool with input:", removeNewLines(input))
}

func (l ChatHandler) HandleToolEnd(_ context.Context, output string) {
	fmt.Println("Exiting tool with output:", removeNewLines(output))
}

func (l ChatHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	fmt.Println("Agent selected action:", formatAgentAction(action))
}

func (l ChatHandler) HandleRetrieverStart(_ context.Context, query string) {
	fmt.Println("Entering retriever with query:", removeNewLines(query))
}

func (l ChatHandler) HandleRetrieverEnd(_ context.Context, documents []schema.Document) {
	fmt.Println("Exiting retirer with documents:", documents)
}

func formatChainValues(values map[string]any) string {
	output := ""
	for key, value := range values {
		output += fmt.Sprintf("\"%s\" : \"%s\", ", removeNewLines(key), removeNewLines(value))
	}

	return output
}

func formatLLMResult(output llms.LLMResult) string {
	results := "[ "
	for i := 0; i < len(output.Generations); i++ {
		for j := 0; j < len(output.Generations[i]); j++ {
			results += output.Generations[i][j].Text
			//fmt.Println(output.Generations[i][j].GenerationInfo)
			//results += fmt.Sprintf("%v", output.Generations[i][j].GenerationInfo)
		}
	}

	return results + " ]"
}

func formatAgentAction(action schema.AgentAction) string {
	return fmt.Sprintf("\"%s\" with input \"%s\"", removeNewLines(action.Tool), removeNewLines(action.ToolInput))
}

func removeNewLines(s any) string {
	return strings.ReplaceAll(fmt.Sprint(s), "\n", " ")
}
