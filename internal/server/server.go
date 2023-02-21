package server

import (
	"context"
	"github.com/quic-go/quic-go"
	"log"
	"net"
	"stun/internal/utils"
	"stun/package/pack"
)

func server(conn *net.UDPConn, alpn string) {
	listener, err := quic.Listen(conn, utils.GenerateTLSConfig(alpn), nil)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		quicConn, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		stream, err := quicConn.AcceptStream(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		rw := pack.NewReadWriter(stream)
		//todo
	}
}
