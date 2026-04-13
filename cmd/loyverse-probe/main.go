// Command loyverse-probe exercises the Loyverse adapter end-to-end without
// the rest of the fireline server. It calls Initialize, then pulls menu,
// orders, and employees and prints a summary.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/opsnerve/fireline/internal/adapter"
	"github.com/opsnerve/fireline/internal/adapter/loyverse"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if os.Getenv("LOYVERSE_API_TOKEN") == "" {
		fmt.Fprintln(os.Stderr, "LOYVERSE_API_TOKEN not set")
		os.Exit(1)
	}

	a := loyverse.New()

	// Use an empty cfg so background polling has a long interval and doesn't
	// spam; we'll call the reader methods directly.
	cfg := adapter.Config{
		OrgID:        "probe-org",
		LocationID:   "probe-location",
		PollInterval: 60 * time.Minute,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := a.Initialize(ctx, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "initialize:", err)
		os.Exit(1)
	}
	defer a.Shutdown(context.Background())

	fmt.Println("\n=== HealthCheck ===")
	if err := a.HealthCheck(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		os.Exit(1)
	}
	fmt.Println("OK — status:", a.GetStatus())

	// Menu
	fmt.Println("\n=== ReadMenu ===")
	menuReader := a.(adapter.MenuReader)
	items, err := menuReader.ReadMenu(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read menu:", err)
	} else {
		fmt.Printf("menu items: %d\n", len(items))
		catCount := 0
		noCatCount := 0
		for _, it := range items {
			if it.Category == "" {
				noCatCount++
			} else {
				catCount++
			}
		}
		fmt.Printf("  with category:    %d\n  without category: %d\n", catCount, noCatCount)
		fmt.Println("  sample categorized items:")
		shown := 0
		for _, it := range items {
			if it.Category == "" || shown >= 5 {
				continue
			}
			fmt.Printf("    - %-40s  cat=%-25s  price=%d cents\n", it.Name, it.Category, it.Price)
			shown++
		}
	}

	// Orders (last 30 days)
	fmt.Println("\n=== ReadOrders (last 30 days) ===")
	orderReader := a.(adapter.OrderReader)
	since := time.Now().AddDate(0, 0, -30)
	orders, err := orderReader.ReadOrders(ctx, since, 1000)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read orders:", err)
	} else {
		fmt.Printf("orders: %d (since %s)\n", len(orders), since.Format(time.RFC3339))
		for i, o := range orders {
			if i >= 5 {
				fmt.Printf("... (%d more)\n", len(orders)-5)
				break
			}
			fmt.Printf("  - %s  total=%d cents  items=%d  opened=%s\n",
				o.ExternalID, o.Total, len(o.Items), o.OpenedAt.Format(time.RFC3339))
		}
	}

	// Employees
	fmt.Println("\n=== ReadEmployees ===")
	empReader := a.(adapter.EmployeeReader)
	emps, err := empReader.ReadEmployees(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read employees:", err)
	} else {
		fmt.Printf("employees: %d\n", len(emps))
		for i, e := range emps {
			if i >= 5 {
				fmt.Printf("... (%d more)\n", len(emps)-5)
				break
			}
			fmt.Printf("  - %s  name=%s %s  role=%s  active=%v\n", e.ExternalID, e.FirstName, e.LastName, e.Role, e.Active)
		}
	}

	fmt.Println("\n=== Freshness ===")
	for _, dt := range []string{"menu", "orders", "employees"} {
		if f, err := a.GetDataFreshness(dt); err == nil {
			fmt.Printf("  %-10s  last_sync=%s  records=%d\n", dt, f.LastSyncAt.Format(time.RFC3339), f.RecordCount)
		}
	}
}
