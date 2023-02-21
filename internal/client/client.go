package client

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"log"
	"net"
	"stun/package/pack"
)

func client(conn *net.UDPConn, remote net.Addr, alpn string) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{alpn},
	}
	quicConn, err := quic.Dial(conn, remote, "", tlsConf, nil)
	if err != nil {
		log.Fatalln(err)
	}

	stream, err := quicConn.OpenStreamSync(context.Background())
	if err != nil {
		log.Println(err)
	}

	rw := pack.NewReadWriter(stream)

	//todo

}
