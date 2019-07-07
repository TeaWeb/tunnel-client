package tunnel_client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/logs"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	localHost := this.config.LocalHost()
	root := this.config.Root
	scheme := this.config.LocalScheme()

	if len(localHost) == 0 && len(root) == 0 {
		return errors.New("'local' or 'root' should not be empty")
	}

	hasLocal := len(localHost) > 0
	hasRoot := len(root) > 0

	if len(scheme) == 0 {
		scheme = "http"
	}

	for {
		this.connLocker.Lock()
		if len(this.conns) >= runtime.NumCPU()*2 {
			this.connLocker.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		this.connLocker.Unlock()

		conn, err := net.Dial("tcp", this.config.Remote)
		if err != nil {
			logs.Println("[error]" + err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		this.connLocker.Lock()
		this.conns = append(this.conns, conn)
		this.connLocker.Unlock()

		go func(conn net.Conn) {
			if len(this.config.Secret) > 0 {
				conn.Write([]byte(this.config.Secret + "\n"))
			}

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

				// special urls
				if len(req.Host) == 0 {
					if req.URL.Path == "/$$TEA/ping" { // ping
						body := []byte("OK")
						resp := &http.Response{
							StatusCode:    http.StatusOK,
							Status:        "Ok",
							Proto:         "HTTP/1.1",
							ProtoMajor:    1,
							ProtoMinor:    1,
							ContentLength: int64(len(body)),
							Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
						}
						data, err := httputil.DumpResponse(resp, true)
						if err != nil {
							logs.Error(err)
						} else {
							_, err = conn.Write(data)
							if err != nil {
								logs.Error(err)
							}
						}
						resp.Body.Close()
						continue
					}
				}

				logs.Println(req.Header.Get("X-Forwarded-For") + " - \"" + req.Method + " " + req.URL.String() + "\" \"" + req.Header.Get("User-Agent") + "\"")

				if hasLocal { // read from local web server
					req.RequestURI = ""
					req.URL.Host = localHost
					req.URL.Scheme = scheme

					if len(this.config.Host) > 0 {
						req.Host = this.config.Host
					} else {
						forwardedHost := req.Header.Get("X-Forwarded-Host")
						if len(forwardedHost) > 0 {
							req.Host = forwardedHost
						} else {
							req.Host = localHost
						}
					}

					resp, err := HttpClient.Do(req)
					if err != nil {
						logs.Error(err)
						resp := &http.Response{
							StatusCode: http.StatusBadGateway,
							Status:     "Bad Gateway",
							Header: map[string][]string{
								"Content-Type": {"text/plain"},
								"Connection":   {"keep-alive"},
							},
							Proto:      "HTTP/1.1",
							ProtoMajor: 1,
							ProtoMinor: 1,
						}
						data, err := httputil.DumpResponse(resp, false)
						if err != nil {
							logs.Error(err)
							this.writeServerError(conn)
							continue
						}
						conn.Write(data)
					} else {
						resp.Header.Set("Connection", "keep-alive")
						data, err := httputil.DumpResponse(resp, true)
						if err != nil {
							logs.Error(err)
							resp.Body.Close()
							this.writeServerError(conn)
							continue
						}
						conn.Write(data)
						resp.Body.Close()
					}
				} else if hasRoot { // read from root directory
					requestPath := req.URL.Path
					if len(requestPath) == 0 || requestPath == "/" {
						path, stat, found := this.findIndexPage(root + Tea.DS)
						if found {
							this.writeFile(conn, req, stat, path)
							continue
						}
						this.writeNotFound(conn)
						continue
					}

					path := root + Tea.DS + requestPath

					if strings.HasSuffix(path, "/") {
						path, stat, found := this.findIndexPage(path)
						if found {
							this.writeFile(conn, req, stat, path)
							continue
						}
						this.writeNotFound(conn)
						continue
					}

					stat, err := os.Stat(path)
					if err != nil {
						if os.IsNotExist(err) {
							this.writeString(conn, http.StatusNotFound, "File Not Found")
						} else {
							logs.Error(err)
							this.writeServerError(conn)
						}
						continue
					} else if stat.IsDir() { // try again
						path, stat, found := this.findIndexPage(path)
						if found {
							this.writeFile(conn, req, stat, path)
							continue
						}
						this.writeNotFound(conn)
						continue
					}

					this.writeFile(conn, req, stat, path)
				}
			}
		}(conn)
	}
	return nil
}

func (this *Tunnel) writeString(conn net.Conn, code int, data string) {
	dataBytes := []byte(data)
	this.writeBytes(conn, code, dataBytes)
}

func (this *Tunnel) writeBytes(conn net.Conn, code int, data []byte) {
	resp := &http.Response{
		StatusCode:    code,
		ContentLength: int64(len(data)),
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header: map[string][]string{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	respData, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logs.Error(err)
		resp.Body.Close()
		return
	}

	_, err = conn.Write(respData)
	if err != nil {
		logs.Error(err)
	}
	resp.Body.Close()
}

func (this *Tunnel) writeServerError(conn net.Conn) {
	this.writeString(conn, http.StatusInternalServerError, "Internal Server Error")
}

func (this *Tunnel) writeNotFound(conn net.Conn) {
	this.writeString(conn, http.StatusNotFound, "File Not Found")
}

func (this *Tunnel) writeFile(conn net.Conn, req *http.Request, stat os.FileInfo, path string) {
	reader, err := os.OpenFile(path, os.O_RDONLY, 444)
	if err != nil {
		logs.Error(err)
		this.writeServerError(conn)
		return
	}
	defer reader.Close()

	resp := &http.Response{
		StatusCode:    200,
		ContentLength: stat.Size(),
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        map[string][]string{},
	}

	// mime type
	ext := filepath.Ext(path)
	if len(ext) > 0 {
		mimeType := mime.TypeByExtension(ext)
		if len(mimeType) > 0 {
			resp.Header.Set("Content-Type", mimeType)
		}
	}

	// supports Last-Modified
	modifiedTime := stat.ModTime().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	resp.Header.Set("Last-Modified", modifiedTime)

	// supports ETag
	eTag := "\"et" + stringutil.Md5(fmt.Sprintf("%d,%d", stat.ModTime().UnixNano(), stat.Size())) + "\""
	resp.Header.Set("ETag", eTag)

	// supports If-None-Match
	if req.Header.Get("If-None-Match") == eTag {
		this.writeBytes(conn, http.StatusNotModified, []byte{})
		return
	}

	// supports If-Modified-Since
	if req.Header.Get("If-Modified-Since") == modifiedTime {
		this.writeBytes(conn, http.StatusNotModified, []byte{})
		return
	}

	//write body
	resp.Body = reader
	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		logs.Error(err)
		this.writeServerError(conn)
		return
	}

	conn.Write(data)
}

func (this *Tunnel) findIndexPage(dir string) (path string, stat os.FileInfo, found bool) {
	if len(this.config.Index) == 0 {
		this.config.Index = []string{"index.html", "default.html"}
	}
	for _, index := range this.config.Index {
		path := dir + Tea.DS + index
		stat, err := os.Stat(path)
		if err != nil || stat.IsDir() {
			continue
		}
		return path, stat, true
	}
	return
}
