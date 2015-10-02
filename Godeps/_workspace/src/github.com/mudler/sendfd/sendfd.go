// Copyright (c) 2011, Moritz Bitsch <mortizbitsch@googlemail.com>
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// The package sendfd implements sending and receiving Filedescriptors over
// unix domain sockets
package sendfd

import (
	"net"
	"os"
	"syscall"
	"unsafe"
)

// Send File over UnixConn
func SendFD(conn *net.UnixConn, file *os.File) (err error) {
	cmsgb := make([]byte, CmsgSpace(unsafe.Sizeof(int(0))))

	cms := (*syscall.Cmsghdr)(unsafe.Pointer(&cmsgb[0]))
	cms.Len = CmsgLen(unsafe.Sizeof(int(0)))
	cms.Level = 1
	cms.Type = 1

	fdnum := file.Fd()
	fdArea := cmsgb[unsafe.Sizeof(syscall.Cmsghdr{}):]
	fdArea[0] = byte(fdnum)
	fdArea[1] = byte(fdnum >> 8)
	fdArea[2] = byte(fdnum >> 16)
	fdArea[3] = byte(fdnum >> 24)

	_, _, err = conn.WriteMsgUnix([]byte{}, cmsgb, nil)
	return
}

// Receive File from UnixConn
func RecvFD(conn *net.UnixConn) (file *os.File, err error) {
	cmsgb := make([]byte, CmsgSpace(unsafe.Sizeof(int(0))))

	cms := (*syscall.Cmsghdr)(unsafe.Pointer(&cmsgb[0]))
	cms.Len = CmsgLen(unsafe.Sizeof(int(0)))
	cms.Level = 1
	cms.Type = 1

	_, _, _, _, err = conn.ReadMsgUnix([]byte{}, cmsgb)
	if err != nil {
		return
	}

	fdArea := cmsgb[unsafe.Sizeof(syscall.Cmsghdr{}):]
	fdnum := uintptr(fdArea[0]) | uintptr(byte(fdArea[1]<<8)) | uintptr(byte(fdArea[2]<<16)) | uintptr(byte(fdArea[3]<<24))

	file = os.NewFile(fdnum, "")

	return
}
