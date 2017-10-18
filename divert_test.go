package divert

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

func TestRecv(t *testing.T) {
	const (
		n      = 19
		length = 1234
	)
	msgs := make([]byte, n*length)
	_, err := rand.Read(msgs)
	if err != nil {
		log.Fatal(err)
	}

	handle, err := Open(fmt.Sprintf("outbound and ip.DstAddr = 127.0.0.8 and udp.PayloadLength = %v", length), LayerNetwork, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = handle.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	go func() {
		conn, err := net.Dial("udp", "127.0.0.8:0")
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < n; i++ {
			_, err = conn.Write(msgs[i*length : (i+1)*length])
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	go timeout(ctx, time.Second*5)
	defer cancel()

	packet := make([]byte, 2048)
	var recvMsg []byte
	for i := 0; i < n; i++ {
		n, _, err := handle.Recv(packet)
		if err != nil {
			t.Fatal(err)
		}
		recvMsg = append(recvMsg, udpPayload(packet[:n])...)
	}
	if !bytes.Equal(recvMsg, msgs) {
		t.Log(msgs[:20], recvMsg[:20])
		t.Error("wrong message")
	}
}

func udpPayload(packet []byte) []byte {
	return packet[20+8:]
}

func timeout(ctx context.Context, d time.Duration) {
	<-time.After(d)
	select {
	case <-ctx.Done():
	default:
		log.Fatal("timeout")
	}
}
