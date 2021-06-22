package client

import "net"

var (
	// possible client states
	ClientState = struct {
		InQueue string
		InPool  string
		InGame  string
	}{
		InQueue: "in_queue",
		InPool:  "in_pool",
		InGame:  "in_game",
	}
)

type Client struct {
	ClientID      uint16
	TCPconn       net.Conn
	Addr          string
	State         string
	UDPRegistered bool
}

func NewClient(clientID uint16, conn net.Conn) *Client {
	return &Client{
		ClientID: clientID,
		TCPconn:  conn,
		State:    ClientState.InQueue,
	}
}

func (c *Client) ChangeState(state string) {
	c.State = state
}

func (c *Client) IsRegistered() bool {
	return c.UDPRegistered
}

func (c *Client) Register() {
	c.UDPRegistered = true
}
