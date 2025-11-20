package dwc

import (
	"io"
	"sync"
)

type syncedWriter struct {
	protectedDst io.Writer
	mux          *sync.Mutex
}

func (sw *syncedWriter) Write(p []byte) (n int, err error) {
	sw.mux.Lock()
	defer sw.mux.Unlock()
	return sw.protectedDst.Write(p)
}
