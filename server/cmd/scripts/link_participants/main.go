package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"kpcms/server/core/document"
)

// Inhabitant represents an inhabitant from SQLite
type Inhabitant struct {
	ID         int64
	FirstName  string
	LastName   string
	MiddleName string
	Suffix     string
}

// ParticipantMapping represents a mapping between document participant names and inhabitant codes
type ParticipantMapping struct {
	DocumentCode    string  `json:"document_code"`
	DocumentUUID    string  `json:"document_uuid"`
	ParticipantType string  `json:"participant_type"` // "complainant" or "respondent"
	ParticipantName string  `json:"participant_name"`
	InhabitantCode  string  `json:"inhabitant_code,omitempty"`
	Confidence      float64 `json:"confidence"`
	MatchedBy       string  `json:"matched_by"` // "full_name", "short_name", "formal_name", "none"
}

type MappingReport struct {
	GeneratedAt       time.Time            `json:"generated_at"`
	TotalDocuments    int                  `json:"total_documents"`
	TotalParticipants int                  `json:"total_participants"`
	Mappings          []ParticipantMapping `json:"mappings"`
	UnmatchedNames    []string             `json:"unmatched_names"`
}

func main() {
	dataPath := flag.String("data", "./data", "Path to the data directory")
	dbPath := flag.String("db", "", "Path to SQLite database (defaults to dataPath/app.db)")
	report := flag.Bool("report", false, "Generate mapping report (name → inhabitant_code with confidence scores)")
	apply := flag.Bool("apply", false, "Apply mappings to documents")
	mappingFile := flag.String("mapping", "participant_mappings.json", "Path to mapping file for --apply mode")
	dryRun := flag.Bool("dry-run", false, "Preview changes without modifying files (use with --apply)")
	flag.Parse()

	if *dataPath == "" {
		log.Fatal("Data path is required")
	}

	casesPath := filepath.Join(*dataPath, "cases")
	if _, err := os.Stat(casesPath); os.IsNotExist(err) {
		log.Fatalf("Cases directory does not exist: %s", casesPath)
	}

	dbFile := *dbPath
	if dbFile == "" {
		dbFile = filepath.Join(*dataPath, "app.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Printf("WARNING: Database file does not exist: %s", dbFile)
		log.Println("Running in database-less mode (name matching will be skipped)")
		dbFile = ""
	}

	log.Printf("Data path: %s", *dataPath)
	log.Printf("Cases path: %s", casesPath)
	if dbFile != "" {
		log.Printf("Database: %s", dbFile)
	}

	// Load all documents
	documents, err := loadDocuments(casesPath)
	if err != nil {
		log.Fatalf("Failed to load documents: %v", err)
	}
	log.Printf("Loaded %d documents", len(documents))

	// Load all inhabitants from database
	var inhabitants []Inhabitant
	if dbFile != "" {
		inhabitants, err = loadInhabitants(dbFile)
		if err != nil {
			log.Printf("WARNING: Failed to load inhabitants from database: %v", err)
		} else {
			log.Printf("Loaded %d inhabitants from database", len(inhabitants))
		}
	}

	// Build name index for matching
	nameIndex := buildNameIndex(inhabitants)

	// Generate mappings
	mappings, unmatchedNames := generateMappings(documents, nameIndex)

	if *report {
		report := MappingReport{
			GeneratedAt:       time.Now(),
			TotalDocuments:    len(documents),
			TotalParticipants: len(mappings),
			Mappings:          mappings,
			UnmatchedNames:    unmatchedNames,
		}
		reportJSON, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(reportJSON))
	}

	if *apply {
		if err := applyMappings(casesPath, mappings, *mappingFile, *dryRun); err != nil {
			log.Fatalf("Failed to apply mappings: %v", err)
		}
	}
}

func loadDocuments(casesPath string) ([]document.Document, error) {
	var documents []document.Document

	entries, err := os.ReadDir(casesPath)
	if err != nil {
		return nil, fmt.Errorf("read cases directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(casesPath, entry.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			log.Printf("WARNING: Failed to read %s: %v", metaPath, err)
			continue
		}

		var doc document.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			log.Printf("WARNING: Failed to parse %s: %v", metaPath, err)
			continue
		}

		documents = append(documents, doc)
	}

	return documents, nil
}

func loadInhabitants(dbPath string) ([]Inhabitant, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, first_name, last_name, middle_name, suffix FROM inhabitants")
	if err != nil {
		return nil, fmt.Errorf("query inhabitants: %w", err)
	}
	defer rows.Close()

	var inhabitants []Inhabitant
	for rows.Next() {
		var inh Inhabitant
		if err := rows.Scan(&inh.ID, &inh.FirstName, &inh.LastName, &inh.MiddleName, &inh.Suffix); err != nil {
			log.Printf("WARNING: Failed to scan inhabitant row: %v", err)
			continue
		}
		inhabitants = append(inhabitants, inh)
	}

	return inhabitants, nil
}

func buildNameIndex(inhabitants []Inhabitant) map[string][]Inhabitant {
	index := make(map[string][]Inhabitant)

	for _, inh := range inhabitants {
		// Full name: "FirstName MiddleName LastName Suffix"
		fullName := buildFullName(inh)
		if fullName != "" {
			index[strings.ToLower(fullName)] = append(index[strings.ToLower(fullName)], inh)
		}

		// Short name: "FirstName LastName"
		if inh.FirstName != "" && inh.LastName != "" {
			shortName := strings.ToLower(inh.FirstName + " " + inh.LastName)
			index[shortName] = append(index[shortName], inh)
		}

		// Formal name: "LastName, FirstName"
		if inh.LastName != "" && inh.FirstName != "" {
			formalName := strings.ToLower(inh.LastName + ", " + inh.FirstName)
			index[formalName] = append(index[formalName], inh)
		}

		// Formal with middle initial: "LastName, FirstName M."
		if inh.LastName != "" && inh.FirstName != "" && inh.MiddleName != "" {
			middleInitial := string([]rune(inh.MiddleName)[0])
			formalWithMiddle := strings.ToLower(inh.LastName + ", " + inh.FirstName + " " + middleInitial)
			index[formalWithMiddle] = append(index[formalWithMiddle], inh)
		}
	}

	return index
}

func buildFullName(inh Inhabitant) string {
	parts := []string{}
	if inh.FirstName != "" {
		parts = append(parts, inh.FirstName)
	}
	if inh.MiddleName != "" {
		parts = append(parts, inh.MiddleName)
	}
	if inh.LastName != "" {
		parts = append(parts, inh.LastName)
	}
	if inh.Suffix != "" {
		parts = append(parts, inh.Suffix)
	}
	return strings.Join(parts, " ")
}

func generateMappings(documents []document.Document, nameIndex map[string][]Inhabitant) ([]ParticipantMapping, []string) {
	var mappings []ParticipantMapping
	var unmatchedNames []string
	seenNames := make(map[string]bool)

	for _, doc := range documents {
		// Process complainants
		if complainants, ok := doc.Fields["complainants"]; ok {
			if complainantList, ok := complainantListFromField(complainants); ok {
				for _, name := range complainantList {
					if name == "" {
						continue
					}
					mapping := matchName(name, "complainant", doc, nameIndex)
					mappings = append(mappings, mapping)
					if mapping.InhabitantCode == "" && !seenNames[name] {
						unmatchedNames = append(unmatchedNames, name)
						seenNames[name] = true
					}
				}
			}
		}

		// Process respondents
		if respondents, ok := doc.Fields["respondents"]; ok {
			if respondentList, ok := respondentListFromField(respondents); ok {
				for _, name := range respondentList {
					if name == "" {
						continue
					}
					mapping := matchName(name, "respondent", doc, nameIndex)
					mappings = append(mappings, mapping)
					if mapping.InhabitantCode == "" && !seenNames[name] {
						unmatchedNames = append(unmatchedNames, name)
						seenNames[name] = true
					}
				}
			}
		}
	}

	return mappings, unmatchedNames
}

func complainantListFromField(field any) ([]string, bool) {
	if list, ok := field.([]any); ok {
		result := make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result, true
	}
	if s, ok := field.(string); ok && s != "" {
		return []string{s}, true
	}
	return nil, false
}

func respondentListFromField(field any) ([]string, bool) {
	return complainantListFromField(field)
}

func matchName(name string, participantType string, doc document.Document, nameIndex map[string][]Inhabitant) ParticipantMapping {
	mapping := ParticipantMapping{
		DocumentCode:    doc.Code,
		DocumentUUID:    doc.UUID,
		ParticipantType: participantType,
		ParticipantName: name,
		Confidence:      0,
	}

	normalizedName := normalizeName(name)

	// Try to find matches in the name index
	if candidates, exists := nameIndex[normalizedName]; exists && len(candidates) > 0 {
		// Found exact match
		inh := candidates[0]
		// Extract year from document code (e.g., "001-26" -> "26")
		year := extractYear(doc.Code)
		mapping.InhabitantCode = fmt.Sprintf("inhabitant-%03d-%s", inh.ID, year)
		mapping.Confidence = 1.0
		mapping.MatchedBy = detectMatchType(inh, name)
		return mapping
	}

	// Try tokenized matching
	bestMatch := findBestTokenizedMatch(name, nameIndex, doc.Code)
	if bestMatch.inhabitantCode != "" {
		mapping.InhabitantCode = bestMatch.inhabitantCode
		mapping.Confidence = bestMatch.confidence
		mapping.MatchedBy = "tokenized"
		return mapping
	}

	return mapping
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func detectMatchType(inh Inhabitant, originalName string) string {
	normalizedOriginal := normalizeName(originalName)

	// Full name match
	fullName := strings.ToLower(buildFullName(inh))
	if normalizedOriginal == fullName {
		return "full_name"
	}

	// Short name match
	if inh.FirstName != "" && inh.LastName != "" {
		shortName := strings.ToLower(inh.FirstName + " " + inh.LastName)
		if normalizedOriginal == shortName {
			return "short_name"
		}
	}

	// Formal name match
	if inh.LastName != "" && inh.FirstName != "" {
		formalName := strings.ToLower(inh.LastName + ", " + inh.FirstName)
		if normalizedOriginal == formalName {
			return "formal_name"
		}
	}

	return "partial"
}

// extractYear extracts the year suffix from a document code (e.g., "001-26" -> "26")
func extractYear(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return "00"
}

type tokenizedMatch struct {
	inhabitantCode string
	confidence     float64
}

func findBestTokenizedMatch(name string, nameIndex map[string][]Inhabitant, docCode string) tokenizedMatch {
	normalizedName := normalizeName(name)
	tokens := strings.FieldsFunc(normalizedName, func(r rune) bool {
		return r == ' ' || r == ','
	})

	if len(tokens) == 0 {
		return tokenizedMatch{}
	}

	var bestMatch tokenizedMatch
	bestMatchConfidence := 0.0

	for indexedName, candidates := range nameIndex {
		indexedTokens := strings.FieldsFunc(indexedName, func(r rune) bool {
			return r == ' ' || r == ','
		})

		// Check if all tokens in the search name exist in the indexed name
		matchedTokens := 0
		for _, token := range tokens {
			for _, indexedToken := range indexedTokens {
				if strings.Contains(indexedToken, token) {
					matchedTokens++
					break
				}
			}
		}

		confidence := float64(matchedTokens) / float64(len(tokens))
		if confidence > bestMatchConfidence {
			bestMatchConfidence = confidence
			if len(candidates) > 0 {
				// Extract year from document code
				year := extractYear(docCode)
				bestMatch = tokenizedMatch{
					inhabitantCode: fmt.Sprintf("inhabitant-%03d-%s", candidates[0].ID, year),
					confidence:     confidence,
				}
			}
		}
	}

	if bestMatchConfidence > 0.5 {
		return bestMatch
	}

	return tokenizedMatch{}
}

func applyMappings(casesPath string, mappings []ParticipantMapping, mappingFile string, dryRun bool) error {
	log.Printf("Applying mappings from %s (dry-run: %v)", mappingFile, dryRun)

	// Group mappings by document
	docMappings := make(map[string][]ParticipantMapping)
	for _, m := range mappings {
		docMappings[m.DocumentUUID] = append(docMappings[m.DocumentUUID], m)
	}

	var updatedCount int
	for uuid, docMappings := range docMappings {
		metaPath := filepath.Join(casesPath, getFolderNameFromUUID(casesPath, uuid), "meta.json")
		if _, err := os.Stat(metaPath); os.IsNotExist(err) {
			log.Printf("WARNING: Document %s not found", uuid)
			continue
		}

		data, err := os.ReadFile(metaPath)
		if err != nil {
			log.Printf("WARNING: Failed to read %s: %v", metaPath, err)
			continue
		}

		var doc document.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			log.Printf("WARNING: Failed to parse %s: %v", metaPath, err)
			continue
		}

		// Build participant IDs
		if doc.ParticipantIDs == nil {
			doc.ParticipantIDs = &document.ParticipantLinks{}
		}

		updated := false
		for _, m := range docMappings {
			if m.InhabitantCode == "" {
				continue
			}

			if m.ParticipantType == "complainant" {
				if !contains(doc.ParticipantIDs.Complainants, m.InhabitantCode) {
					doc.ParticipantIDs.Complainants = append(doc.ParticipantIDs.Complainants, m.InhabitantCode)
					updated = true
				}
			} else if m.ParticipantType == "respondent" {
				if !contains(doc.ParticipantIDs.Respondents, m.InhabitantCode) {
					doc.ParticipantIDs.Respondents = append(doc.ParticipantIDs.Respondents, m.InhabitantCode)
					updated = true
				}
			}
		}

		if updated {
			updatedCount++
			if dryRun {
				log.Printf("[DRY-RUN] Would update %s: %s", doc.Code, metaPath)
			} else {
				newData, err := json.MarshalIndent(doc, "", "  ")
				if err != nil {
					log.Printf("WARNING: Failed to marshal %s: %v", metaPath, err)
					continue
				}

				if err := os.WriteFile(metaPath, newData, 0644); err != nil {
					log.Printf("WARNING: Failed to write %s: %v", metaPath, err)
					continue
				}

				log.Printf("Updated %s: %s", doc.Code, metaPath)
			}
		}
	}

	log.Printf("%s: %d documents updated", map[bool]string{true: "[DRY-RUN]", false: ""}[dryRun], updatedCount)

	// Save mapping file if it doesn't exist
	if _, err := os.Stat(mappingFile); os.IsNotExist(err) {
		mappingJSON, _ := json.MarshalIndent(mappings, "", "  ")
		if err := os.WriteFile(mappingFile, mappingJSON, 0644); err != nil {
			log.Printf("WARNING: Failed to write mapping file: %v", err)
		} else {
			log.Printf("Saved mapping file: %s", mappingFile)
		}
	}

	return nil
}

func getFolderNameFromUUID(casesPath string, uuid string) string {
	entries, _ := os.ReadDir(casesPath)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(casesPath, entry.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var doc document.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			continue
		}
		if doc.UUID == uuid {
			return entry.Name()
		}
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var _ = regexp.MustCompile // Import for compilation check
var _ = sort.Strings       // Import for compilation check
