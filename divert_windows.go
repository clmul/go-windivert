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
var closee *windows.Proc
var calcChecksums *windows.Proc

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

type address struct {
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
	r, _, err := closee.Call(uintptr(h))
	if r == False {
		return err
	}
	return nil
}
func (h Handle) Recv(packet []byte) (n int, ifidx uint32, direction uint8, err error) {
	var addr address
	r, _, err := recv.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Pointer(&n)))
	if r == False {
		return 0, 0, 0, err
	}
	return n, addr.ifIdx, addr.direction, nil
}
func (h Handle) SendOut(packet []byte) (int, error) {
	var n int
	var addr address
	addr.direction = WinDivertDirectionOutbound
	r, _, err := send.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Pointer(&n)))
	if r == False {
		return 0, err
	}
	if len(packet) != n {
		return n, io.ErrShortWrite
	}
	return n, nil
}
func (h Handle) SendIn(packet []byte, ifidx uint32) (int, error) {
	var n int
	var addr address
	addr.ifIdx = ifidx
	addr.direction = WinDivertDirectionInbound
	r, _, err := send.Call(uintptr(h), bytesToPtr(packet), uintptr(len(packet)),
		uintptr(unsafe.Pointer(&addr)), uintptr(unsafe.Pointer(&n)))
	if r == False {
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
