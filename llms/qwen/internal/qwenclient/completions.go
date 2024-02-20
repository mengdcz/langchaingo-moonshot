package qwenclient

import "context"

func (c *Client) createCompletion(ctx context.Context, payload *ChatRequestUser) (*ChatResponse, error) {
	return c.createChat(ctx, payload)
}
