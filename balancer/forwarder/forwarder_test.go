package forwarder

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestForward(t *testing.T) {

	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<html><body>server1</body></html>\n")
	}))

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<html><body>server2</body></html>")
	}))

	servers := []string{srv1.Listener.Addr().String(), srv2.Listener.Addr().String()}

	f := NewForwarder(servers)

	// Forward
	lAddr, _ := net.ResolveTCPAddr("tcp", "localhost:0")
	listener, err := net.ListenTCP("tcp", lAddr)
	assert.NoError(t, err)
	errorsChan := make(chan error, 1)
	go listenForTestRequest(f, listener, servers, errorsChan)
	// Write to client side
	resp, err := http.Get("http://" + listener.Addr().String() + "/anything")
	assert.NoError(t, err)

	t.Cleanup(func() {
		srv1.Close()
		srv2.Close()
		resp.Body.Close()
	})
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(body), "<html><body>server1</body></html>\n")

	select {
	case errs := <-errorsChan:
		assert.NoError(t, errs)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "timeout waiting for errors")
	}

}

func listenForTestRequest(f *forward, listener *net.TCPListener, urls []string, errorsChan chan<- error) {

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			continue
		}

		err = f.Forward(clientConn, urls)
		errorsChan <- err
	}
}
