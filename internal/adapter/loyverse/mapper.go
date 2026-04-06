package loyverse

import (
	"fmt"
	"strings"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
)

// MapItem converts a Loyverse Item into a FireLine NormalizedMenuItem.
// The primary variant's price is used; if there are multiple variants the
// caller should call MapItemVariant per-variant instead.
func MapItem(item Item, categoryName string) adapter.NormalizedMenuItem {
	var price float64
	if len(item.Variants) > 0 {
		price = item.Variants[0].DefaultPrice
	}

	return adapter.NormalizedMenuItem{
		ExternalID:  item.ID,
		Name:        item.ItemName,
		Description: item.Description,
		Category:    categoryName,
		Price:       centsFromDecimal(price),
		Available:   true, // availability is driven by inventory; assume available on sync
		Source:      "loyverse",
	}
}

// MapItemVariant converts a Loyverse Item + Variant pair into a NormalizedMenuItem.
// Useful when an item has multiple purchasable variants (sizes, etc.).
func MapItemVariant(item Item, v Variant, categoryName string) adapter.NormalizedMenuItem {
	name := item.ItemName
	if v.Option1 != "" {
		name = fmt.Sprintf("%s - %s", item.ItemName, v.Option1)
	}
	return adapter.NormalizedMenuItem{
		ExternalID:  v.VariantID,
		Name:        name,
		Description: item.Description,
		Category:    categoryName,
		Price:       centsFromDecimal(v.DefaultPrice),
		Available:   true,
		Source:      "loyverse",
	}
}

// MapReceipt converts a Loyverse Receipt into a FireLine NormalizedOrder.
func MapReceipt(receipt Receipt) adapter.NormalizedOrder {
	openedAt, _ := time.Parse(time.RFC3339, receipt.CreatedAt)
	closedAt := openedAt // Loyverse receipts are always closed at creation time
	closedAtPtr := &closedAt

	items := make([]adapter.NormalizedOrderItem, 0, len(receipt.LineItems))
	for _, li := range receipt.LineItems {
		items = append(items, MapLineItem(li))
	}

	channel := mapDiningOption(receipt.DiningOption)

	status := "closed"
	if strings.EqualFold(receipt.ReceiptType, "REFUND") {
		status = "voided"
	}

	return adapter.NormalizedOrder{
		ExternalID:  receipt.ReceiptNumber,
		OrderNumber: receipt.ReceiptNumber,
		Status:      status,
		Channel:     channel,
		Items:       items,
		Subtotal:    centsFromDecimal(receipt.TotalMoney - receipt.TotalTax),
		Tax:         centsFromDecimal(receipt.TotalTax),
		Discount:    centsFromDecimal(receipt.TotalDiscount),
		Total:       centsFromDecimal(receipt.TotalMoney),
		OpenedAt:    openedAt,
		ClosedAt:    closedAtPtr,
		Source:      "loyverse",
	}
}

// MapLineItem converts a Loyverse LineItem into a NormalizedOrderItem.
func MapLineItem(li LineItem) adapter.NormalizedOrderItem {
	name := li.ItemName
	if li.VariantName != "" && li.VariantName != li.ItemName {
		name = fmt.Sprintf("%s - %s", li.ItemName, li.VariantName)
	}
	return adapter.NormalizedOrderItem{
		ExternalID: li.VariantID,
		MenuItemID: li.ItemID,
		Name:       name,
		Quantity:   int(li.Quantity),
		UnitPrice:  centsFromDecimal(li.Price),
	}
}

// MapEmployee converts a Loyverse Employee into a FireLine NormalizedEmployee.
func MapEmployee(emp Employee) adapter.NormalizedEmployee {
	firstName, lastName := splitName(emp.Name)
	return adapter.NormalizedEmployee{
		ExternalID: emp.ID,
		FirstName:  firstName,
		LastName:   lastName,
		Role:       normalizeRole(emp.Role),
		Active:     !emp.Active, // emp.Active reads json:"is_deleted", so invert
		Source:     "loyverse",
	}
}

// mapDiningOption converts Loyverse dining option strings to FireLine channel names.
func mapDiningOption(opt string) string {
	switch strings.ToUpper(opt) {
	case "DINE_IN":
		return "dine_in"
	case "TAKEOUT", "TAKE_OUT", "TO_GO":
		return "takeout"
	case "DELIVERY":
		return "delivery"
	default:
		return "dine_in"
	}
}

// normalizeRole maps Loyverse role strings to canonical FireLine role strings.
func normalizeRole(role string) string {
	switch strings.ToUpper(role) {
	case "OWNER", "ADMIN", "MANAGER":
		return "manager"
	case "CASHIER":
		return "cashier"
	case "BARISTA":
		return "barista"
	case "BARTENDER":
		return "bartender"
	case "WAITER", "SERVER":
		return "server"
	default:
		return strings.ToLower(role)
	}
}

// splitName splits a full name string into first/last.
func splitName(fullName string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(fullName), " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// centsFromDecimal converts a decimal currency amount to integer cents.
func centsFromDecimal(amount float64) int64 {
	// Round to nearest cent to avoid floating point drift.
	return int64(amount*100 + 0.5)
}
