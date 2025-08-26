package turvo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/maceo-kwik/drumkit/backend/internal/config"
)

// RateLimitedError represents an HTTP 429 response from Turvo or an internal
// cooldown state. RetryAfter indicates how long to wait before retrying.
type RateLimitedError struct {
	RetryAfter time.Duration
	Message    string
}

func (e RateLimitedError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("rate limited: retry after %s - %s", e.RetryAfter, e.Message)
	}
	return fmt.Sprintf("rate limited: retry after %s", e.RetryAfter)
}

// Client implements a minimal Turvo Public API client with OAuth token
// acquisition, API key support, and shipment/customer operations used by Drumkit.
type Client struct {
	httpClient *http.Client
	config     *config.Config
	mu         sync.Mutex
	token      string
	tokenExp   time.Time
	refresh    string
	// simple cooldown to avoid hammering oauth on 429
	nextOAuthAttempt time.Time
}

// NewClient creates a new Turvo API client.
func NewClient(cfg *config.Config) (*Client, error) {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config:     cfg,
	}
	return c, nil
}

func (c *Client) oauthTokenEndpoint() string {
	base := strings.TrimRight(c.config.TurvoBaseURL, "/")
	// OAuth docs specify /v1/oauth/token on publicapi host
	return fmt.Sprintf("%s/%s/oauth/token", base, "v1")
}

// fetchToken ensures there is a valid bearer token. It can use a refresh token
// when available and sets a simple cooldown after 429 responses.
func (c *Client) fetchToken(ctx context.Context, useRefresh bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Until(c.tokenExp) > 60*time.Second {
		return nil
	}
	// backoff respect
	if time.Now().Before(c.nextOAuthAttempt) {
		wait := time.Until(c.nextOAuthAttempt)
		return RateLimitedError{RetryAfter: wait}
	}

	endpoint := c.oauthTokenEndpoint()
	q := url.Values{}
	q.Set("client_id", c.config.TurvoClientID)
	q.Set("client_secret", c.config.TurvoClientSecret)
	endpointWithQuery := endpoint + "?" + q.Encode()

	form := url.Values{}
	headers := make(http.Header)
	if c.config.TurvoAPIKey != "" {
		headers.Set("x-api-key", c.config.TurvoAPIKey)
	}
	if useRefresh && c.refresh != "" {
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", c.refresh)
	} else {
		form.Set("grant_type", "password")
		form.Set("username", c.config.TurvoOAuthUsername)
		form.Set("password", c.config.TurvoOAuthPassword)
		form.Set("scope", c.config.TurvoOAuthScope)
		form.Set("type", c.config.TurvoOAuthUserType)
	}

	log.Printf("Turvo OAuth: POST %s", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointWithQuery, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header = headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body once for logging and subsequent handling
	bodyBytes, _ := io.ReadAll(resp.Body)
	// Truncate for logging to avoid huge outputs
	maxLog := 2048
	bodyPreview := bodyBytes
	if len(bodyBytes) > maxLog {
		bodyPreview = bodyBytes[:maxLog]
	}
	log.Printf("Turvo OAuth response: %s - %s", resp.Status, string(bodyPreview))

	if resp.StatusCode == http.StatusTooManyRequests { // 429
		cooldown := 60 * time.Second
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
				cooldown = time.Duration(secs) * time.Second
			}
		}
		c.nextOAuthAttempt = time.Now().Add(cooldown)
		return RateLimitedError{RetryAfter: cooldown, Message: string(bodyBytes)}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oauth token error: %s - %s", resp.Status, string(bodyBytes))
	}
	var tok struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(bodyBytes, &tok); err != nil {
		return err
	}
	if strings.TrimSpace(tok.AccessToken) == "" {
		return fmt.Errorf("empty access_token from oauth")
	}
	c.token = strings.TrimSpace(tok.AccessToken)
	c.refresh = tok.RefreshToken
	if tok.ExpiresIn <= 0 {
		tok.ExpiresIn = 12 * 60 * 60
	}
	c.tokenExp = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	c.nextOAuthAttempt = time.Time{} // clear cooldown
	return nil
}

func (c *Client) buildPath(p string) string {
	base := strings.TrimRight(c.config.TurvoBaseURL, "/")
	prefix := strings.Trim(c.config.TurvoAPIPrefix, "/")
	// If we have a bearer token, use /v1; otherwise for API key-only mode on publicapi, use /public/v1 when prefix is v1
	if c.token != "" {
		prefix = "v1"
	} else if strings.Contains(base, "publicapi.") && (prefix == "v1" || prefix == "/v1") {
		prefix = "public/v1"
	}
	res := strings.TrimLeft(p, "/")
	addPrefix := prefix != "" && !(strings.HasSuffix(base, "/"+prefix) || strings.HasSuffix(base, prefix))
	full := base
	if addPrefix {
		full += "/" + prefix
	}
	full += "/" + res
	return full
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if err := c.fetchToken(ctx, false); err != nil {
		return nil, err
	}
	fullURL := c.buildPath(path)
	log.Printf("Turvo request: %s %s", method, fullURL)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}
	// Bearer and x-api-key on data requests per working curl
	if c.config.TurvoAPIKey != "" {
		req.Header.Set("x-api-key", c.config.TurvoAPIKey)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.config.TurvoTenant != "" {
		req.Header.Set("Tenant", c.config.TurvoTenant)
	}
	return req, nil
}

// Minimal customer projection
type MinimalCustomer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ListCustomers fetches customers with filters (minimal fields)
func (c *Client) ListCustomers(ctx context.Context, q url.Values) ([]MinimalCustomer, error) {
	if q == nil {
		q = url.Values{}
	}
	if _, ok := q["start"]; !ok {
		q.Set("start", "0")
	}
	if _, ok := q["pageSize"]; !ok {
		q.Set("pageSize", "50")
	}
	req, err := c.newRequest(ctx, http.MethodGet, "customers/list?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.fetchToken(ctx, true); err == nil {
			return c.ListCustomers(ctx, q)
		}
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list customers: %s - %s", resp.Status, string(b))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var wrapped struct {
		Status  string `json:"Status"`
		Details struct {
			Customers []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"customers"`
		} `json:"details"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapped); err == nil && wrapped.Details.Customers != nil {
		var out []MinimalCustomer
		for _, cst := range wrapped.Details.Customers {
			out = append(out, MinimalCustomer{ID: cst.ID, Name: cst.Name})
		}
		return out, nil
	}
	// fallback to array form
	var arr []MinimalCustomer
	if err := json.Unmarshal(bodyBytes, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// ListShipmentsPage fetches one page of shipments from Turvo.
func (c *Client) ListShipmentsPage(ctx context.Context, start, pageSize int) ([]Shipment, struct {
	Start, PageSize, TotalRecordsInPage int
	MoreAvailable                       bool
	LastObjectKey                       interface{}
}, error) {
	var pagination struct {
		Start              int
		PageSize           int
		TotalRecordsInPage int
		MoreAvailable      bool
		LastObjectKey      interface{}
	}
	path := fmt.Sprintf("shipments/list?start=%d&pageSize=%d", start, pageSize)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, pagination, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, pagination, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.fetchToken(ctx, true); err == nil {
			return c.ListShipmentsPage(ctx, start, pageSize)
		}
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, pagination, fmt.Errorf("failed to list shipments: %s - %s", resp.Status, string(b))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pagination, err
	}
	var wrapped struct {
		Status  string `json:"Status"`
		Details struct {
			Shipments  []Shipment `json:"shipments"`
			Pagination struct {
				Start              int         `json:"start"`
				PageSize           int         `json:"pageSize"`
				TotalRecordsInPage int         `json:"totalRecordsInPage"`
				MoreAvailable      bool        `json:"moreAvailable"`
				LastObjectKey      interface{} `json:"lastObjectKey"`
			} `json:"pagination"`
		} `json:"details"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapped); err == nil && wrapped.Details.Shipments != nil {
		pagination.Start = wrapped.Details.Pagination.Start
		pagination.PageSize = wrapped.Details.Pagination.PageSize
		pagination.TotalRecordsInPage = wrapped.Details.Pagination.TotalRecordsInPage
		pagination.MoreAvailable = wrapped.Details.Pagination.MoreAvailable
		pagination.LastObjectKey = wrapped.Details.Pagination.LastObjectKey
		return wrapped.Details.Shipments, pagination, nil
	}
	var shipments []Shipment
	if err := json.Unmarshal(bodyBytes, &shipments); err != nil {
		return nil, pagination, err
	}
	pagination.Start = start
	pagination.PageSize = len(shipments)
	pagination.TotalRecordsInPage = len(shipments)
	pagination.MoreAvailable = false
	return shipments, pagination, nil
}

// ListShipments fetches all shipments by paging until completion.
func (c *Client) ListShipments(ctx context.Context) ([]Shipment, error) {
	var all []Shipment
	start := 0
	pageSize := 100
	maxPages := 100
	for page := 0; page < maxPages; page++ {
		items, meta, err := c.ListShipmentsPage(ctx, start, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if !meta.MoreAvailable {
			break
		}
		incr := meta.TotalRecordsInPage
		if incr <= 0 {
			incr = len(items)
		}
		if incr <= 0 {
			break
		}
		start += incr
	}
	log.Println("Shipments listed from Turvo:", len(all))
	return all, nil
}

// GetShipment fetches a shipment by ID.
func (c *Client) GetShipment(ctx context.Context, id string) (*Shipment, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("shipments/%s", id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.fetchToken(ctx, true); err == nil {
			return c.GetShipment(ctx, id)
		}
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get shipment: %s - %s", resp.Status, string(b))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Try direct shipment first
	var shipment Shipment
	if err := json.Unmarshal(bodyBytes, &shipment); err == nil && (shipment.ID != 0 || shipment.CustomID != "") {
		return &shipment, nil
	}
	// Fallback to wrapped structure
	var wrapped struct {
		Status  string `json:"Status"`
		Details struct {
			Shipment  *Shipment  `json:"shipment"`
			Shipments []Shipment `json:"shipments"`
		} `json:"details"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Details.Shipment != nil {
		return wrapped.Details.Shipment, nil
	}
	if len(wrapped.Details.Shipments) > 0 {
		s := wrapped.Details.Shipments[0]
		return &s, nil
	}
	return nil, fmt.Errorf("empty shipment response")
}

// CreateShipment creates a shipment in Turvo.
func (c *Client) CreateShipment(ctx context.Context, shipment Shipment) (*Shipment, error) {
	payload, err := json.Marshal(shipment)
	if err != nil {
		return nil, err
	}
	log.Printf("Turvo create payload: %s", string(payload))
	req, err := c.newRequest(ctx, http.MethodPost, "shipments?fullResponse=true", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.fetchToken(ctx, true); err == nil {
			return c.CreateShipment(ctx, shipment)
		}
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("Turvo create failed: %s - %s", resp.Status, string(bodyBytes))
		log.Printf("Request URL: %s", resp.Request.URL.String())
		log.Printf("Request Body: %s", string(payload))
		log.Printf("Request Headers: %+v", resp.Request.Header)
		return nil, fmt.Errorf("failed to create shipment: %s - %s", resp.Status, string(bodyBytes))
	}
	// Try wrapped response first
	var wrapped struct {
		Status  string          `json:"Status"`
		Details json.RawMessage `json:"details"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapped); err == nil && len(wrapped.Details) > 0 {
		var created Shipment
		if err := json.Unmarshal(wrapped.Details, &created); err == nil && (created.ID != 0 || created.CustomID != "") {

			return &created, nil
		}
	}
	// Fallback to direct shipment decode
	var created Shipment
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		return nil, fmt.Errorf("create decode error: %w", err)
	}
	return &created, nil
}

// FindShipmentByExternalID lists shipments and filters by CustomID as an external reference.
func (c *Client) FindShipmentByExternalID(ctx context.Context, externalID string) (*Shipment, error) {
	shipments, err := c.ListShipments(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range shipments {
		if s.CustomID == externalID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("shipment not found for external id %s", externalID)
}

// ListShipmentsPageWithQuery fetches one page with additional filters.
func (c *Client) ListShipmentsPageWithQuery(ctx context.Context, q url.Values) ([]Shipment, struct {
	Start, PageSize, TotalRecordsInPage int
	MoreAvailable                       bool
	LastObjectKey                       interface{}
}, error) {
	// Ensure start/pageSize exist
	if q == nil {
		q = url.Values{}
	}
	if _, ok := q["start"]; !ok {
		q.Set("start", "0")
	}
	if _, ok := q["pageSize"]; !ok {
		q.Set("pageSize", "50")
	}
	path := "shipments/list?" + q.Encode()
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	var pagination struct {
		Start, PageSize, TotalRecordsInPage int
		MoreAvailable                       bool
		LastObjectKey                       interface{}
	}
	if err != nil {
		return nil, pagination, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, pagination, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.fetchToken(ctx, true); err == nil {
			return c.ListShipmentsPageWithQuery(ctx, q)
		}
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, pagination, fmt.Errorf("failed to list shipments: %s - %s", resp.Status, string(b))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pagination, err
	}
	var wrapped struct {
		Status  string `json:"Status"`
		Details struct {
			Shipments  []Shipment `json:"shipments"`
			Pagination struct {
				Start              int         `json:"start"`
				PageSize           int         `json:"pageSize"`
				TotalRecordsInPage int         `json:"totalRecordsInPage"`
				MoreAvailable      bool        `json:"moreAvailable"`
				LastObjectKey      interface{} `json:"lastObjectKey"`
			} `json:"pagination"`
		} `json:"details"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapped); err == nil && wrapped.Details.Shipments != nil {
		pagination.Start = wrapped.Details.Pagination.Start
		pagination.PageSize = wrapped.Details.Pagination.PageSize
		pagination.TotalRecordsInPage = wrapped.Details.Pagination.TotalRecordsInPage
		pagination.MoreAvailable = wrapped.Details.Pagination.MoreAvailable
		pagination.LastObjectKey = wrapped.Details.Pagination.LastObjectKey
		return wrapped.Details.Shipments, pagination, nil
	}
	var shipments []Shipment
	if err := json.Unmarshal(bodyBytes, &shipments); err != nil {
		return nil, pagination, err
	}
	pagination.Start = atoiOrZero(q.Get("start"))
	pagination.PageSize = len(shipments)
	pagination.TotalRecordsInPage = len(shipments)
	pagination.MoreAvailable = false
	return shipments, pagination, nil
}

func atoiOrZero(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
