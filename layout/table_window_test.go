// Copyright 2026 Carlos Munoz and the Folio Authors
// SPDX-License-Identifier: Apache-2.0

package layout

import (
	"fmt"
	"testing"

	"github.com/carlos7ags/folio/font"
)

// windowTestTable builds a long table exercising the windowed planner:
// header and footer rows, rows of varying height, periodic rowspans and,
// optionally, border-collapse with alternating border widths so border
// resolution depends on each row's neighbors.
func windowTestTable(n int, collapse, footer bool) *Table {
	tbl := NewTable()
	tbl.SetAutoColumnWidths()
	tbl.SetBorderCollapse(collapse)
	h := tbl.AddHeaderRow()
	h.AddCell("Date", font.HelveticaBold, 10)
	h.AddCell("Description", font.HelveticaBold, 10)
	h.AddCell("Amount", font.HelveticaBold, 10)
	for i := range n {
		row := tbl.AddRow()
		if i%21 == 0 && i+3 < n {
			row.AddCell("span", font.Helvetica, 10).SetRowspan(3)
		} else if i%21 != 1 && i%21 != 2 {
			row.AddCell("2026-10-25", font.Helvetica, 10)
		}
		text := fmt.Sprintf("Line item %d", i)
		if i%5 == 0 {
			text += " with a much longer description that wraps onto several lines in its column"
		}
		row.AddCellElement(NewParagraph(text, font.Helvetica, 10))
		c := row.AddCell("$1.00", font.Helvetica, 10)
		if i%2 == 0 {
			b := DefaultBorder()
			b.Width = 2
			c.SetBorders(AllBorders(b))
		}
	}
	if footer {
		f := tbl.AddFooterRow()
		f.AddCell("", font.Helvetica, 10)
		f.AddCell("Total", font.HelveticaBold, 10)
		f.AddCell("$n", font.HelveticaBold, 10)
	}
	return tbl
}

// planChain plans tbl page by page, recording each page's consumed height
// and row-block geometry. full forces the unwindowed planner.
func planChain(t *testing.T, tbl *Table, area LayoutArea, full bool) []string {
	t.Helper()
	var pages []string
	for range 1000 {
		var plan LayoutPlan
		if full {
			cw := tbl.resolveColWidths(area.Width)
			plan, _ = tbl.planGrid(tbl, area, cw, -1)
		} else {
			plan = tbl.PlanLayout(area)
		}
		sig := fmt.Sprintf("status=%v consumed=%.3f", plan.Status, plan.Consumed)
		for _, b := range plan.Blocks {
			for _, c := range b.Children {
				sig += fmt.Sprintf(" [%.3f+%.3f]", c.Y, c.Height)
			}
		}
		pages = append(pages, sig)
		if plan.Status != LayoutPartial {
			return pages
		}
		tbl = plan.Overflow.(*Table)
	}
	t.Fatal("overflow chain did not terminate")
	return nil
}

// TestTableWindowedPlanMatchesFull asserts the windowed planner splits a
// long table into exactly the same pages, with the same row geometry, as
// planning the whole remaining table on every page.
func TestTableWindowedPlanMatchesFull(t *testing.T) {
	for _, collapse := range []bool{false, true} {
		for _, footer := range []bool{false, true} {
			// Sweep page heights so splits land at every offset relative
			// to the window boundaries (64, 128, ... body rows).
			for h := 90.0; h < 3200; h += 45 {
				name := fmt.Sprintf("collapse=%v/footer=%v/h=%v", collapse, footer, h)
				t.Run(name, func(t *testing.T) {
					area := LayoutArea{Width: 400, Height: h}
					got := planChain(t, windowTestTable(300, collapse, footer), area, false)
					want := planChain(t, windowTestTable(300, collapse, footer), area, true)
					if len(got) != len(want) {
						t.Fatalf("pages: windowed %d, full %d", len(got), len(want))
					}
					for i := range got {
						if got[i] != want[i] {
							t.Fatalf("page %d differs:\nwindowed: %s\nfull:     %s", i, got[i], want[i])
						}
					}
				})
			}
		}
	}
}

// TestTableCellHeightMemo asserts cell heights are memoised across the
// overflow chain and dropped by sizing mutators.
func TestTableCellHeightMemo(t *testing.T) {
	tbl := windowTestTable(200, false, false)
	plan := tbl.PlanLayout(LayoutArea{Width: 400, Height: 300})
	if plan.Status != LayoutPartial {
		t.Fatalf("status = %v, want partial", plan.Status)
	}
	if tbl.cellHeights == nil || len(tbl.cellHeights.m) == 0 {
		t.Fatal("expected PlanLayout to populate the cell height memo")
	}
	overflow := plan.Overflow.(*Table)
	if overflow.cellHeights != tbl.cellHeights {
		t.Error("continuation table should share the original's height memo")
	}
	tbl.AddRow().AddCell("x", font.Helvetica, 10)
	if tbl.cellHeights != nil {
		t.Error("AddRow should drop the height memo")
	}
}
