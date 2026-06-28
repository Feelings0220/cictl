package jenkins

import "context"

type Me struct {
	ID            string `json:"id"`
	FullName      string `json:"fullName"`
	Authenticated bool   `json:"authenticated"`
}

func (c *Client) WhoAmI(ctx context.Context) (Me, error) {
	var m Me
	if _, err := c.GET(ctx, "/me/api/json", &m); err != nil {
		return Me{}, err
	}
	return m, nil
}
