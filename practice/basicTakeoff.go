package main

import "net"

func main() {
	// // udp通信でdroneへCommandを送る。
	conn, _ := net.Dial("udp", "192.168.10.1:8889")
	conn.Write([]byte("command"))
	conn.Write([]byte("takeoff"))

	// conn.Write([]byte{0x00cc, 0x0058, 0x0000, 0x007c, 0x0068, 0x0054, 0x0000, 0x00e4, 0x0001, 0x00c2, 0x0016})
}
