package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/blackspace/gofb/framebuffer"
)

const (
	pxOffset = 2
)

var (
	bindAddr string
	port     int
	xRes     int
	yRes     int
)

func main() {
	flag.StringVar(&bindAddr, "addr", "0.0.0.0", "IP address")
	flag.IntVar(&port, "port", 1234, "TCP port")
	flag.Parse()

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindAddr, port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s:%d\n", bindAddr, port)

	fb := framebuffer.NewFramebuffer()
	defer fb.Release()
	fb.Init()

	// hotfix on github.com/blackspace/gofb/framebuffer issue
	// causing index out of range error
	xRes = fb.Xres - pxOffset
	yRes = fb.Yres - pxOffset

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn, fb)
	}
}

func handleRequest(conn net.Conn, fb *framebuffer.Framebuffer) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		split := strings.Split(msg, " ")

		switch split[0] {
		case "SIZE":
			msg := fmt.Sprintf("SIZE %d %d\r\n", xRes, yRes)
			conn.Write([]byte(msg))
		case "PX":
			x, errX := strconv.Atoi(split[1])
			y, errY := strconv.Atoi(split[2])
			if errX != nil || errY != nil {
				break
			}
			if x > xRes || y > yRes {
				break
			}
			r, g, b, a := hexToRGB(split[3])
			fb.SetPixel(x, y, r, g, b, a)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print("error: ", err)
	}
}

func hexToRGB(hex string) (r uint32, g uint32, b uint32, a uint32) {
	a = 255

	switch len(hex) {
	case 3:
		fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
	case 6:
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	case 8:
		fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
	}

	return
}
