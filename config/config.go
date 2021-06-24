package config

// maybe Those config values can be set with cmd arguments
// or a configuration file
const (
	// game size
	GameSize             int    = 2
	ServerID             uint16 = 0
	ServerListenAddress  string = "0.0.0.0"
	ClientRequestAddress string = "localhost"
	TCPPort              string = "8080"
	UDPPort              string = "9090"

	// if all client are simulating in the same machine
	// ip and ports will be the same
	// this flag indicates
	// client simulation must change its udp listen port to avoid port collision
	Simulation bool = true

	MinGameOverTime int   = 10000
	MaxGameOverTime int   = 15000
	NullData        int32 = 0
)
