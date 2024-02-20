package emb_qwen

import "github.com/tmc/langchaingo/llms/qwen"

const (
	defaultBatchSize     = 2048
	defaultStripNewLines = true
)

type Option func(p *Qwen)

func WithClient(client qwen.LLM) Option {
	return func(c *Qwen) {
		c.client = &client
	}
}

func WithBatchSize(batchSize int) Option {
	return func(c *Qwen) {
		c.batchSize = batchSize
	}
}

func WithStripNewLines(stripNewLines bool) Option {
	return func(c *Qwen) {
		c.stripNewLines = stripNewLines
	}
}
