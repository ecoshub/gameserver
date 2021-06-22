package game

func getIP(addr string) string {
	separator := 0
	for i := range addr {
		if addr[i] == ':' {
			separator = i
		}
	}
	return addr[:separator]
}
