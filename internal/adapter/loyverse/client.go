package loyverse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.loyverse.com/v1.0"

// Client is an HTTP client for the Loyverse REST API.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewClient creates a new Loyverse API client with the given bearer token.
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: defaultBaseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// get performs an authenticated GET request and JSON-decodes the response body.
func (c *Client) get(path string, query url.Values, out any) error {
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("loyverse: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("loyverse: execute request %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loyverse: %s returned %d: %s", path, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("loyverse: decode response from %s: %w", path, err)
	}
	return nil
}

// put performs an authenticated PUT request with a JSON body.
func (c *Client) put(path string, body any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("loyverse: marshal PUT body for %s: %w", path, err)
	}

	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("loyverse: build PUT request for %s: %w", path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("loyverse: execute PUT %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("loyverse: PUT %s returned %d: %s", path, resp.StatusCode, string(respBody))
	}
	return nil
}

// GetItems fetches a page of menu items. Pass an empty cursor for the first page.
func (c *Client) GetItems(cursor string) (*ItemsResponse, error) {
	q := url.Values{}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var resp ItemsResponse
	if err := c.get("/items", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetReceipts fetches a page of receipts created after `since`.
// Pass an empty cursor for the first page.
func (c *Client) GetReceipts(since time.Time, cursor string) (*ReceiptsResponse, error) {
	q := url.Values{}
	q.Set("created_at_min", since.UTC().Format(time.RFC3339))
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var resp ReceiptsResponse
	if err := c.get("/receipts", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEmployees fetches all employees.
func (c *Client) GetEmployees() (*EmployeesResponse, error) {
	var resp EmployeesResponse
	if err := c.get("/employees", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetInventory fetches stock levels for the given store.
func (c *Client) GetInventory(storeID string) (*InventoryResponse, error) {
	q := url.Values{}
	if storeID != "" {
		q.Set("store_id", storeID)
	}
	var resp InventoryResponse
	if err := c.get("/inventory", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCustomers fetches a page of customer profiles.
func (c *Client) GetCustomers(cursor string) (*CustomersResponse, error) {
	q := url.Values{}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var resp CustomersResponse
	if err := c.get("/customers", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCategories fetches all menu categories.
func (c *Client) GetCategories() (*CategoriesResponse, error) {
	var resp CategoriesResponse
	if err := c.get("/categories", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetStores fetches all stores/locations.
func (c *Client) GetStores() (*StoresResponse, error) {
	var resp StoresResponse
	if err := c.get("/stores", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// inventoryUpdateBody is the request payload for PUT /inventory.
type inventoryUpdateBody struct {
	InventoryLevels []inventoryUpdateLevel `json:"inventory_levels"`
}

type inventoryUpdateLevel struct {
	VariantID string  `json:"variant_id"`
	StoreID   string  `json:"store_id"`
	InStock   float64 `json:"in_stock"`
}

// UpdateVariantStock sets the stock level for a variant at a store.
// Pass inStock=0 to 86 an item.
func (c *Client) UpdateVariantStock(variantID, storeID string, inStock float64) error {
	body := inventoryUpdateBody{
		InventoryLevels: []inventoryUpdateLevel{
			{VariantID: variantID, StoreID: storeID, InStock: inStock},
		},
	}
	return c.put("/inventory", body)
}
