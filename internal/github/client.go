package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/auth"
)

// Client は GitHub Notifications API クライアント。
//
// 取得（GET /notifications）は条件付きリクエスト（If-Modified-Since）と
// X-Poll-Interval ヘッダを扱う必要があるため、go-gh の RESTClient ではなく
// 素の *http.Client を使う（RESTClient は 304 をエラー化しヘッダを失うため）。
// 認証ヘッダは go-gh の DefaultHTTPClient の RoundTripper が自動付与する。
// 書き込み（既読化）は 2xx で完結するため RESTClient を使う。
type Client struct {
	httpc *http.Client
	rest  *api.RESTClient
	base  string // REST prefix 例: https://api.github.com/
}

// NewClient は gh のログイン情報を再利用してクライアントを生成する。
func NewClient() (*Client, error) {
	host, _ := auth.DefaultHost()
	httpc, err := api.DefaultHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("HTTP クライアント生成: %w", err)
	}
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("REST クライアント生成: %w", err)
	}
	return &Client{httpc: httpc, rest: rest, base: restPrefix(host)}, nil
}

// restPrefix は go-gh 内部の restPrefix を再現する。
func restPrefix(hostname string) string {
	h := strings.TrimSuffix(strings.ToLower(hostname), "/")
	if h == "" || h == "github.com" || h == "api.github.com" {
		return "https://api.github.com/"
	}
	// GitHub Enterprise Server
	return fmt.Sprintf("https://%s/api/v3/", h)
}

// ListOptions は GET /notifications のクエリ。
type ListOptions struct {
	All           bool
	Participating bool
	PerPage       int
}

// ListResult は取得結果。NotModified=true のとき Notifications は空（前回から変化なし）。
type ListResult struct {
	Notifications []Notification
	NotModified   bool
	LastModified  string // 次回 If-Modified-Since に渡す値
	PollInterval  int    // X-Poll-Interval（秒）。0 のときは未指定
}

// List は通知を取得する。lastModified が非空なら条件付きリクエストを行う。
func (c *Client) List(ctx context.Context, opts ListOptions, lastModified string) (ListResult, error) {
	q := url.Values{}
	if opts.All {
		q.Set("all", "true")
	}
	if opts.Participating {
		q.Set("participating", "true")
	}
	per := opts.PerPage
	if per <= 0 {
		per = 50
	}
	q.Set("per_page", strconv.Itoa(per))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"notifications?"+q.Encode(), nil)
	if err != nil {
		return ListResult{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return ListResult{}, err
	}
	defer resp.Body.Close()

	res := ListResult{LastModified: lastModified}
	if s := resp.Header.Get("X-Poll-Interval"); s != "" {
		if n, e := strconv.Atoi(s); e == nil {
			res.PollInterval = n
		}
	}
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		res.LastModified = lm
	}

	switch resp.StatusCode {
	case http.StatusNotModified:
		res.NotModified = true
		return res, nil
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, err
		}
		if err := json.Unmarshal(body, &res.Notifications); err != nil {
			return res, fmt.Errorf("通知のデコード: %w", err)
		}
		return res, nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return res, fmt.Errorf("通知取得に失敗 (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}
}

// MarkThreadRead は単一スレッドを既読にする（PATCH /notifications/threads/{id}、205 Reset Content）。
func (c *Client) MarkThreadRead(ctx context.Context, threadID string) error {
	resp, err := c.rest.RequestWithContext(ctx, http.MethodPatch, "notifications/threads/"+threadID, nil)
	if err != nil {
		return fmt.Errorf("スレッド %s の既読化: %w", threadID, err)
	}
	resp.Body.Close()
	return nil
}

// MarkAllRead は通知をすべて既読にする（PUT /notifications）。
func (c *Client) MarkAllRead(ctx context.Context) error {
	resp, err := c.rest.RequestWithContext(ctx, http.MethodPut, "notifications", strings.NewReader(`{"read":true}`))
	if err != nil {
		return fmt.Errorf("全件既読化: %w", err)
	}
	resp.Body.Close()
	return nil
}
