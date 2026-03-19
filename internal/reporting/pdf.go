package reporting

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// GeneratePDF renders a DailyReport as a landscape A4 PDF and returns the bytes.
func (s *Service) GeneratePDF(report *DailyReport) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	contentW := pageW - 24 // left + right margins

	// ─── Fonts ───────────────────────────────────────────────────────────────
	pdf.SetFont("Helvetica", "B", 18)

	// ─── Header ──────────────────────────────────────────────────────────────
	pdf.SetFillColor(30, 30, 30)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(contentW, 10, "FireLine Daily Report  |  OpsNerve", "", 0, "L", true, 0, "")
	pdf.Ln(12)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(contentW/2, 7, report.LocationName, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.CellFormat(contentW/2, 7, "Report Date: "+report.ReportDate, "", 1, "R", false, 0, "")
	pdf.Ln(3)

	// ─── Health Score ─────────────────────────────────────────────────────────
	score := report.HealthScore
	var r, g, b int
	var label string
	switch {
	case score >= 70:
		r, g, b = 22, 163, 74 // green
		label = "GOOD"
	case score >= 40:
		r, g, b = 202, 138, 4 // amber
		label = "FAIR"
	default:
		r, g, b = 220, 38, 38 // red
		label = "POOR"
	}
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(r, g, b)
	pdf.CellFormat(contentW, 8,
		fmt.Sprintf("Health Score: %d / 100  (%s)", score, label),
		"", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(4)

	// ─── KPI Table ───────────────────────────────────────────────────────────
	sectionHeader(pdf, contentW, "Key Performance Indicators")

	col := contentW / 7
	headers := []string{"Net Revenue", "Gross Margin", "Labor Cost %", "Orders", "Avg Ticket", "Active Alerts", "Critical"}
	values := []string{
		formatCents(report.NetRevenue),
		fmt.Sprintf("%.1f%%", report.GrossMarginPct),
		fmt.Sprintf("%.1f%%", report.LaborCostPct),
		fmt.Sprintf("%d", report.OrdersToday),
		fmt.Sprintf("%.1f min", report.AvgTicketTime),
		fmt.Sprintf("%d", report.ActiveAlerts),
		fmt.Sprintf("%d", report.CriticalCount),
	}

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(50, 50, 50)
	pdf.SetTextColor(255, 255, 255)
	for _, h := range headers {
		pdf.CellFormat(col, 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(0, 0, 0)
	for _, v := range values {
		pdf.CellFormat(col, 8, v, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(6)

	// ─── Critical Issues ─────────────────────────────────────────────────────
	if len(report.CriticalIssues) > 0 {
		sectionHeader(pdf, contentW, "Critical Issues")
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(220, 38, 38)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(contentW*0.25, 7, "Title", "1", 0, "L", true, 0, "")
		pdf.CellFormat(contentW*0.55, 7, "Description", "1", 0, "L", true, 0, "")
		pdf.CellFormat(contentW*0.20, 7, "Module", "1", 1, "C", true, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(0, 0, 0)
		for i, ci := range report.CriticalIssues {
			if i%2 == 0 {
				pdf.SetFillColor(255, 230, 230)
			} else {
				pdf.SetFillColor(255, 245, 245)
			}
			pdf.CellFormat(contentW*0.25, 6, ci.Title, "1", 0, "L", true, 0, "")
			pdf.CellFormat(contentW*0.55, 6, truncate(ci.Description, 80), "1", 0, "L", true, 0, "")
			pdf.CellFormat(contentW*0.20, 6, ci.Module, "1", 1, "C", true, 0, "")
		}
		pdf.Ln(4)
	}

	// ─── Channel Breakdown ───────────────────────────────────────────────────
	if len(report.Channels) > 0 {
		sectionHeader(pdf, contentW, "Channel Breakdown")
		chCol := contentW / 4
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(50, 50, 50)
		pdf.SetTextColor(255, 255, 255)
		for _, h := range []string{"Channel", "Orders", "Revenue", "Avg Ticket (min)"} {
			pdf.CellFormat(chCol, 7, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(0, 0, 0)
		for i, ch := range report.Channels {
			if i%2 == 0 {
				pdf.SetFillColor(245, 245, 245)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(chCol, 6, ch.Channel, "1", 0, "L", true, 0, "")
			pdf.CellFormat(chCol, 6, fmt.Sprintf("%d", ch.Orders), "1", 0, "C", true, 0, "")
			pdf.CellFormat(chCol, 6, formatCents(ch.Revenue), "1", 0, "R", true, 0, "")
			pdf.CellFormat(chCol, 6, fmt.Sprintf("%.1f", ch.AvgTicket), "1", 1, "C", true, 0, "")
		}
		pdf.Ln(4)
	}

	// ─── Top Menu Items ──────────────────────────────────────────────────────
	sectionHeader(pdf, contentW, "Menu Performance")
	miCol := contentW / 4

	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetFillColor(50, 50, 50)
	pdf.SetTextColor(255, 255, 255)
	for _, h := range []string{"Item", "Category", "Units Sold", "Revenue"} {
		pdf.CellFormat(miCol, 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(0, 0, 0)

	printMenuItem := func(item ReportMenuItem, label string, fill [3]int) {
		pdf.SetFillColor(fill[0], fill[1], fill[2])
		prefix := ""
		if label != "" {
			prefix = "[" + label + "] "
		}
		pdf.CellFormat(miCol, 6, prefix+item.Name, "1", 0, "L", true, 0, "")
		pdf.CellFormat(miCol, 6, item.Category, "1", 0, "C", true, 0, "")
		pdf.CellFormat(miCol, 6, fmt.Sprintf("%d", item.Units), "1", 0, "C", true, 0, "")
		pdf.CellFormat(miCol, 6, formatCents(item.Revenue), "1", 1, "R", true, 0, "")
	}

	for i, item := range report.TopItems {
		fill := [3]int{240, 253, 244} // light green
		if i == 1 {
			fill = [3]int{245, 245, 245}
		} else if i == 2 {
			fill = [3]int{255, 255, 255}
		}
		printMenuItem(item, map[int]string{0: "Top 1", 1: "Top 2", 2: "Top 3"}[i], fill)
	}
	if report.WorstItem != nil {
		printMenuItem(*report.WorstItem, "Worst", [3]int{255, 237, 237})
	}
	pdf.Ln(4)

	// ─── Category Revenue ────────────────────────────────────────────────────
	if len(report.CategoryRevenue) > 0 {
		sectionHeader(pdf, contentW, "Revenue by Category")
		catCol := contentW / 3
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(50, 50, 50)
		pdf.SetTextColor(255, 255, 255)
		for _, h := range []string{"Category", "Units", "Revenue"} {
			pdf.CellFormat(catCol, 7, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(0, 0, 0)
		for i, cat := range report.CategoryRevenue {
			if i%2 == 0 {
				pdf.SetFillColor(245, 245, 245)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(catCol, 6, cat.Category, "1", 0, "L", true, 0, "")
			pdf.CellFormat(catCol, 6, fmt.Sprintf("%d", cat.Units), "1", 0, "C", true, 0, "")
			pdf.CellFormat(catCol, 6, formatCents(cat.Revenue), "1", 1, "R", true, 0, "")
		}
		pdf.Ln(4)
	}

	// ─── Staff Summary ────────────────────────────────────────────────────────
	if len(report.StaffSummary) > 0 {
		sectionHeader(pdf, contentW, fmt.Sprintf("Staff Summary  |  Total Hours: %.1f  |  Total Labor Cost: %s",
			report.TotalHoursWorked, formatCents(report.TotalLaborCost)))

		staffCol := contentW / 4
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(50, 50, 50)
		pdf.SetTextColor(255, 255, 255)
		for _, h := range []string{"Name", "Role", "Hours", "Cost"} {
			pdf.CellFormat(staffCol, 7, h, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(0, 0, 0)

		overtimeSet := map[string]bool{}
		for _, name := range report.OvertimeFlags {
			overtimeSet[name] = true
		}

		for _, entry := range report.StaffSummary {
			if overtimeSet[entry.Name] {
				pdf.SetFillColor(255, 243, 199) // amber for overtime
			} else {
				pdf.SetFillColor(245, 245, 245)
			}
			overtimeSuffix := ""
			if overtimeSet[entry.Name] {
				overtimeSuffix = " *** OT"
			}
			pdf.CellFormat(staffCol, 6, entry.Name+overtimeSuffix, "1", 0, "L", true, 0, "")
			pdf.CellFormat(staffCol, 6, entry.Role, "1", 0, "C", true, 0, "")
			pdf.CellFormat(staffCol, 6, fmt.Sprintf("%.2f", entry.Hours), "1", 0, "C", true, 0, "")
			pdf.CellFormat(staffCol, 6, formatCents(entry.Cost), "1", 1, "R", true, 0, "")
		}
		pdf.Ln(4)
	}

	// ─── Footer ───────────────────────────────────────────────────────────────
	pdf.SetY(-14)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(contentW, 5,
		fmt.Sprintf("Generated by FireLine by OpsNerve  |  %s  |  %s", report.LocationName, report.ReportDate),
		"", 0, "C", false, 0, "")

	// ─── Write to buffer ──────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	return buf.Bytes(), nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func sectionHeader(pdf *fpdf.Fpdf, w float64, title string) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(70, 70, 70)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(w, 7, "  "+title, "", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(1)
}

// formatCents converts a cents int64 to a dollar string like "$1,234.56".
func formatCents(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	dollars := cents / 100
	pennies := cents % 100

	// Thousands-separate the dollars
	s := fmt.Sprintf("%d", dollars)
	result := ""
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(ch)
	}
	out := fmt.Sprintf("$%s.%02d", result, pennies)
	if negative {
		out = "-" + out
	}
	return out
}

// truncate shortens a string to max runes, appending "…" if trimmed.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
