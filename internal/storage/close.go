package storage

func (client *Client) Close() {
	client.env.Close()
}
