package openviking

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type Config struct{ BaseURL, APIKey, Account, User, RootURI string }
type Client struct {
	config Config
	http   *http.Client
}

func NewClient(c Config) (*Client, error) {
	for _, v := range []string{c.BaseURL, c.APIKey, c.Account, c.User, c.RootURI} {
		if v != strings.TrimSpace(v) || strings.ContainsAny(v, "\r\n") {
			return nil, fmt.Errorf("OpenViking configuration must use canonical values")
		}
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.Contains(c.BaseURL, "#") || u.Opaque != "" || u.RawPath != "" {
		return nil, fmt.Errorf("OPENVIKING_BASE_URL requires an HTTP(S) URL without credentials, query or fragment")
	}
	if trimmed := strings.TrimSuffix(u.Path, "/"); trimmed != "" && path.Clean(u.Path) != trimmed {
		return nil, fmt.Errorf("OPENVIKING_BASE_URL requires a canonical path")
	}
	u.Host = strings.ToLower(u.Host)
	if u.Scheme == "http" && u.Port() == "80" || u.Scheme == "https" && u.Port() == "443" {
		u.Host = u.Hostname()
		if strings.Contains(u.Host, ":") {
			u.Host = "[" + u.Host + "]"
		}
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	c.BaseURL = u.String()
	if (c.Account == "") != (c.User == "") {
		return nil, fmt.Errorf("OPENVIKING_ACCOUNT and OPENVIKING_USER must be configured together")
	}
	if c.RootURI == "" {
		c.RootURI = "viking://~/resources/astrocyte"
	}
	c.RootURI, err = domain.CanonicalVikingURI(c.RootURI)
	if err != nil {
		return nil, fmt.Errorf("OPENVIKING_ROOT_URI: %w", err)
	}
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")
	return &Client{config: c, http: &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return errors.New("OpenViking redirects are not permitted")
	}}}, nil
}

func failure(code apierrors.Code, message, action string) error {
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: action}
}

func (c *Client) request(ctx context.Context, method, endpoint, contentType string, body io.Reader) ([]byte, error) {
	mutation := method == http.MethodPost && endpoint != "/api/v1/search/find"
	code := apierrors.ProviderUnavailable
	if mutation {
		code = apierrors.DeliveryUnknown
	}
	r, err := http.NewRequestWithContext(ctx, method, c.config.BaseURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	if c.config.APIKey != "" {
		r.Header.Set("X-API-Key", c.config.APIKey)
	}
	if c.config.Account != "" {
		r.Header.Set("X-OpenViking-Account", c.config.Account)
		r.Header.Set("X-OpenViking-User", c.config.User)
	}
	response, err := c.http.Do(r)
	if err != nil {
		return nil, failure(code, "OpenViking 请求未取得确定响应；请核实原操作", "inspect_original_operation")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if response.StatusCode == 401 || response.StatusCode == 403 {
			code = apierrors.ScopeDenied
		} else if !mutation || (response.StatusCode >= 400 && response.StatusCode < 500) {
			code = apierrors.ProviderUnavailable
		}
		return nil, failure(code, fmt.Sprintf("OpenViking 返回 HTTP %d", response.StatusCode), "inspect_upstream_response")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (12<<20)+1))
	if err != nil || len(raw) > 12<<20 {
		return nil, failure(code, "OpenViking 响应读取失败或超过 12 MiB", "inspect_upstream_response")
	}
	var envelope struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.Status != "ok" {
		return nil, failure(code, "OpenViking 响应格式或业务状态无效", "inspect_upstream_response")
	}
	return raw, nil
}

func decodeResult[T any](raw []byte) (T, error) {
	var result T
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return result, err
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return result, fmt.Errorf("OpenViking result is missing")
	}
	err := json.Unmarshal(envelope.Result, &result)
	return result, err
}

var identifier = regexp.MustCompile(`^[A-Za-z0-9_-]{1,160}$`)

func (c *Client) Status(ctx context.Context) (app.OpenVikingStatus, error) {
	result := app.OpenVikingStatus{SchemaVersion: 1, Configured: true, BaseURL: c.config.BaseURL, RootURI: c.config.RootURI}
	raw, err := c.request(ctx, http.MethodGet, "/health", "", nil)
	if err != nil {
		return result, err
	}
	var health struct {
		Healthy   bool   `json:"healthy"`
		Version   string `json:"version"`
		AccountID string `json:"account_id"`
		UserID    string `json:"user_id"`
		Role      string `json:"role"`
	}
	if err := json.Unmarshal(raw, &health); err != nil {
		return result, err
	}
	if !health.Healthy {
		return result, failure(apierrors.ProviderUnavailable, "OpenViking 健康检查失败", "check_upstream_health")
	}
	raw, err = c.request(ctx, http.MethodGet, "/api/v1/system/status", "", nil)
	if err != nil {
		return result, err
	}
	identity, err := decodeResult[struct {
		Initialized bool   `json:"initialized"`
		User        string `json:"user"`
	}](raw)
	if err != nil || !identity.Initialized || !identifier.MatchString(identity.User) || !identifier.MatchString(health.AccountID) || health.UserID != identity.User || health.Role == "" || (c.config.Account != "" && c.config.Account != health.AccountID) || (c.config.User != "" && c.config.User != health.UserID) {
		return result, failure(apierrors.ProviderUnavailable, "OpenViking 未返回有效的已初始化身份", "verify_upstream_identity")
	}
	root := c.config.RootURI
	if strings.HasPrefix(root, "viking://~/") {
		root = strings.Replace(root, "viking://~/", "viking://user/"+identity.User+"/", 1)
	}
	if err := domain.VikingWithin(root, "viking://user/"+identity.User+"/resources"); err != nil && !strings.HasPrefix(root, "viking://resources/") {
		return result, failure(apierrors.ScopeDenied, "OpenViking 配置根不属于当前身份的资料范围", "correct_server_root")
	}
	result.RootURI = root
	result.User = identity.User
	result.Account = health.AccountID
	result.Connected = true
	result.Version = health.Version
	result.Message = "已核验服务健康与认证身份"
	return result, nil
}

func (c *Client) Upload(ctx context.Context, title, text string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "astrocyte-source.md")
	if err != nil {
		return "", err
	}
	if _, err = io.WriteString(file, text); err != nil {
		return "", err
	}
	if err = writer.Close(); err != nil {
		return "", err
	}
	raw, err := c.request(ctx, http.MethodPost, "/api/v1/resources/temp_upload", writer.FormDataContentType(), &body)
	if err != nil {
		return "", err
	}
	result, err := decodeResult[struct {
		ID string `json:"temp_file_id"`
	}](raw)
	if err != nil || result.ID == "" || len(result.ID) > 512 {
		return "", failure(apierrors.DeliveryUnknown, "临时上传未返回有效标识，禁止重发", "inspect_original_upload")
	}
	return result.ID, nil
}

func (c *Client) Submit(ctx context.Context, tempID, parent, title string) (string, string, error) {
	body, err := json.Marshal(map[string]any{"temp_file_id": tempID, "parent": parent, "create_parent": true, "source_name": title, "wait": false})
	if err != nil {
		return "", "", err
	}
	raw, err := c.request(ctx, http.MethodPost, "/api/v1/resources", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	result, err := decodeResult[struct {
		TaskID  string `json:"task_id"`
		RootURI string `json:"root_uri"`
	}](raw)
	if err != nil || !identifier.MatchString(result.TaskID) {
		return "", "", failure(apierrors.DeliveryUnknown, "资源提交未返回有效任务标识，禁止重发", "inspect_original_submission")
	}
	if result.RootURI != "" {
		if err := domain.VikingWithin(result.RootURI, parent); err != nil || result.RootURI == parent {
			return "", "", failure(apierrors.DeliveryUnknown, "资源提交返回越界 URI，禁止重发", "inspect_original_submission")
		}
	}
	if result.RootURI != "" {
		result.RootURI, err = domain.CanonicalVikingURI(result.RootURI)
		if err != nil {
			return "", "", failure(apierrors.DeliveryUnknown, "资源 URI 规范化失败", "inspect_original_submission")
		}
	}
	return result.TaskID, result.RootURI, nil
}

func (c *Client) Task(ctx context.Context, id string) (app.OpenVikingTask, error) {
	if !identifier.MatchString(id) {
		return app.OpenVikingTask{}, failure(apierrors.ValidationFailed, "任务标识无效", "inspect_local_submission")
	}
	raw, err := c.request(ctx, http.MethodGet, "/api/v1/tasks/"+url.PathEscape(id)+"?include_events=true", "", nil)
	if err != nil {
		return app.OpenVikingTask{}, err
	}
	result, err := decodeResult[app.OpenVikingTask](raw)
	if err != nil {
		return result, failure(apierrors.ProviderUnavailable, "OpenViking 任务响应格式无效", "inspect_upstream_task")
	}
	return result, nil
}

func (c *Client) Find(ctx context.Context, root, query string) ([]app.OpenVikingHit, error) {
	body, err := json.Marshal(map[string]any{"query": query, "target_uri": root, "context_type": "resource", "limit": 10, "include_provenance": true})
	if err != nil {
		return nil, err
	}
	raw, err := c.request(ctx, http.MethodPost, "/api/v1/search/find", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	result, err := decodeResult[struct {
		Resources []struct {
			URI         *string  `json:"uri"`
			Abstract    *string  `json:"abstract"`
			Overview    *string  `json:"overview"`
			Level       *int     `json:"level"`
			Score       *float64 `json:"score"`
			ContextType *string  `json:"context_type"`
		} `json:"resources"`
	}](raw)
	if err != nil || result.Resources == nil || len(result.Resources) > 10 {
		return nil, failure(apierrors.ProviderUnavailable, "OpenViking 检索集合缺失或超过上限", "inspect_upstream_results")
	}
	hits := make([]app.OpenVikingHit, 0, len(result.Resources))
	for _, hit := range result.Resources {
		if hit.URI == nil || hit.Abstract == nil || hit.Level == nil || hit.Score == nil || hit.ContextType == nil {
			return nil, failure(apierrors.ProviderUnavailable, "OpenViking 检索结果缺少必需字段", "inspect_upstream_results")
		}
		canonical, err := domain.CanonicalVikingURI(*hit.URI)
		if err != nil {
			return nil, failure(apierrors.ProviderUnavailable, "OpenViking 检索 URI 无效", "inspect_upstream_results")
		}
		hits = append(hits, app.OpenVikingHit{URI: canonical, Abstract: *hit.Abstract, Overview: hit.Overview, Level: *hit.Level, Score: *hit.Score, ContextType: *hit.ContextType})
	}
	return hits, nil
}

func (c *Client) Read(ctx context.Context, uri string, level int) (string, error) {
	if level < 0 || level > 2 {
		return "", failure(apierrors.ValidationFailed, "无效内容层级", "choose_supported_level")
	}
	action := []string{"abstract", "overview", "read"}[level]
	raw, err := c.request(ctx, http.MethodGet, "/api/v1/content/"+action+"?"+url.Values{"uri": []string{uri}}.Encode(), "", nil)
	if err != nil {
		return "", err
	}
	result, err := decodeResult[string](raw)
	if err != nil {
		return "", failure(apierrors.ProviderUnavailable, "OpenViking 内容必须返回字符串", "inspect_upstream_content")
	}
	return result, nil
}
