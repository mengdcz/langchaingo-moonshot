package emb_spark

import (
	"context"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/spark"
)

type Spark struct {
	client        *spark.LLM
	batchSize     int
	stripNewLines bool
}

var _ embeddings.Embedder = &Spark{}

func NewSpark(opts ...Option) (*Spark, error) {
	v := &Spark{
		stripNewLines: defaultStripNewLines,
		batchSize:     defaultBatchSize,
	}

	for _, opt := range opts {
		opt(v)
	}

	if v.client == nil {
		client, err := spark.New()
		if err != nil {
			return nil, err
		}
		v.client = client
	}
	return v, nil
}

// EmbedDocuments use ernie Embedding-V1.
func (e *Spark) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.stripNewLines),
		e.batchSize,
	)

	emb := make([][]float64, 0, len(texts))
	for _, texts := range batchedTexts {
		curTextEmbeddings, err := e.client.CreateEmbedding(ctx, texts)
		if err != nil {
			return nil, err
		}

		textLengths := make([]int, 0, len(texts))
		for _, text := range texts {
			textLengths = append(textLengths, len(text))
		}

		combined, err := embeddings.CombineVectors(curTextEmbeddings, textLengths)
		if err != nil {
			return nil, err
		}

		emb = append(emb, combined)
	}

	return emb, nil
}

// EmbedQuery use ernie Embedding-V1.
func (e *Spark) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	emb, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}
