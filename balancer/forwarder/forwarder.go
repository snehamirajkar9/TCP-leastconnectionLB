package forwarder

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	terminationDelay = 2 * time.Second
	maxRetry         = 3
)

type upstream struct {
	addr            string
	load            *atomic.Int32
	healthy         *atomic.Bool
	healthcheckStop chan bool
}

// forward forwards connections from src to dst using the least connections algorithm.
type forward struct {
	upstreams map[string]*upstream
}

type Forwarder interface {
	Forward(src net.Conn, allowedUpstreams []string) error
	Stop()
}

// Stop stops the healthchecks for all upstreams.
func (f *forward) Stop() {
	for _, u := range f.upstreams {
		u.healthcheckStop <- true
		close(u.healthcheckStop)
	}
}

// NewForwarder creates a new Forwarder.
func NewForwarder(upstreams []string) *forward {
	upstreamsMap := make(map[string]*upstream)
	for _, u := range upstreams {
		atomicZero := atomic.Int32{}
		atomicZero.Store(0)
		atomicTrue := atomic.Bool{}
		atomicTrue.Store(true)
		upstreamsMap[u] = &upstream{
			addr:    u,
			load:    &atomicZero,
			healthy: &atomicTrue,
		}
		stopChan := upstreamsMap[u].startHealthchecks()
		upstreamsMap[u].healthcheckStop = stopChan
	}
	return &forward{
		upstreams: upstreamsMap,
	}
}

// Forward forwards the connection src to the destination with the least connections.
func (f *forward) Forward(src net.Conn, allowedUpstreams []string) error {
	if src == nil {
		return fmt.Errorf("connection is null")
	}

	for i := 0; i < maxRetry; i++ {
		if i > 0 {
			log.Debug().Str("remoteAddr", src.RemoteAddr().String()).Msg("Retrying")
		}
		dst := f.getLeastConn(allowedUpstreams)
		if dst == nil {
			continue
		}
		if success, err := f.forward(src, dst); success {
			return err
		}
	}

	// Retries exceeded, return error
	return fmt.Errorf("max retries exceeded for %s", src.RemoteAddr())

}

func (f *forward) getLeastConn(allowed []string) *upstream {
	var leastUsed *upstream
	var leastCount int32 = math.MaxInt32
	for _, dst := range allowed {
		// if f.isUnhealthy(dst) {
		// 	continue
		// }
		// no checking, we consider the instanciator creates correct lists
		up := f.upstreams[dst]
		if up.isUnhealthy() {
			continue
		}
		count := up.load.Load()
		if count < leastCount {
			leastUsed = up
			leastCount = count
		}
	}

	return leastUsed
}

// forward returns false if the upstream is unhealthy
func (f *forward) forward(src net.Conn, dst *upstream) (bool, error) {
	defer src.Close()
	log.Debug().Str("destination", dst.addr).Msg("Forwarding to")

	dstConn, err := net.Dial("tcp", dst.addr)
	if err != nil {
		dst.setUnhealthy()
		log.Debug().Str("upstream", dst.addr).Msg("Marking as unhealthy")
		return false, nil
	}
	defer dstConn.Close()

	// TODO find the reason why it's needed in unit test
	dstConn.SetReadDeadline(time.Now().Add(terminationDelay))

	dst.incrementConnectionCount()
	var wg sync.WaitGroup
	var err1, err2 error
	wg.Add(1)
	go func() {
		defer wg.Done()
		err1 = f.copyData(dstConn, src)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		err2 = f.copyData(src, dstConn)
	}()

	wg.Wait()
	dst.decrementConnectionCount()

	return true, errors.Join(err1, err2)
}

// Function to copy data between two connections
func (f *forward) copyData(dst io.WriteCloser, src io.Reader) error {
	defer dst.Close()
	_, err := io.Copy(dst, src)
	// hack to remove normal close from errors
	if os.IsTimeout(err) || errors.Is(err, net.ErrClosed) {
		err = nil
	}
	return err
}

// Function to increment the connection count for an upstream server
func (u *upstream) incrementConnectionCount() {
	u.load.Add(1)
}

// Function to decrement the connection count for an upstream server
func (u *upstream) decrementConnectionCount() {
	u.load.Add(-1)
}
