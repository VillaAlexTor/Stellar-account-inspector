package horizon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiClient    *http.Client
	streamClient *http.Client
	testnetURL   string
}

func NewClient(testnetURL string, timeout time.Duration) *Client {
	return &Client{
		apiClient:    &http.Client{Timeout: timeout},
		streamClient: &http.Client{},
		testnetURL:   strings.TrimRight(testnetURL, "/"),
	}
}

func (c *Client) Account(ctx context.Context, publicKey, network string) (AccountSnapshot, error) {
	endpoint, err := c.accountURL(publicKey, network)
	if err != nil {
		return AccountSnapshot{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return AccountSnapshot{}, fmt.Errorf("crear consulta de cuenta: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := c.apiClient.Do(request)
	if err != nil {
		return AccountSnapshot{}, fmt.Errorf("consultar cuenta en Horizon: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return AccountSnapshot{}, fmt.Errorf("Horizon respondió %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var snapshot AccountSnapshot
	if err := json.NewDecoder(response.Body).Decode(&snapshot); err != nil {
		return AccountSnapshot{}, fmt.Errorf("decodificar cuenta Horizon: %w", err)
	}
	return snapshot, nil
}

func (c *Client) OperationsURL(publicKey, network, cursor string) (string, error) {
	baseURL, err := c.baseURL(network)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("cursor", cursor)
	query.Set("order", "asc")
	query.Set("limit", "200")
	return fmt.Sprintf("%s/accounts/%s/operations?%s", baseURL, url.PathEscape(publicKey), query.Encode()), nil
}

func (c *Client) accountURL(publicKey, network string) (string, error) {
	baseURL, err := c.baseURL(network)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/accounts/%s", baseURL, url.PathEscape(publicKey)), nil
}

func (c *Client) baseURL(network string) (string, error) {
	if network != "testnet" {
		return "", fmt.Errorf("red Stellar no soportada: %s", network)
	}
	return c.testnetURL, nil
}
