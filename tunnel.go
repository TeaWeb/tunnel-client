package tunnel_client

import (
	"bufio"
	"errors"
	"github.com/iwind/TeaGo/logs"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

// tunnel definition
type Tunnel struct {
	config *TunnelConfig

	conns      []net.Conn
	connLocker sync.Mutex
}

func NewTunnel(config *TunnelConfig) *Tunnel {
	return &Tunnel{
		config: config,
	}
}

func (this *Tunnel) Start() error {
	host := this.config.LocalHost()
	scheme := this.config.LocalScheme()

	if len(host) == 0 {
		return errors.New("local host should not be empty")
	}

	if len(scheme) == 0 {
		scheme = "http"
	}

	for {
		this.connLocker.Lock()
		if len(this.conns) >= 16 {
			this.connLocker.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		this.connLocker.Unlock()

		conn, err := net.Dial("tcp", this.config.Remote)
		if err != nil {
			time.Sleep(10 * time.Second)
			logs.Println("[error]" + err.Error())
			continue
		}

		this.connLocker.Lock()
		this.conns = append(this.conns, conn)
		this.connLocker.Unlock()

		go func(conn net.Conn) {
			reader := bufio.NewReader(conn)
			for {
				req, err := http.ReadRequest(reader)
				if err != nil {
					if err != io.EOF {
						log.Println("[error]" + err.Error())
					}
					this.connLocker.Lock()
					result := []net.Conn{}
					for _, c := range this.conns {
						if c == conn {
							continue
						}
						result = []net.Conn{}
					}
					this.conns = result
					this.connLocker.Unlock()
					break
				}

				req.RequestURI = ""
				req.URL.Host = host
				req.URL.Scheme = scheme

				if len(this.config.Host) > 0 {
					req.Host = this.config.Host
				} else {
					req.Host = host
				}

				resp, err := HttpClient.Do(req)
				if err != nil {
					logs.Error(err)
					conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n"))
					conn.Write([]byte("Content-Type: text/plain\r\n\r\n"))
				} else {
					data, err := httputil.DumpResponse(resp, true)
					if err != nil {
						logs.Error(err)
						continue
					}
					conn.Write(data)
					resp.Body.Close()
				}
			}
		}(conn)
	}
	return nil
}
