package divert

import (
	"golang.org/x/sys/windows"
	"reflect"
	"strings"
	"unsafe"
)

var dll *windows.DLL
var open *windows.Proc
var recv *windows.Proc
var send *windows.Proc
var closee *windows.Proc

const (
	False              = 0
	InvalidHandleValue = -1

	WinDivertLayerNetwork        = 0
	WinDivertLayerNetworkForward = 1

	WinDirvertFlagSniff = 1
	WinDivertFlagDrop   = 2

	WinDivertDirectionOutbound = 0
	WinDivertDirectionInbound  = 1
)

type Address struct {
	ifIdx     uint32 // Packet's interface index
	subIfIdx  uint32 // Packet's sub-interface index
	direction uint8  // Packet's direction
}

type Handle uintptr

func init() {
	dll = windows.MustLoadDLL("WinDivert")
	open = dll.MustFindProc("WinDivertOpen")
	recv = dll.MustFindProc("WinDivertRecv")
	send = dll.MustFindProc("WinDivertSend")
	closee = dll.MustFindProc("WinDivertClose")
}

func Open(filter string, layer, priority, flags int) (Handle, error) {
	r, _, err := open.Call(stringToPtr(filter+"\x00"), uintptr(layer), uintptr(priority), uintptr(flags))
	if int(r) == InvalidHandleValue {
		return 0, err
	}
	return Handle(r), nil
}

func (h Handle) Close() error {
	r, _, err := closee.Call(uintptr(h))
	if r == False {
		return err
	}
	return nil
}
func (h Handle) Recv(packet []byte) (int, *Address, error) {
	recvLen := int(0)
	addr := &Address{}
	r, _, err := recv.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(addr)), uintptr(unsafe.Pointer(&recvLen)))
	if r == False {
		return 0, nil, err
	}
	return recvLen, addr, nil
}
func (h Handle) SendOut(packet []byte) (int, error) {
	sendLen := int(0)
	addr := &Address{direction: WinDivertDirectionOutbound}
	r, _, err := send.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(addr)), uintptr(unsafe.Pointer(&sendLen)))
	if r == False {
		return 0, err
	}
	return sendLen, nil
}
func (h Handle) SendIn(packet []byte, addr *Address) (int, error) {
	sendLen := int(0)
	addr.direction = WinDivertDirectionInbound
	r, _, err := send.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(addr)), uintptr(unsafe.Pointer(&sendLen)))
	if r == False {
		return 0, err
	}
	return sendLen, nil
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
