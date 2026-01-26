package document

import (
	"fmt"
	"strconv"
	"strings"
)

type CodeFormat interface {
	Parse(s string) (int64, error)
	Format(val int64) string
}

type DecimalFormat struct {
	Padding int
}

func (d *DecimalFormat) Parse(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (d *DecimalFormat) Format(val int64) string {
	return fmt.Sprintf("%0*d", d.Padding, val)
}

type HexFormat struct {
	Padding int
}

func (h *HexFormat) Parse(s string) (int64, error) {
	return strconv.ParseInt(s, 16, 64)
}

func (h *HexFormat) Format(val int64) string {
	return fmt.Sprintf("%0*x", h.Padding, val)
}

type YearlyDecimalFormat struct {
	Year    int
	Padding int
}

func (y *YearlyDecimalFormat) Parse(s string) (int64, error) {
	// Expects format "2026_0001"
	// Split by underscore and take the second part
	parts := strings.Split(s, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format")
	}
	return strconv.ParseInt(parts[1], 10, 64)
}

func (y *YearlyDecimalFormat) Format(val int64) string {
	// Returns "2026_0001"
	return fmt.Sprintf("%d_%0*d", y.Year, y.Padding, val)
}

// YearSuffixDecimalFormat formats codes like "001-26" where the suffix is a two-digit year.
type YearSuffixDecimalFormat struct {
	Year    int
	Padding int
}

func (y *YearSuffixDecimalFormat) Parse(s string) (int64, error) {
	// Expect format "001-26"; extract the numeric portion before the dash.
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format")
	}
	// Validate the year suffix matches the configured year to avoid mixing years.
	suffixYear, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}
	if int(suffixYear) != y.Year%100 {
		return 0, fmt.Errorf("year suffix mismatch")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func (y *YearSuffixDecimalFormat) Format(val int64) string {
	return fmt.Sprintf("%0*d-%02d", y.Padding, val, y.Year%100)
}
