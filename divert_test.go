package windivert

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

func TestParam(t *testing.T) {
	const (
		// FIXME document says default WINDIVERT_PARAM_QUEUE_LEN is 512
		DefaultQueueLen  = 1024
		DefaultQueueTime = 512
		// TODO document says this is supported in 1.3.0, but it isn't
		DefaultQueueSize = 4 * 1024 * 1024
	)
	handle, err := Open("true", LayerNetwork, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()
	assertParam := func(name string, param uintptr, value uint64) {
		v, err := handle.GetParam(param)
		if err != nil {
			t.Errorf("fail to get param %v, err: %v", name, err)
			return
		}
		if v != value {
			t.Errorf("expect %v to be %v, but got %v", name, value, v)
		}
	}
	setParam := func(name string, param uintptr, value uint64) {
		err = handle.SetParam(param, value)
		if err != nil {
			t.Errorf("fail to set param %v, err: %v", name, err)
		}
	}
	assertParam("QueueLen", ParamQueueLen, DefaultQueueLen)
	setParam("QueueLen", ParamQueueLen, DefaultQueueLen * 2)
	assertParam("QueueLen", ParamQueueLen, DefaultQueueLen * 2)

	assertParam("QueueTime", ParamQueueTime, DefaultQueueTime)
	setParam("QueueTime", ParamQueueTime, DefaultQueueTime * 2)
	assertParam("QueueTime", ParamQueueTime, DefaultQueueTime * 2)
}

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
