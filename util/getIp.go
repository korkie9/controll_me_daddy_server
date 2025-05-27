package util

import "net"

func GetLocalIP() (string, error) {
	// Connect to an unreachable IP just to get the local IP used for outbound connection
	conn, err := net.Dial("udp", "192.168.0.1:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
