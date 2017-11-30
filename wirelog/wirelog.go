package wirelog

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/proxy"
)

// NewHTTPTransport provides the default http.Transport
// from https://golang.org/src/net/http/transport.go
func NewHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// EnableProxy sets up a proxy
func EnableProxy(trans *http.Transport, viaAddress string) error {
	d, err := proxy.SOCKS5("tcp", viaAddress, nil, proxy.Direct)
	if err != nil {
		return err
	}
	trans.Dial = d.Dial
	return nil
}

// LogToFile provides quick setup for file logging
func LogToFile(trans *http.Transport, log string, disableZip, insecure bool) (io.Closer, error) {
	file, err := os.OpenFile(log, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		err = os.MkdirAll(filepath.Dir(log), 0777)
		if err != nil {
			return nil, err
		}
		file, err = os.Create(log)
		if err != nil {
			return nil, err
		}
	}
	return file, LogToWriter(trans, file, disableZip, insecure)
}

// LogToWriter provides the basic setup for http wirelogging
func LogToWriter(trans *http.Transport, log io.Writer, disableZip, insecure bool) error {
	trans.DisableCompression = disableZip
	trans.Dial = Plain(log, trans.Dial)
	trans.DialTLS = TLS(log, insecure)
	return nil
}

// Dialer just makes the return type for the Dialer function reasonable
type Dialer func(network, addr string) (net.Conn, error)

// Plain wraps the standard Dialer
func Plain(log io.Writer, dial Dialer) Dialer {
	return func(network, addr string) (net.Conn, error) {
		conn, err := dial(network, addr)
		wire := Conn{
			log:  log,
			Conn: conn,
		}
		return &wire, err
	}
}

// TLS wraps encrypted dialers
func TLS(log io.Writer, insecureSkipVerify bool) Dialer {
	return TLSConfig(log, &tls.Config{InsecureSkipVerify: insecureSkipVerify})
}

// TLSConfig wraps encrypted dialers
func TLSConfig(log io.Writer, config *tls.Config) Dialer {
	return func(network, addr string) (net.Conn, error) {
		c, err := tls.Dial(network, addr, config)
		if err != nil {
			return nil, err
		}
		wire := Conn{
			log:  log,
			Conn: c,
		}
		return &wire, c.Handshake()
	}
}

// Conn ....
type Conn struct {
	// embedded
	net.Conn
	// the destination for the split stream
	log io.Writer
}

func (c *Conn) Read(b []byte) (n int, err error) {
	read, err := c.Conn.Read(b)
	c.log.Write(b[0:read])
	return read, err
}
func (c *Conn) Write(b []byte) (n int, err error) {
	write, err := c.Conn.Write(b)
	c.log.Write(b[0:write])
	return write, err
}
