package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/schollz/progressbar/v3"
	"log"
	"math/rand"
	"net"
	"os"
	"stun/package/myip"
	"stun/package/utils"
	"time"
)

var token = flag.String("t", "20232023", "Token")
var listenPort = flag.Int("p", 20232, "Listen port")
var dstIpaddr = flag.String("ip", "", "Target IP address")
var delayNum = flag.Int("i", -1, "Scan interval")
var debug = flag.Bool("debug", false, "")

const alpn = "stun"

type Handshake struct {
	Token string `json:"token"`
	Code  int    `json:"code"`
}

func createConn() (*net.UDPConn, error) {
	lUdpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", lPort))
	listenConn, err := net.ListenUDP("udp4", lUdpAddr)
	if err != nil {
		return nil, err
	}
	return listenConn, nil
}

func scan(listenConn *net.UDPConn, rIp net.IP) (*net.UDPAddr, int) {
	defer func(listenConn *net.UDPConn) {
		err := listenConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(listenConn)

	var rAddr *net.UDPAddr
	recvD := make(chan struct{})
	recvCode := 0

	go func() {
		for {
			buf := make([]byte, 4096)
			var n int
			n, rAddr, err = listenConn.ReadFromUDP(buf)
			if err != nil {
				log.Fatalln(err)
			}

			var handshake Handshake
			err = json.Unmarshal(buf[:n], &handshake)

			if err != nil || handshake.Token != *token {
				continue
			}

			recvCode = handshake.Code
			close(recvD)
			break
		}
	}()

	send := func(port int, code int) {
		addr := net.UDPAddr{
			Port: port,
			IP:   rIp,
		}

		b, err := json.Marshal(Handshake{
			Token: *token,
			Code:  code,
		})
		if err != nil {
			log.Panicln(err)
		}

		_, err = listenConn.WriteToUDP(b, &addr)
		if err != nil {
			log.Fatalln(err)
		}
	}

	exit := false
	for j := 1; j <= 4096 && !exit; j++ {
		fmt.Printf("第 %d 轮探测中... (listen on %s)\n", j, lUdpAddr)

		bar := progressbar.NewOptions(65535,
			progressbar.OptionSetWidth(25),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowCount(),
			progressbar.OptionThrottle(100*time.Millisecond),
		)
		for i := 1; i <= 65535 && !exit; i++ {
			send(i, 0)

			_ = bar.Add(1)
			//fmt.Printf("%d%%..", i*100/65535)
			time.Sleep(time.Duration(500+*delayNum) * time.Microsecond)

			select {
			case <-recvD:
				_ = bar.Close()
				exit = true
			default:
			}
		}

		//time.Sleep(time.Microsecond)
		time.Sleep(time.Duration(500+*delayNum) * time.Microsecond)
	}

	select {
	case <-recvD:
	default:
		fmt.Println("探测失败.")
		os.Exit(0)
	}

	fmt.Printf("探测成功. (remote %s)\n", rAddr.String())

	rIp = rAddr.IP
	if recvCode == 0 {
		for i := 0; i < 5; i++ {
			send(rAddr.Port, 1)
			time.Sleep(time.Millisecond)
		}
	}

	return rAddr, recvCode
}

func communicat(lPort int, rAddr *net.UDPAddr) {
	log.Println("建立连接...")
	host, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}

	lUdpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", lPort))
	listenConn, err := net.ListenUDP("udp4", lUdpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer func(listenConn *net.UDPConn) {
		err = listenConn.Close()
		if err != nil {
			log.Println(err)
		}
	}(listenConn)

	go func() {
		for {
			buf := make([]byte, 4096)
			var n int
			n, _, err = listenConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			log.Println("Recv:", buf[:n])
		}
	}()

	for {
		send := "test-" + host
		log.Println("Send:", send)
		_, err = listenConn.WriteToUDP([]byte(send), rAddr)
		time.Sleep(2 * time.Second)
	}
}

func main() {
	log.Println("Started")

	flag.Parse()

	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	rand.Seed(time.Now().UnixNano())
	if *delayNum < 0 {
		*delayNum = rand.Intn(4000)
		if *debug {
			log.Println("delayNum:", *delayNum)
		}
	}

	ip, err := myip.GetMyIp()
	if err != nil {
		log.Println(err)
	}
	log.Println("你的IP地址:", ip)

	rIpStr := *dstIpaddr
	if rIpStr == "" {
		fmt.Print("请输入对方IP地址: ")
		_, err := fmt.Scanln(&rIpStr)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	//lPort := generatePort()
	lPort := *listenPort
	conn, err := createConn()
	if err != nil {
		log.Fatalln(conn)
	}

	rAddr, recvCode := scan(conn, net.ParseIP(rIpStr))

	communicat(lPort, rAddr)

	switch recvCode {
	case 0:
		server(conn)
	case 1:
		client(conn, rAddr)
	}
}

func server(conn *net.UDPConn) {
	listener, err := quic.Listen(conn, utils.GenerateTLSConfig(alpn), nil)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		go echo(conn)
		//todo
	}
}

func client(conn *net.UDPConn, remote net.Addr) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{alpn},
	}
	quicConn, err := quic.Dial(conn, remote, "", tlsConf, nil)
	if err != nil {
		log.Fatalln()
	}
}
