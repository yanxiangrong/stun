package pack

import (
	"encoding/binary"
	"errors"
	"github.com/someonegg/msgpump"
	"io"
	"math"
)

const version = 1
const recvBufferSize = 4096

var errorMessageFormat = errors.New("wrong message format")

type MyPkg struct {
	Version       uint8
	PayloadLength []byte
	Payload       []byte
}

func NewMyPkg(data []byte) *MyPkg {
	var payloadLength []byte

	if len(data) > math.MaxUint16 {
		payloadLength = make([]byte, 9)

		payloadLength[0] = byte(255)
		binary.BigEndian.PutUint64(payloadLength[1:9], uint64(len(data)))
	} else if len(data) > 253 {
		payloadLength = make([]byte, 3)

		payloadLength[0] = byte(254)
		binary.BigEndian.PutUint16(payloadLength[1:3], uint16(len(data)))
	} else {
		payloadLength = make([]byte, 1)
		payloadLength[0] = byte(len(data))
	}

	return &MyPkg{
		Version:       version,
		PayloadLength: payloadLength,
		Payload:       data,
	}
}

func (p *MyPkg) Slice() []byte {
	buf := make([]byte, len(p.Payload)+len(p.PayloadLength)+1)
	buf[0] = p.Version
	copy(buf[1:1+len(p.PayloadLength)], p.PayloadLength)
	copy(buf[1+len(p.PayloadLength):], p.Payload)
	return buf
}

func ParseBytes(b []byte) (*MyPkg, error) {
	v := b[0]
	if v != version {
		return nil, errors.New("illegal version")
	}

	p := 0

	switch b[1] {
	case 255:
		p = 10
	case 254:
		p = 4
	default:
		p = 2
	}

	return &MyPkg{
		Version:       v,
		PayloadLength: b[1:p],
		Payload:       b[p:],
	}, nil
}

type ReadWriter struct {
	rw        io.ReadWriter
	buf       []byte
	bufOffset int
	bufLength int
}

type Message []byte

type MessageReadWriter struct {
	rw ReadWriter
}

func NewReadWriter(rw io.ReadWriter) *ReadWriter {
	return &ReadWriter{
		rw:  rw,
		buf: make([]byte, recvBufferSize),
	}
}

func NewMessageReadWriter(rw io.ReadWriter) *MessageReadWriter {
	return &MessageReadWriter{
		rw: *NewReadWriter(rw),
	}
}

func (r *ReadWriter) Read() ([]byte, error) {
	var data []byte
	length := 0
	offset := 0

	needRead := false
	if r.bufLength == 0 {
		needRead = true
	}

	for {
		if needRead {
			if r.bufOffset > recvBufferSize/2 {
				if r.bufLength != 0 {
					copy(r.buf[:r.bufLength], r.buf[r.bufOffset:r.bufOffset+r.bufLength])
				}
				r.bufOffset = 0
			}

			n, err := r.rw.Read(r.buf[r.bufOffset+r.bufLength:])
			if err != nil {
				return nil, err
			}

			r.bufLength += n
			needRead = false
		}

		if length != 0 {
			if offset == length {
				break
			}

			l := length - offset
			if l > r.bufLength {
				l = r.bufLength
				needRead = true
			}

			copy(data[offset:offset+l], r.buf[r.bufOffset:r.bufOffset+l])
			r.bufOffset += l
			r.bufLength -= l
			offset += l

			continue
		}

		if r.bufLength < 2 {
			needRead = true
			continue
		}

		if r.bufLength < 4 && r.buf[r.bufOffset+1] > 253 {
			needRead = true
			continue
		} else if r.bufLength < 10 && r.buf[r.bufOffset+1] > 254 {
			needRead = true
			continue
		}

		if r.buf[r.bufOffset] != byte(version) {
			r.bufLength = 0
			continue
		}
		switch r.buf[r.bufOffset+1] {
		case 255:
			length = int(binary.BigEndian.Uint64(r.buf[r.bufOffset+2 : r.bufOffset+10]))
			r.bufOffset += 10
			r.bufLength -= 10
		case 254:
			length = int(binary.BigEndian.Uint16(r.buf[r.bufOffset+2 : r.bufOffset+4]))
			r.bufOffset += 4
			r.bufLength -= 4
		default:
			length = int(r.buf[r.bufOffset+1])
			r.bufOffset += 2
			r.bufLength -= 2
		}

		data = make([]byte, length)
	}
	return data, nil
}

func (r *ReadWriter) Write(data []byte) error {
	pkg := NewMyPkg(data)

	_, err := r.rw.Write(pkg.Slice())
	return err
}

func (r *MessageReadWriter) ReadMessage() (string, msgpump.Message, error) {
	buf, err := r.rw.Read()
	if err != nil {
		return "", nil, err
	}

	i := 0
	for ; i < len(buf); i++ {
		if buf[i] == byte('=') {
			break
		}
	}
	if i == len(buf) {
		return "", buf, errorMessageFormat
	}
	return string(buf[:i]), buf[i+1:], err
}

func (r *MessageReadWriter) WriteMessage(t string, m msgpump.Message) error {
	buf := make([]byte, len(t)+len(m)+1)
	p := 0
	copy(buf[p:p+len(t)], t)
	p += len(t)
	copy(buf[p:p+1], "=")
	p += 1
	copy(buf[p:p+len(m)], m)

	return r.rw.Write(buf)
}
