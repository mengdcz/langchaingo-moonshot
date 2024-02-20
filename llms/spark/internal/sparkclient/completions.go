package sparkclient

import "context"

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
type Completion struct {
	Text  string `json:"text"`
	Usage Usage  `json:"usage"`
}
type CompletionResponse struct {
}

func (c *Client) createCompletion(ctx context.Context, payload *ChatRequestUser) (*ChatResponse, error) {
	return c.createChat(ctx, payload)

}
