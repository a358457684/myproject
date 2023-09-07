package utils

import (
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// http请求相关默认配置
const (
	DefaultHttpDialTimeout         = 20 * time.Second
	DefaultHttpKeepAlive           = 120 * time.Second
	DefaultHttpMaxIdleConns        = 1000
	DefaultHttpMaxIdleConnsPerHost = 1000
	DefaultHttpIdleConnTimeout     = 90 * time.Second
	DefaultHttpTimeout             = 30 * time.Second
)

// http请求配置
type RequestPromise struct {
	headers     http.Header
	encoding    Charset
	timeout     time.Duration
	proxy       func(*http.Request) (*url.URL, error)
	dialContext func(ctx context.Context, network, addr string) (net.Conn, error)
	client      *http.Client
	isSkipTls   bool
}

// 返回一个http请求配置对象，默认带上压缩头
func NewRequest() *RequestPromise {
	return (&RequestPromise{}).
		SetHeader("Accept-Encoding", "gzip, deflate, zlib")
}

// 返回一个http请求配置对象，默认带上压缩头和Content-Type = application/json; charset=utf-8
func JSONRequest() *RequestPromise {
	return (&RequestPromise{}).
		SetHeader("Accept-Encoding", "gzip, deflate, zlib").
		SetHeader("Content-Type", "application/json; charset=utf-8")
}

// 返回一个http请求配置对象，默认带上压缩头和Content-Type = application/xml; charset=utf-8
func XMLRequest() *RequestPromise {
	return (&RequestPromise{}).
		SetHeader("Accept-Encoding", "gzip, deflate, zlib").
		SetHeader("Content-Type", "application/xml; charset=utf-8")
}

// 返回一个http请求配置对象，默认带上压缩头和Content-Type = application/x-www-form-urlencoded; charset=utf-8
func FormRequest() *RequestPromise {
	return (&RequestPromise{}).
		SetHeader("Accept-Encoding", "gzip, deflate, zlib").
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
}

// 返回一个采用了连接池和Keepalive配置的http client，可以配合RequestPromise的SetClient函数使用
// 默认不使用它，而是每次请求新建连接
func NewPoolingHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   DefaultHttpDialTimeout,
				KeepAlive: DefaultHttpKeepAlive,
			}).DialContext,
			MaxIdleConns:        DefaultHttpMaxIdleConns,
			MaxIdleConnsPerHost: DefaultHttpMaxIdleConnsPerHost,
			IdleConnTimeout:     DefaultHttpIdleConnTimeout,
		},
		Timeout: DefaultHttpTimeout, // 此处设置小于等于零的值，意为不超时
	}
}

// 设置https忽略本地证书校验
func (r *RequestPromise) SetSkipTls() *RequestPromise {
	r.isSkipTls = true
	return r
}

// 设置http header
func (r *RequestPromise) SetHeader(key string, value string) *RequestPromise {
	if len(strings.TrimSpace(key)) == 0 {
		return r
	}
	key = strings.TrimSpace(key)
	if nil == r.headers {
		r.headers = make(http.Header)
	}
	r.headers.Set(key, value)
	return r
}

// 设置http响应的编码，默认utf8
func (r *RequestPromise) SetEncoding(encoding Charset) *RequestPromise {
	if encoding == UTF8 {
		return r
	}
	r.encoding = encoding
	return r
}

// 设置超时时间，从连接到接收到响应的总时间
// 如果此处不设置则采用http client中设置的超时时间，默认http client超时时间30秒
// 如果此处设置不等于零的值，则覆盖http client中设置的超时时间
// 如果此处设置小于零的值，意为不超时
func (r *RequestPromise) SetTimeout(timeout time.Duration) *RequestPromise {
	if timeout == 0 {
		return r
	}
	r.timeout = timeout
	return r
}

// 设置http或https代理，默认无代理
func (r *RequestPromise) SetHttpProxy(proxyUri string) *RequestPromise {
	if len(strings.TrimSpace(proxyUri)) == 0 {
		return r
	}
	proxyUri = strings.TrimSpace(proxyUri)
	uri, err := (&url.URL{}).Parse(proxyUri)
	if nil != err {
		return r
	}
	r.proxy = http.ProxyURL(uri)
	return r
}

// 设置socket5代理，默认无代理
func (r *RequestPromise) SetSocket5Proxy(proxyUri string) *RequestPromise {
	if len(strings.TrimSpace(proxyUri)) == 0 {
		return r
	}
	proxyUri = strings.TrimSpace(proxyUri)
	dialer, err := proxy.SOCKS5("tcp", proxyUri, nil, proxy.Direct)
	if nil != err {
		return r
	}
	r.dialContext = func(_ context.Context, network string, address string) (net.Conn, error) {
		return dialer.Dial(network, address)
	}
	return r
}

// 设置事先实例化好的http client，默认每次请求会新建一个http client
func (r *RequestPromise) SetClient(client *http.Client) *RequestPromise {
	if nil == client {
		return r
	}
	r.client = client
	return r
}

// 发起请求并返回响应内容
// FORM方式提交数据请设置Content-Type=application/x-www-form-urlencoded请求头，且io.Reader传url.Values.Encode得到的字符串的reader
func (r *RequestPromise) Call(httpMethod string, targetUri string, data io.Reader) ([]byte, error) {
	targetUri = strings.TrimSpace(targetUri)
	if len(targetUri) == 0 {
		return nil, nil
	}

	// http request handle
	if len(strings.TrimSpace(httpMethod)) == 0 {
		httpMethod = http.MethodGet
	} else {
		httpMethod = strings.ToUpper(strings.TrimSpace(httpMethod))
	}
	req, err := http.NewRequest(httpMethod, targetUri, data)
	if err != nil {
		return nil, err
	}
	if nil != r.headers {
		req.Header = r.headers
	}

	r.initClient()

	// send http request & get http response
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	return r.readResponseBody(resp)
}

func (r *RequestPromise) initClient() {
	// http client handle
	if nil == r.client { // create new http client instance
		r.client = &http.Client{Timeout: DefaultHttpTimeout} // default timeout
	}
	if r.timeout < 0 {
		r.timeout = DefaultHttpTimeout // default timeout
	}
	if r.timeout > 0 {
		r.client.Timeout = r.timeout
	}
	if r.isSkipTls {
		if nil == r.client.Transport {
			r.client.Transport = &http.Transport{}
		}
		transport := (r.client.Transport).(*http.Transport)
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
			//VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			//	return nil
			//},
		}
	}
	if nil != r.proxy || nil != r.dialContext {
		if nil == r.client.Transport {
			r.client.Transport = &http.Transport{}
		}
		transport := (r.client.Transport).(*http.Transport)
		if nil != r.proxy {
			transport.Proxy = r.proxy
		}
		if nil != r.dialContext {
			transport.DialContext = r.dialContext
		}
	}
}

func (r *RequestPromise) readResponseBody(resp *http.Response) ([]byte, error) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(resp.Body)
		defer func(reader io.Closer) {
			_ = reader.Close()
		}(reader)
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer func(reader io.Closer) {
			_ = reader.Close()
		}(reader)
	case "zlib":
		reader, _ = zlib.NewReader(resp.Body)
		defer func(reader io.Closer) {
			_ = reader.Close()
		}(reader)
	default:
		reader = resp.Body
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if r.encoding != "" {
		body = ConvertToEncodingBytes(body, r.encoding)
	}
	return body, nil
}
