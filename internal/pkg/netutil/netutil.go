package netutil

import (
    "net"
	"github.com/sirupsen/logrus"
)

// GetOutboundIP gets the preferred outbound ip of this machine.
func GetOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        logrus.Fatal("GetOutboundIP", err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}
