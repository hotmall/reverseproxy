package reverseproxy

import "github.com/valyala/bytebufferpool"

type bytesPool struct {
}

func (b *bytesPool) Get() []byte {
	bf := bytebufferpool.Get()
	return bf.Bytes()
}

func (b *bytesPool) Put(bytes []byte) {
	bf := bytebufferpool.ByteBuffer{
		B: bytes,
	}
	bf.Reset()
	bytebufferpool.Put(&bf)
}
