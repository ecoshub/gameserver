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
)
