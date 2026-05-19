package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const MaxSize = 65507
const TimeFormat = "2006-01-02 03:04:05"

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Printf("Usage: %s <local port> <remote address> [remote port]\n", os.Args[0])
		os.Exit(1)
	}

	localPort := flag.Arg(0)
	remoteAddress := flag.Arg(1)
	remotePort := flag.Arg(2)

	if remotePort == "" {
		remotePort = localPort
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+localPort)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}

	to, err := net.ResolveUDPAddr("udp", remoteAddress+":"+remotePort)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		var buf [MaxSize]byte
		rlen, from, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			fmt.Printf("error: %+v\n", err)
			continue
		}

		msg := buf[:rlen]
		fmt.Printf("%s\tReceive\t%5d bytes from %s\n", time.Now().Format(TimeFormat), rlen, from.String())
		if err != nil {
			fmt.Printf("error: %+v\n", err)
			continue
		}

		remote, err := net.DialUDP("udp", nil, to)
		if err != nil {
			fmt.Printf("error: %+v\n", err)
			continue
		}

		n, err := remote.Write(msg)
		if err != nil {
			fmt.Printf("error: %+v\n", err)
			continue
		}

		fmt.Printf("%s\tSend\t%5d bytes to   %s (packet from %s)\n", time.Now().Format(TimeFormat), n, to.String(), from.String())
	}

}
