package emb_spark

import "github.com/tmc/langchaingo/llms/spark"

const (
	defaultBatchSize     = 256
	defaultStripNewLines = true
)

type Option func(p *Spark)

func WithClient(client spark.LLM) Option {
	return func(c *Spark) {
		c.client = &client
	}
}

func WithBatchSize(batchSize int) Option {
	return func(c *Spark) {
		c.batchSize = batchSize
	}
}

func WithStripNewLines(stripNewLines bool) Option {
	return func(c *Spark) {
		c.stripNewLines = stripNewLines
	}
}
