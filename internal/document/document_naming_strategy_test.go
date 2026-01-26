package document

import "testing"

func TestRegexStrategy_ExtractCode(t *testing.T) {
	tests := []struct {
		name      string
		strategy  *RegexStrategy
		dirName   string
		wantCode  string
		wantFound bool
	}{
		{
			name:      "DDDYY valid",
			strategy:  NewNamingStrategyDDDYY(2026),
			dirName:   "001-26",
			wantCode:  "001-26",
			wantFound: true,
		},
		{
			name:      "PrefixDDDYY valid",
			strategy:  NewNamingStrategyPrefixDDDYY("case", 2023),
			dirName:   "case-099-23",
			wantCode:  "099-23",
			wantFound: true,
		},
		{
			name:      "CaseDDDD valid",
			strategy:  NewNamingStrategyCaseDDDD("case"),
			dirName:   "case_0001",
			wantCode:  "0001",
			wantFound: true,
		},
		{
			name:      "CaseDDDD valid multiple digits",
			strategy:  NewNamingStrategyCaseDDDD("case"),
			dirName:   "case_9999",
			wantCode:  "9999",
			wantFound: true,
		},
		{
			name:      "CaseDDDD invalid - no underscore",
			strategy:  NewNamingStrategyCaseDDDD("case"),
			dirName:   "case0001",
			wantCode:  "",
			wantFound: false,
		},
		{
			name:      "CaseDDDD invalid - wrong prefix",
			strategy:  NewNamingStrategyCaseDDDD("case"),
			dirName:   "doc_0001",
			wantCode:  "",
			wantFound: false,
		},
		{
			name:      "CaseYYYYDDDD valid",
			strategy:  NewNamingStrategyCaseYYYYDDDD("case", 2026),
			dirName:   "case_2026_0001",
			wantCode:  "2026_0001",
			wantFound: true,
		},
		{
			name:      "CaseYYYYDDDD invalid - missing year",
			strategy:  NewNamingStrategyCaseYYYYDDDD("case", 2026),
			dirName:   "case_0001",
			wantCode:  "",
			wantFound: false,
		},
		{
			name:      "CaseHHHH valid hex",
			strategy:  NewNamingStrategyCaseHHHH("case"),
			dirName:   "case_00ff",
			wantCode:  "00ff",
			wantFound: true,
		},
		{
			name:      "CaseHHHH valid hex uppercase",
			strategy:  NewNamingStrategyCaseHHHH("case"),
			dirName:   "case_ABCD",
			wantCode:  "ABCD",
			wantFound: true,
		},
		{
			name:      "SimpleCaseD valid",
			strategy:  NewSimpleStrategyCaseD("case"),
			dirName:   "case_1",
			wantCode:  "1",
			wantFound: true,
		},
		{
			name:      "SimpleCaseD valid large number",
			strategy:  NewSimpleStrategyCaseD("case"),
			dirName:   "case_99999",
			wantCode:  "99999",
			wantFound: true,
		},
		{
			name:      "SimpleCaseD invalid - hex digits",
			strategy:  NewSimpleStrategyCaseD("case"),
			dirName:   "case_ABCD",
			wantCode:  "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, found := tt.strategy.ExtractCode(tt.dirName)
			if found != tt.wantFound {
				t.Errorf("ExtractCode() found = %v, want %v", found, tt.wantFound)
			}
			if code != tt.wantCode {
				t.Errorf("ExtractCode() code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestRegexStrategy_GenerateDirName(t *testing.T) {
	tests := []struct {
		name     string
		strategy *RegexStrategy
		code     string
		title    string
		want     string
	}{
		{
			name:     "DDDYY generates correct format",
			strategy: NewNamingStrategyDDDYY(2026),
			code:     "001-26",
			title:    "My Document",
			want:     "001-26",
		},
		{
			name:     "PrefixDDDYY generates correct format",
			strategy: NewNamingStrategyPrefixDDDYY("case", 2023),
			code:     "099-23",
			title:    "My Document",
			want:     "case-099-23",
		},
		{
			name:     "CaseDDDD generates correct format",
			strategy: NewNamingStrategyCaseDDDD("case"),
			code:     "0042",
			title:    "My Document",
			want:     "case_0042",
		},
		{
			name:     "CaseYYYYDDDD generates correct format",
			strategy: NewNamingStrategyCaseYYYYDDDD("case", 2026),
			code:     "2026_0042",
			title:    "My Document",
			want:     "case_2026_0042",
		},
		{
			name:     "CaseHHHH generates correct format",
			strategy: NewNamingStrategyCaseHHHH("case"),
			code:     "00ff",
			title:    "My Document",
			want:     "case_00ff",
		},
		{
			name:     "SimpleCaseD generates correct format",
			strategy: NewSimpleStrategyCaseD("case"),
			code:     "42",
			title:    "My Document",
			want:     "case_42",
		},
		{
			name:     "Different prefix",
			strategy: NewNamingStrategyCaseDDDD("doc"),
			code:     "0100",
			title:    "Some Title",
			want:     "doc_0100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.strategy.GenerateDirName(tt.code, tt.title)
			if got != tt.want {
				t.Errorf("GenerateDirName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRegexStrategy_CalculateNextCode(t *testing.T) {
	tests := []struct {
		name         string
		strategy     *RegexStrategy
		existingCode []string
		want         string
	}{
		{
			name:         "DDDYY empty list",
			strategy:     NewNamingStrategyDDDYY(2026),
			existingCode: []string{},
			want:         "001-26",
		},
		{
			name:         "DDDYY with codes",
			strategy:     NewNamingStrategyDDDYY(2026),
			existingCode: []string{"001-26", "010-26", "009-26"},
			want:         "011-26",
		},
		{
			name:         "PrefixDDDYY empty list",
			strategy:     NewNamingStrategyPrefixDDDYY("case", 2023),
			existingCode: []string{},
			want:         "001-23",
		},
		{
			name:         "PrefixDDDYY with codes",
			strategy:     NewNamingStrategyPrefixDDDYY("case", 2023),
			existingCode: []string{"099-23", "100-23", "010-23"},
			want:         "101-23",
		},
		{
			name:         "CaseDDDD empty list",
			strategy:     NewNamingStrategyCaseDDDD("case"),
			existingCode: []string{},
			want:         "0001",
		},
		{
			name:         "CaseDDDD single code",
			strategy:     NewNamingStrategyCaseDDDD("case"),
			existingCode: []string{"0001"},
			want:         "0002",
		},
		{
			name:         "CaseDDDD multiple codes",
			strategy:     NewNamingStrategyCaseDDDD("case"),
			existingCode: []string{"0001", "0005", "0003"},
			want:         "0006",
		},
		{
			name:         "CaseDDDD finds maximum",
			strategy:     NewNamingStrategyCaseDDDD("case"),
			existingCode: []string{"0010", "0002", "0099"},
			want:         "0100",
		},
		{
			name:         "CaseYYYYDDDD empty",
			strategy:     NewNamingStrategyCaseYYYYDDDD("case", 2026),
			existingCode: []string{},
			want:         "2026_0001",
		},
		{
			name:         "CaseYYYYDDDD with codes",
			strategy:     NewNamingStrategyCaseYYYYDDDD("case", 2026),
			existingCode: []string{"2026_0001", "2026_0005", "2026_0003"},
			want:         "2026_0006",
		},
		{
			name:         "CaseHHHH empty",
			strategy:     NewNamingStrategyCaseHHHH("case"),
			existingCode: []string{},
			want:         "0001",
		},
		{
			name:         "CaseHHHH with codes",
			strategy:     NewNamingStrategyCaseHHHH("case"),
			existingCode: []string{"0001", "000a", "000f"},
			want:         "0010",
		},
		{
			name:         "SimpleCaseD empty",
			strategy:     NewSimpleStrategyCaseD("case"),
			existingCode: []string{},
			want:         "1",
		},
		{
			name:         "SimpleCaseD with codes",
			strategy:     NewSimpleStrategyCaseD("case"),
			existingCode: []string{"1", "5", "3"},
			want:         "6",
		},
		{
			name:         "SimpleCaseD large numbers",
			strategy:     NewSimpleStrategyCaseD("case"),
			existingCode: []string{"99", "100", "50"},
			want:         "101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.strategy.CalculateNextCode(tt.existingCode)
			if got != tt.want {
				t.Errorf("CalculateNextCode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNamingStrategyConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() *RegexStrategy
		testDirName  string
		expectedCode string
	}{
		{
			name:         "NewNamingStrategyDDDYY",
			constructor:  func() *RegexStrategy { return NewNamingStrategyDDDYY(2026) },
			testDirName:  "001-26",
			expectedCode: "001-26",
		},
		{
			name:         "NewNamingStrategyPrefixDDDYY",
			constructor:  func() *RegexStrategy { return NewNamingStrategyPrefixDDDYY("case", 2023) },
			testDirName:  "case-099-23",
			expectedCode: "099-23",
		},
		{
			name:         "NewNamingStrategyCaseDDDD",
			constructor:  func() *RegexStrategy { return NewNamingStrategyCaseDDDD("case") },
			testDirName:  "case_0050",
			expectedCode: "0050",
		},
		{
			name:         "NewNamingStrategyCaseYYYYDDDD",
			constructor:  func() *RegexStrategy { return NewNamingStrategyCaseYYYYDDDD("case", 2026) },
			testDirName:  "case_2026_0050",
			expectedCode: "2026_0050",
		},
		{
			name:         "NewNamingStrategyCaseHHHH",
			constructor:  func() *RegexStrategy { return NewNamingStrategyCaseHHHH("case") },
			testDirName:  "case_00a5",
			expectedCode: "00a5",
		},
		{
			name:         "NewSimpleStrategyCaseD",
			constructor:  func() *RegexStrategy { return NewSimpleStrategyCaseD("case") },
			testDirName:  "case_50",
			expectedCode: "50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := tt.constructor()
			if strategy == nil {
				t.Fatal("constructor returned nil")
			}

			code, found := strategy.ExtractCode(tt.testDirName)
			if !found {
				t.Errorf("ExtractCode() found = false, want true")
			}
			if code != tt.expectedCode {
				t.Errorf("ExtractCode() code = %q, want %q", code, tt.expectedCode)
			}
		})
	}
}

func TestRegexStrategy_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		strategy *RegexStrategy
		code     string
	}{
		{
			name:     "DDDYY round trip",
			strategy: NewNamingStrategyDDDYY(2026),
			code:     "001-26",
		},
		{
			name:     "PrefixDDDYY round trip",
			strategy: NewNamingStrategyPrefixDDDYY("case", 2023),
			code:     "099-23",
		},
		{
			name:     "CaseDDDD round trip",
			strategy: NewNamingStrategyCaseDDDD("case"),
			code:     "0042",
		},
		{
			name:     "CaseYYYYDDDD round trip",
			strategy: NewNamingStrategyCaseYYYYDDDD("case", 2026),
			code:     "2026_0042",
		},
		{
			name:     "CaseHHHH round trip",
			strategy: NewNamingStrategyCaseHHHH("case"),
			code:     "00ff",
		},
		{
			name:     "SimpleCaseD round trip",
			strategy: NewSimpleStrategyCaseD("case"),
			code:     "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirName := tt.strategy.GenerateDirName(tt.code, "")
			extracted, found := tt.strategy.ExtractCode(dirName)

			if !found {
				t.Errorf("ExtractCode() found = false, want true")
			}
			if extracted != tt.code {
				t.Errorf("Round trip failed: code = %q, extracted = %q", tt.code, extracted)
			}
		})
	}
}
