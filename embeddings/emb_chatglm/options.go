package emb_chatglm

import "github.com/tmc/langchaingo/llms/chatglm"

const (
	defaultBatchSize     = 256
	defaultStripNewLines = true
)

type Option func(p *Chatglm)

func WithClient(client chatglm.LLM) Option {
	return func(c *Chatglm) {
		c.client = &client
	}
}

func WithBatchSize(batchSize int) Option {
	return func(c *Chatglm) {
		c.batchSize = batchSize
	}
}

func WithStripNewLines(stripNewLines bool) Option {
	return func(c *Chatglm) {
		c.stripNewLines = stripNewLines
	}
}
