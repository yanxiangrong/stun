package service

import (
	"fmt"
	"net"
)

func CreateConn(lPort int) (*net.UDPConn, error) {
	lUdpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", lPort))
	listenConn, err := net.ListenUDP("udp4", lUdpAddr)
	if err != nil {
		return nil, err
	}
	return listenConn, nil
}
