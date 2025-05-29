// Ref: https://apidocs.online-go.com/
package googs

func (c *Client) Me() (any, error) {
	res := Me{}
	if err := c.Get("/me", &res); err != nil {
		return nil, err
	}
	return res, nil
}
