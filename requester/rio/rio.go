package rio

import (
	"io"
)

// Lener 返回32-bit长度接口
type Lener interface {
	Len() int
}

// Lener64 返回64-bit长度接口
type Lener64 interface {
	Len() int64
}

// ReaderLen 实现io.Reader和32-bit长度接口
type ReaderLen interface {
	io.Reader
	Lener
}

// ReaderLen64 实现io.Reader和64-bit长度接口
type ReaderLen64 interface {
	io.Reader
	Lener64
}

// WriterLen64 实现io.Writer和64-bit长度接口
type WriterLen64 interface {
	io.Writer
	Lener64
}

type WriteCloserAt interface {
	io.WriteCloser
	io.WriterAt
}

type WriteCloserLen64At interface {
	WriteCloserAt
	Lener64
}
