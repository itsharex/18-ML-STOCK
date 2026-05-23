package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// defaultUA 全包默认 User-Agent。站点对此较敏感（东财/腾讯/雅虎都需要真实的浏览器 UA）。
const defaultUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// sharedTransport 全包共享的连接池。
// 注意：默认开启 keep-alive；需要禁用复用的反爬路径走 httpGetNoKeepAlive。
var sharedTransport = &http.Transport{
	MaxIdleConns:        32,
	MaxIdleConnsPerHost: 8,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
}

// sharedClient 全包共享的 HTTP 客户端。
// 不设 Client.Timeout —— 用 per-call ctx.WithTimeout 精细控制。
var sharedClient = &http.Client{Transport: sharedTransport}

// SharedTransport 暴露共享 Transport，给需要嵌入 client 到 struct 的 SDK 风格调用方
// （如 SFLClient）使用，从而共享连接池。
func SharedTransport() *http.Transport { return sharedTransport }

// RequestOption 自定义请求构造，例如附加 Referer 或覆盖 Accept。
type RequestOption func(*http.Request)

// WithHeader 设置任意 header。
func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) { r.Header.Set(key, value) }
}

// WithReferer 等价于 WithHeader("Referer", ref)。
func WithReferer(ref string) RequestOption { return WithHeader("Referer", ref) }

// HTTPGet 用共享 client 发 GET，默认带 UA 与 Accept-Language。
// 调用方应在 ctx 上携带 timeout（context.WithTimeout）。
// 返回 body 字节；非 2xx 返回 error（含 status + body 摘要便于诊断）。
func HTTPGet(ctx context.Context, url string, opts ...RequestOption) ([]byte, error) {
	return doRequest(ctx, sharedClient, http.MethodGet, url, nil, opts...)
}

// HTTPGetNoKeepAlive 用每次新建的、禁用 keep-alive 的 client 发 GET。
// 仅当目标服务端会主动断开 keep-alive 复用的旧连接导致 EOF 时使用
// （例如东财部分反爬路径）。其它情况一律用 HTTPGet。
func HTTPGetNoKeepAlive(ctx context.Context, url string, opts ...RequestOption) ([]byte, error) {
	c := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	return doRequest(ctx, c, http.MethodGet, url, nil, opts...)
}

// HTTPDo 给特殊请求构造（自定义 method/body/headers）兜底的低层入口。
// 调用方拥有 req 后可自由设置；本函数只负责执行与 body 读取/状态校验。
func HTTPDo(req *http.Request) ([]byte, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", defaultUA)
	}
	resp, err := sharedClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := data
		if len(snippet) > 256 {
			snippet = snippet[:256]
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(snippet))
	}
	return data, nil
}

func doRequest(ctx context.Context, client *http.Client, method, url string,
	body io.Reader, opts ...RequestOption) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUA)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	for _, opt := range opts {
		opt(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := data
		if len(snippet) > 256 {
			snippet = snippet[:256]
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(snippet))
	}
	return data, nil
}
