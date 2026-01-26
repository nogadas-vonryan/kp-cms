package document

import (
	"fmt"
	"regexp"
)

type NamingStrategy interface {
	ExtractCode(dirName string) (string, bool)
	GenerateDirName(code string, title string) string
	CalculateNextCode(existingCodes []string) string
}

type RegexStrategy struct {
	Pattern *regexp.Regexp
	Format  string
	CodeFmt CodeFormat
}

func (s *RegexStrategy) ExtractCode(dirName string) (string, bool) {
	matches := s.Pattern.FindStringSubmatch(dirName)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

func (s *RegexStrategy) GenerateDirName(code string, title string) string {
	return fmt.Sprintf(s.Format, code)
}

func (s *RegexStrategy) CalculateNextCode(existingCodes []string) string {
	var maxID int64 = 0
	for _, code := range existingCodes {
		val, err := s.CodeFmt.Parse(code)
		if err == nil && val > maxID {
			maxID = val
		}
	}
	return s.CodeFmt.Format(maxID + 1)
}

// NewNamingStrategyCaseDDDD: "case_0001", "case_0002" , ... , "case_9999"
func NewNamingStrategyCaseDDDD(prefix string) *RegexStrategy {
	return &RegexStrategy{
		// Regex: start of string + prefix + underscore + (capture code) + end of string
		Pattern: regexp.MustCompile(fmt.Sprintf(`^%s_([a-zA-Z0-9]+)$`, prefix)),
		Format:  prefix + "_%s",
		CodeFmt: &DecimalFormat{Padding: 4},
	}
}

// NewNamingStrategyCaseYYYYDDDD: "case_2026_0001"
func NewNamingStrategyCaseYYYYDDDD(prefix string, year int) *RegexStrategy {
	return &RegexStrategy{
		// Regex: prefix + underscore + (Year_Digits)
		// Example: ^case_([0-9]{4}_[0-9]+)$
		Pattern: regexp.MustCompile(fmt.Sprintf(`^%s_([0-9]{4}_[0-9]+)$`, prefix)),
		Format:  prefix + "_%s",
		CodeFmt: &YearlyDecimalFormat{
			Year:    year,
			Padding: 4,
		},
	}
}

// NewNamingStrategyDDDYY: "001-26", "099-23"
func NewNamingStrategyDDDYY(year int) *RegexStrategy {
	return &RegexStrategy{
		Pattern: regexp.MustCompile(`^([0-9]{3}-[0-9]{2})$`),
		Format:  "%s",
		CodeFmt: &YearSuffixDecimalFormat{
			Year:    year,
			Padding: 3,
		},
	}
}

// NewNamingStrategyPrefixDDDYY: "case-001-26", "case-099-23"
func NewNamingStrategyPrefixDDDYY(prefix string, year int) *RegexStrategy {
	return &RegexStrategy{
		Pattern: regexp.MustCompile(fmt.Sprintf(`^%s-([0-9]{3}-[0-9]{2})$`, prefix)),
		Format:  prefix + "-%s",
		CodeFmt: &YearSuffixDecimalFormat{
			Year:    year,
			Padding: 3,
		},
	}
}

// NewNamingStrategyCaseDDDD: "case_0001", "case_0002", ... , "case_FFFF"
func NewNamingStrategyCaseHHHH(prefix string) *RegexStrategy {
	return &RegexStrategy{
		Pattern: regexp.MustCompile(fmt.Sprintf(`^%s_([a-fA-F0-9]+)$`, prefix)),
		Format:  prefix + "_%s",
		CodeFmt: &HexFormat{Padding: 4},
	}
}

// NewSimpleStrategyCaseD: "case_1", "case_2" ... "case_10" ... "case_99999"
func NewSimpleStrategyCaseD(prefix string) *RegexStrategy {
	return &RegexStrategy{
		// Regex: prefix + underscore + (one or more digits)
		Pattern: regexp.MustCompile(fmt.Sprintf(`^%s_([0-9]+)$`, prefix)),
		Format:  prefix + "_%s",
		CodeFmt: &DecimalFormat{Padding: 1}, // Minimum 1 digit, no leading zeros
	}
}
