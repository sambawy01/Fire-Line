package loyverse

// Item represents a Loyverse menu item.
type Item struct {
	ID          string    `json:"id"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	CategoryID  string    `json:"category_id"`
	Variants    []Variant `json:"variants"`
	IsComposite bool      `json:"is_composite"`
}

// Variant represents a single SKU/price variant of a Loyverse item.
type Variant struct {
	VariantID    string  `json:"variant_id"`
	SKU          string  `json:"sku"`
	DefaultPrice float64 `json:"default_price"` // decimal currency units
	Cost         float64 `json:"cost"`
	Option1      string  `json:"option1"`
	Option2      string  `json:"option2"`
	Option3      string  `json:"option3"`
}

// Category represents a Loyverse menu category.
type Category struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Receipt represents a Loyverse sales transaction.
type Receipt struct {
	ReceiptNumber string     `json:"receipt_number"`
	ReceiptType   string     `json:"receipt_type"` // SALE, REFUND
	CreatedAt     string     `json:"created_at"`
	TotalMoney    float64    `json:"total_money"`
	TotalTax      float64    `json:"total_tax"`
	TotalDiscount float64    `json:"total_discount"`
	PointsOfSale  []PointOfSale `json:"points_of_sale"`
	LineItems     []LineItem `json:"line_items"`
	Payments      []Payment  `json:"payments"`
	EmployeeID    string     `json:"employee_id"`
	StoreID       string     `json:"store_id"`
	CustomerID    string     `json:"customer_id"`
	DiningOption  string     `json:"dining_option"` // DINE_IN, TAKEOUT, DELIVERY
	Note          string     `json:"note"`
	OrderSource   string     `json:"order_source"`
}

// PointOfSale is used for nested receipt structure.
type PointOfSale struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LineItem is a single line within a receipt.
type LineItem struct {
	ItemID        string  `json:"item_id"`
	VariantID     string  `json:"variant_id"`
	ItemName      string  `json:"item_name"`
	VariantName   string  `json:"variant_name"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	GrossTotal    float64 `json:"gross_total_money"`
	TotalDiscount float64 `json:"total_discount"`
	TotalTax      float64 `json:"total_tax_money"`
	LineNote      string  `json:"line_note"`
}

// Payment is a single payment within a receipt.
type Payment struct {
	PaymentTypeID string  `json:"payment_type_id"`
	Name          string  `json:"name"`
	Amount        float64 `json:"money_amount"`
}

// Employee represents a Loyverse staff member.
type Employee struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	Active      bool   `json:"is_deleted"` // Loyverse uses is_deleted; inverted
	StoreID     string `json:"store_id"`
}

// Customer represents a Loyverse customer profile.
type Customer struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	PhoneNumber    string  `json:"phone_number"`
	LoyaltyPoints  float64 `json:"total_points"`
	TotalVisits    int     `json:"total_visits"`
	TotalSpend     float64 `json:"total_money_balance"`
	CustomerCode   string  `json:"customer_code"`
	CreatedAt      string  `json:"created_at"`
	Note           string  `json:"note"`
}

// Store represents a Loyverse location/store.
type Store struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
	Country     string `json:"country_code"`
	Currency    string `json:"currency_code"`
}

// InventoryLevel represents a stock level for a variant at a store.
type InventoryLevel struct {
	VariantID      string  `json:"variant_id"`
	StoreID        string  `json:"store_id"`
	InStock        float64 `json:"in_stock"`
	LowStockLevel  float64 `json:"low_stock"`
}

// --- Paginated response envelopes ---

// ItemsResponse is the paginated response for GET /items.
type ItemsResponse struct {
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

// ReceiptsResponse is the paginated response for GET /receipts.
type ReceiptsResponse struct {
	Receipts []Receipt `json:"receipts"`
	Cursor   string    `json:"cursor"`
}

// EmployeesResponse is the response for GET /employees.
type EmployeesResponse struct {
	Employees []Employee `json:"employees"`
}

// CustomersResponse is the paginated response for GET /customers.
type CustomersResponse struct {
	Customers []Customer `json:"customers"`
	Cursor    string     `json:"cursor"`
}

// CategoriesResponse is the response for GET /categories.
type CategoriesResponse struct {
	Categories []Category `json:"categories"`
}

// StoresResponse is the response for GET /stores.
type StoresResponse struct {
	Stores []Store `json:"stores"`
}

// InventoryResponse is the response for GET /inventory.
type InventoryResponse struct {
	InventoryLevels []InventoryLevel `json:"inventory_levels"`
}
