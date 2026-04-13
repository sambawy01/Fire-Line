package loyverse

import (
	"testing"
	"time"
)

// --- MapItem ---

func TestMapItem_BasicItem(t *testing.T) {
	item := Item{
		ID:          "item-abc",
		ItemName:    "Chicken Shawarma",
		Description: "Tender marinated chicken with garlic sauce",
		CategoryID:  "cat-1",
		Variants: []Variant{
			{VariantID: "var-1", DefaultPrice: 8.50},
		},
	}

	mi := MapItem(item, "Wraps")

	if mi.ExternalID != "item-abc" {
		t.Errorf("ExternalID: want item-abc, got %s", mi.ExternalID)
	}
	if mi.Name != "Chicken Shawarma" {
		t.Errorf("Name: want Chicken Shawarma, got %s", mi.Name)
	}
	if mi.Category != "Wraps" {
		t.Errorf("Category: want Wraps, got %s", mi.Category)
	}
	if mi.Price != 850 {
		t.Errorf("Price: want 850 cents, got %d", mi.Price)
	}
	if !mi.Available {
		t.Error("Available: want true")
	}
	if mi.Source != "loyverse" {
		t.Errorf("Source: want loyverse, got %s", mi.Source)
	}
	if mi.Description != item.Description {
		t.Errorf("Description mismatch")
	}
}

func TestMapItem_NoVariants(t *testing.T) {
	item := Item{
		ID:       "item-no-variants",
		ItemName: "Water",
	}
	mi := MapItem(item, "Beverages")
	if mi.Price != 0 {
		t.Errorf("Price with no variants: want 0, got %d", mi.Price)
	}
}

func TestMapItem_PriceConversion(t *testing.T) {
	cases := []struct {
		price    float64
		wantCent int64
	}{
		{0.99, 99},
		{10.00, 1000},
		{12.50, 1250},
		{0.01, 1},
		{99.99, 9999},
	}
	for _, tc := range cases {
		item := Item{
			ID:       "item-x",
			ItemName: "Test Item",
			Variants: []Variant{{VariantID: "v1", DefaultPrice: tc.price}},
		}
		mi := MapItem(item, "Test")
		if mi.Price != tc.wantCent {
			t.Errorf("price %.2f: want %d cents, got %d", tc.price, tc.wantCent, mi.Price)
		}
	}
}

// --- MapItemVariant ---

func TestMapItemVariant_WithOption(t *testing.T) {
	item := Item{ID: "item-1", ItemName: "Coffee"}
	v := Variant{VariantID: "var-large", DefaultPrice: 5.0, Option1: "Large"}

	mi := MapItemVariant(item, v, "Hot Drinks")

	if mi.ExternalID != "var-large" {
		t.Errorf("ExternalID: want var-large, got %s", mi.ExternalID)
	}
	if mi.Name != "Coffee - Large" {
		t.Errorf("Name: want 'Coffee - Large', got %s", mi.Name)
	}
	if mi.Price != 500 {
		t.Errorf("Price: want 500, got %d", mi.Price)
	}
}

func TestMapItemVariant_NoOption(t *testing.T) {
	item := Item{ID: "item-1", ItemName: "Espresso"}
	v := Variant{VariantID: "var-1", DefaultPrice: 3.0}

	mi := MapItemVariant(item, v, "Coffee")
	if mi.Name != "Espresso" {
		t.Errorf("Name: want Espresso, got %s", mi.Name)
	}
}

// --- MapReceipt ---

func TestMapReceipt_DineIn(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R001",
		ReceiptType:   "SALE",
		CreatedAt:     "2026-03-24T12:00:00Z",
		TotalMoney:    25.00,
		TotalTax:      2.00,
		TotalDiscount: 0.00,
		DiningOption:  "DINE_IN",
		LineItems: []LineItem{
			{
				ItemID:      "item-1",
				VariantID:   "var-1",
				ItemName:    "Burger",
				VariantName: "Burger",
				Quantity:    1,
				Price:       23.00,
				GrossTotal:  23.00,
			},
		},
	}

	order := MapReceipt(receipt)

	if order.ExternalID != "R001" {
		t.Errorf("ExternalID: want R001, got %s", order.ExternalID)
	}
	if order.Channel != "dine_in" {
		t.Errorf("Channel: want dine_in, got %s", order.Channel)
	}
	if order.Status != "closed" {
		t.Errorf("Status: want closed, got %s", order.Status)
	}
	if order.Total != 2500 {
		t.Errorf("Total: want 2500 cents, got %d", order.Total)
	}
	if order.Tax != 200 {
		t.Errorf("Tax: want 200 cents, got %d", order.Tax)
	}
	if order.Source != "loyverse" {
		t.Errorf("Source: want loyverse, got %s", order.Source)
	}
	if len(order.Items) != 1 {
		t.Fatalf("Items count: want 1, got %d", len(order.Items))
	}
	if order.Items[0].Name != "Burger" {
		t.Errorf("Item name: want Burger, got %s", order.Items[0].Name)
	}
	if order.Items[0].UnitPrice != 2300 {
		t.Errorf("Item price: want 2300, got %d", order.Items[0].UnitPrice)
	}
	if order.ClosedAt == nil {
		t.Error("ClosedAt should not be nil")
	}
}

func TestMapReceipt_Takeout(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R002",
		ReceiptType:   "SALE",
		CreatedAt:     "2026-03-24T13:00:00Z",
		DiningOption:  "TAKEOUT",
	}
	order := MapReceipt(receipt)
	if order.Channel != "takeout" {
		t.Errorf("Channel: want takeout, got %s", order.Channel)
	}
}

func TestMapReceipt_Delivery(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R003",
		ReceiptType:   "SALE",
		CreatedAt:     "2026-03-24T14:00:00Z",
		DiningOption:  "DELIVERY",
	}
	order := MapReceipt(receipt)
	if order.Channel != "delivery" {
		t.Errorf("Channel: want delivery, got %s", order.Channel)
	}
}

func TestMapReceipt_Refund(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R004",
		ReceiptType:   "REFUND",
		CreatedAt:     "2026-03-24T15:00:00Z",
	}
	order := MapReceipt(receipt)
	if order.Status != "voided" {
		t.Errorf("Status: want voided for REFUND, got %s", order.Status)
	}
}

func TestMapReceipt_TimestampParsed(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R005",
		ReceiptType:   "SALE",
		CreatedAt:     "2026-03-24T09:30:00Z",
	}
	order := MapReceipt(receipt)
	want := time.Date(2026, 3, 24, 9, 30, 0, 0, time.UTC)
	if !order.OpenedAt.Equal(want) {
		t.Errorf("OpenedAt: want %v, got %v", want, order.OpenedAt)
	}
}

func TestMapReceipt_Discount(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R006",
		ReceiptType:   "SALE",
		CreatedAt:     "2026-03-24T10:00:00Z",
		TotalMoney:    18.00,
		TotalDiscount: 2.00,
	}
	order := MapReceipt(receipt)
	if order.Discount != 200 {
		t.Errorf("Discount: want 200 cents, got %d", order.Discount)
	}
}

func TestMapReceipt_MultipleLineItems(t *testing.T) {
	receipt := Receipt{
		ReceiptNumber: "R007",
		ReceiptType:   "SALE",
		CreatedAt:     "2026-03-24T11:00:00Z",
		TotalMoney:    30.00,
		LineItems: []LineItem{
			{ItemID: "i1", VariantID: "v1", ItemName: "Pizza", Quantity: 1, Price: 15.00},
			{ItemID: "i2", VariantID: "v2", ItemName: "Salad", Quantity: 2, Price: 7.50},
		},
	}
	order := MapReceipt(receipt)
	if len(order.Items) != 2 {
		t.Fatalf("Items count: want 2, got %d", len(order.Items))
	}
	if order.Items[1].Quantity != 2 {
		t.Errorf("Second item quantity: want 2, got %d", order.Items[1].Quantity)
	}
}

// --- MapEmployee ---

func TestMapEmployee_Basic(t *testing.T) {
	emp := Employee{
		ID:    "emp-1",
		Name:  "Amal Hassan",
		Email: "amal@example.com",
		// DeletedAt nil → active. IsOwner false → role "staff".
	}
	ne := MapEmployee(emp)

	if ne.ExternalID != "emp-1" {
		t.Errorf("ExternalID: want emp-1, got %s", ne.ExternalID)
	}
	if ne.FirstName != "Amal" {
		t.Errorf("FirstName: want Amal, got %s", ne.FirstName)
	}
	if ne.LastName != "Hassan" {
		t.Errorf("LastName: want Hassan, got %s", ne.LastName)
	}
	if ne.Role != "staff" {
		t.Errorf("Role: want staff, got %s", ne.Role)
	}
	if !ne.Active {
		t.Error("Active: want true (deleted_at=nil)")
	}
	if ne.Source != "loyverse" {
		t.Errorf("Source: want loyverse, got %s", ne.Source)
	}
}

func TestMapEmployee_Owner(t *testing.T) {
	emp := Employee{ID: "emp-owner", Name: "Bistro Owner", IsOwner: true}
	ne := MapEmployee(emp)
	if ne.Role != "owner" {
		t.Errorf("Role: want owner when is_owner=true, got %s", ne.Role)
	}
}

func TestMapEmployee_Deleted(t *testing.T) {
	ts := "2026-01-15T09:00:00Z"
	emp := Employee{
		ID:        "emp-2",
		Name:      "Old Employee",
		DeletedAt: &ts,
	}
	ne := MapEmployee(emp)
	if ne.Active {
		t.Error("Active: want false when deleted_at is set")
	}
}

func TestMapEmployee_SingleName(t *testing.T) {
	emp := Employee{ID: "emp-3", Name: "Madonna"}
	ne := MapEmployee(emp)
	if ne.FirstName != "Madonna" {
		t.Errorf("FirstName: want Madonna, got %s", ne.FirstName)
	}
	if ne.LastName != "" {
		t.Errorf("LastName: want empty, got %s", ne.LastName)
	}
}

// --- mapDiningOption ---

func TestMapDiningOption(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"DINE_IN", "dine_in"},
		{"TAKEOUT", "takeout"},
		{"TAKE_OUT", "takeout"},
		{"TO_GO", "takeout"},
		{"DELIVERY", "delivery"},
		{"", "dine_in"},
		{"UNKNOWN", "dine_in"},
	}
	for _, tc := range cases {
		got := mapDiningOption(tc.input)
		if got != tc.want {
			t.Errorf("mapDiningOption(%q): want %q, got %q", tc.input, tc.want, got)
		}
	}
}

// --- centsFromDecimal ---

func TestCentsFromDecimal(t *testing.T) {
	cases := []struct {
		in   float64
		want int64
	}{
		{0, 0},
		{1.0, 100},
		{0.5, 50},
		{9.99, 999},
		{100.00, 10000},
	}
	for _, tc := range cases {
		got := centsFromDecimal(tc.in)
		if got != tc.want {
			t.Errorf("centsFromDecimal(%.2f): want %d, got %d", tc.in, tc.want, got)
		}
	}
}
