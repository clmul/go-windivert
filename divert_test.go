package divert

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

const (
	n      = 19
	length = 1234
)

var sendMsg []byte

func init() {
	sendMsg = make([]byte, n*length)
	_, err := rand.Read(sendMsg)
	if err != nil {
		log.Fatal(err)
	}
}

func sendUDP() {
	conn, err := net.Dial("udp", "127.0.0.8:0")
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < n; i++ {
		_, err = conn.Write(sendMsg[i*length : (i+1)*length])
		if err != nil {
			log.Fatal(err)
		}
	}
}

func payload(packet []byte) []byte {
	return packet[20+8:]
}

func timeout(d time.Duration) {
	<-time.After(d)
	log.Fatal("timeout")
}

func TestRecv(t *testing.T) {
	t.Log("Hello, Divert")
	handle, err := Open(fmt.Sprintf("outbound and ip.DstAddr = 127.0.0.8 and udp.PayloadLength = %v", length), LayerNetwork, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	go sendUDP()
	go timeout(time.Second * 5)
	packet := make([]byte, 2048)
	var recvMsg []byte
	for i := 0; i < n; i++ {
		n, _, err := handle.Recv(packet)
		if err != nil {
			t.Fatal(err)
		}
		recvMsg = append(recvMsg, payload(packet[:n])...)
	}
	if !bytes.Equal(recvMsg, sendMsg) {
		t.Log(sendMsg[:20], recvMsg[:20])
		t.Error("wrong message")
	}
	err = handle.Close()
	if err != nil {
		t.Fatal(err)
	}
}
