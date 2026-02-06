package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"kpcms/server/core/document"
	"kpcms/server/core/inhabitant"
)

// InhabitantFromDB represents an inhabitant from SQLite
type InhabitantFromDB struct {
	ID                           int64
	FirstName                    string
	LastName                     string
	MiddleName                   string
	Suffix                       string
	Birthdate                    string
	BirthPlace                   string
	InhabitantType               string
	Sex                          string
	CivilStatus                  string
	Citizenship                  string
	Occupation                   string
	EmailAddress                 string
	HighestEducationalAttainment string
	MotherFirstName              string
	MotherMiddleName             string
	MotherLastName               string
	ContactNo                    string
	Address                      string
}

func main() {
	dataPath := flag.String("data", "./data", "Path to the data directory")
	dbPath := flag.String("db", "", "Path to SQLite database (defaults to dataPath/app.db)")
	flag.Parse()

	if *dataPath == "" {
		log.Fatal("Data path is required")
	}

	dbFile := *dbPath
	if dbFile == "" {
		dbFile = filepath.Join(*dataPath, "app.db")
	}

	// Check if database exists
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbFile)
	}

	inhabitantsPath := filepath.Join(*dataPath, "inhabitants")
	if err := os.MkdirAll(inhabitantsPath, 0755); err != nil {
		log.Fatalf("Failed to create inhabitants directory: %v", err)
	}

	log.Printf("Data path: %s", *dataPath)
	log.Printf("Database: %s", dbFile)
	log.Printf("Inhabitants path: %s", inhabitantsPath)

	// Open database
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Load all inhabitants from database
	inhabitants, err := loadInhabitants(db)
	if err != nil {
		log.Fatalf("Failed to load inhabitants: %v", err)
	}
	log.Printf("Loaded %d inhabitants from database", len(inhabitants))

	// Create naming strategy
	namingStrategy := document.NewNamingStrategyPrefixDDDYY("inhabitant")

	// Get existing codes to avoid conflicts
	var existingCodes []string
	entries, err := os.ReadDir(inhabitantsPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				if code, _ := namingStrategy.ExtractCode(entry.Name()); code != "" {
					existingCodes = append(existingCodes, code)
				}
			}
		}
	}

	// Calculate next codes for database inhabitants
	nextCode := namingStrategy.CalculateNextCode(existingCodes)

	// Migrate each inhabitant
	var migratedCount int
	idToCodeMapping := make(map[int64]string)

	for _, inh := range inhabitants {
		// Generate code for this inhabitant
		code := nextCode
		nextCode = incrementCode(nextCode)

		folderName := namingStrategy.GenerateDirName(code, "")

		// Parse birthdate string to time.Time
		var birthdate time.Time
		if inh.Birthdate != "" {
			birthdate, _ = time.Parse(time.RFC3339, inh.Birthdate)
		}

		// Create file-based inhabitant
		fileInhabitant := &inhabitant.Inhabitant{
			UUID:       generateUUID(),
			Code:       code,
			FolderName: folderName,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),

			// Copy existing fields
			ID:                           inh.ID,
			FirstName:                    inh.FirstName,
			LastName:                     inh.LastName,
			MiddleName:                   inh.MiddleName,
			Suffix:                       inh.Suffix,
			Birthdate:                    birthdate,
			BirthPlace:                   inh.BirthPlace,
			InhabitantType:               inh.InhabitantType,
			Sex:                          inh.Sex,
			CivilStatus:                  inh.CivilStatus,
			Citizenship:                  inh.Citizenship,
			Occupation:                   inh.Occupation,
			EmailAddress:                 inh.EmailAddress,
			HighestEducationalAttainment: inh.HighestEducationalAttainment,
			MotherFirstName:              inh.MotherFirstName,
			MotherMiddleName:             inh.MotherMiddleName,
			MotherLastName:               inh.MotherLastName,
			ContactNo:                    inh.ContactNo,
			Address:                      inh.Address,
		}

		// Write to disk
		folderPath := filepath.Join(inhabitantsPath, folderName)
		if err := os.MkdirAll(folderPath, 0755); err != nil {
			log.Printf("WARNING: Failed to create folder %s: %v", folderPath, err)
			continue
		}

		data, err := json.MarshalIndent(fileInhabitant, "", "  ")
		if err != nil {
			log.Printf("WARNING: Failed to marshal inhabitant %d: %v", inh.ID, err)
			continue
		}

		filePath := filepath.Join(folderPath, "inhabitant.json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			log.Printf("WARNING: Failed to write inhabitant %d: %v", inh.ID, err)
			continue
		}

		idToCodeMapping[inh.ID] = code
		migratedCount++

		log.Printf("Migrated inhabitant %d -> %s", inh.ID, folderName)
	}

	log.Printf("Successfully migrated %d inhabitants", migratedCount)

	// Write mapping file for reference
	mappingFile := filepath.Join(*dataPath, "inhabitant_id_mapping.json")
	mappingData, _ := json.MarshalIndent(idToCodeMapping, "", "  ")
	if err := os.WriteFile(mappingFile, mappingData, 0644); err != nil {
		log.Printf("WARNING: Failed to write mapping file: %v", err)
	}
	log.Printf("Wrote ID mapping to %s", mappingFile)
}

func loadInhabitants(db *sql.DB) ([]InhabitantFromDB, error) {
	rows, err := db.Query(`
		SELECT id, first_name, last_name, middle_name, suffix,
			   birthdate, birth_place, inhabitant_type, sex,
			   civil_status, citizenship, occupation, email_address,
			   highest_educational_attainment, mother_first_name,
			   mother_middle_name, mother_last_name, contact_no, address
		FROM inhabitants
	`)
	if err != nil {
		return nil, fmt.Errorf("query inhabitants: %w", err)
	}
	defer rows.Close()

	var inhabitants []InhabitantFromDB
	for rows.Next() {
		var inh InhabitantFromDB
		err := rows.Scan(
			&inh.ID, &inh.FirstName, &inh.LastName, &inh.MiddleName, &inh.Suffix,
			&inh.Birthdate, &inh.BirthPlace, &inh.InhabitantType, &inh.Sex,
			&inh.CivilStatus, &inh.Citizenship, &inh.Occupation, &inh.EmailAddress,
			&inh.HighestEducationalAttainment, &inh.MotherFirstName,
			&inh.MotherMiddleName, &inh.MotherLastName, &inh.ContactNo, &inh.Address,
		)
		if err != nil {
			log.Printf("WARNING: Failed to scan inhabitant row: %v", err)
			continue
		}
		inhabitants = append(inhabitants, inh)
	}

	return inhabitants, nil
}

// incrementCode increments a DDD-YY code (e.g., "001-26" -> "002-26")
func incrementCode(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) != 2 {
		return code
	}
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return code
	}
	num++
	return fmt.Sprintf("%03d-%s", num, parts[1])
}

// generateUUID generates a proper UUID using google/uuid package.
func generateUUID() string {
	return uuid.New().String()
}
