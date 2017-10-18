package divert

import (
	"io"
	"reflect"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var dll *windows.DLL
var open *windows.Proc
var recv *windows.Proc
var send *windows.Proc
var close_ *windows.Proc
var calcChecksums *windows.Proc

const (
	false_             = 0
	InvalidHandleValue = -1

	LayerNetwork        = 0
	LayerNetworkForward = 1

	FlagSniff = 1
	FlagDrop  = 2

	DirectionOutbound = 0
	DirectionInbound  = 1

)

type Address struct {
	IfIdx     uint32 // Packet's interface index
	SubIfIdx  uint32 // Packet's sub-interface index
	Direction uint8  // Packet's direction
}

type Handle uintptr

func init() {
	dll = windows.MustLoadDLL("WinDivert")
	open = dll.MustFindProc("WinDivertOpen")
	recv = dll.MustFindProc("WinDivertRecv")
	send = dll.MustFindProc("WinDivertSend")
	close_ = dll.MustFindProc("WinDivertClose")
	calcChecksums = dll.MustFindProc("WinDivertHelperCalcChecksums")
}

func Open(filter string, layer, priority, flags int) (Handle, error) {
	r, _, err := open.Call(stringToPtr(filter+"\x00"), uintptr(layer), uintptr(priority), uintptr(flags))
	if int(r) == InvalidHandleValue {
		return 0, err
	}
	return Handle(r), nil
}

func (h Handle) Close() error {
	r, _, err := close_.Call(uintptr(h))
	if r == false_ {
		return err
	}
	return nil
}

func (h Handle) Recv(packet []byte) (n int, addr Address, err error) {
	r, _, err := recv.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Pointer(&n)))
	if r == false_ {
		return 0, addr, err
	}
	return n, addr, nil
}

func (h Handle) Send(packet []byte, addr Address) (n int, err error) {
	r, _, err := send.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Pointer(&n)))
	if r == false_ {
		return 0, err
	}
	if len(packet) != n {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func CalcChecksums(packet []byte) []byte {
	calcChecksums.Call(bytesToPtr(packet), uintptr(len(packet)), 0)
	return packet
}

func bytesToPtr(buffer []byte) uintptr {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&buffer))
	return uintptr(unsafe.Pointer(header.Data))
}

func stringToPtr(s string) uintptr {
	if !strings.HasSuffix(s, "\x00") {
		panic("str argument missing null terminator: " + s)
	}
	header := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return uintptr(unsafe.Pointer(header.Data))
}
