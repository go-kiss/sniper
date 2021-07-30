package xhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"sniper/pkg/errors"

	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
)

type GrpcClient interface {
	DoUnary(ctx context.Context, api string, req, resp proto.Message) (h2resp *http.Response, err error)
}

var plainTextH2Transport = &http2.Transport{
	AllowHTTP: true,
	DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
		return net.Dial(network, addr)
	},
}

func NewGrpcClient(timeout time.Duration) GrpcClient {
	return &myClient{
		cli: &http.Client{
			Transport: plainTextH2Transport,
			Timeout:   timeout,
		},
	}
}

func (c *myClient) DoUnary(ctx context.Context, api string, req, resp proto.Message) (h2resp *http.Response, err error) {
	rpb, err := proto.Marshal(req)
	if err != nil {
		return
	}

	// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
	lpb := len(rpb)
	buf := &bytes.Buffer{}
	buf.Grow(lpb + 5)
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(lpb))
	buf.WriteByte(0) // 不压缩
	buf.Write(bs)    // 写入长度
	buf.Write(rpb)   // 写入消息内容

	h2req, err := http.NewRequest("POST", api, buf)
	if err != nil {
		return
	}

	h2req = h2req.WithContext(ctx)

	h2req.Header.Set("trailers", "TE")
	h2req.Header.Set("content-type", "application/grpc+proto")

	h2resp, err = c.Do(ctx, h2req)
	if err != nil {
		return
	}
	defer h2resp.Body.Close()

	pb, err := ioutil.ReadAll(h2resp.Body)
	if err != nil {
		return
	}

	if status := h2resp.Trailer.Get("grpc-status"); status != "0" {
		err = errors.Errorf("grpc status: %s grpc message: %s", status, h2resp.Trailer.Get("grpc-message"))
		return
	}

	// 因为是 Unary，可以直接跳过前五个字节
	err = proto.Unmarshal(pb[5:], resp)

	return
}
