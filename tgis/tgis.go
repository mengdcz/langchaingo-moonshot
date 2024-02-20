package tgis

import (
	"context"
)

// 文生图模型接口定义
type TextGenerationImageModel interface {
	GenerateImage(ctx context.Context, text string) (string, error)
}
