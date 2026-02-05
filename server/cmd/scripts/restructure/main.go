package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const caseFolderPattern = `^case-\d{3}-\d{2}$`

func main() {
	dataPath := flag.String("data", "./data", "Path to the data directory")
	dryRun := flag.Bool("dry-run", false, "Preview changes without modifying files")
	backup := flag.Bool("backup", false, "Create backup of original data folder before migration")
	flag.Parse()

	if *dataPath == "" {
		log.Fatal("Data path is required")
	}

	// Validate and resolve the data path
	absDataPath, err := filepath.Abs(*dataPath)
	if err != nil {
		log.Fatalf("Failed to resolve data path: %v", err)
	}

	// Check if data directory exists
	if _, err := os.Stat(absDataPath); os.IsNotExist(err) {
		log.Fatalf("Data directory does not exist: %s", absDataPath)
	}

	log.Printf("Data path: %s", absDataPath)

	// Create backup if requested
	if *backup {
		if err := createBackup(absDataPath); err != nil {
			log.Fatalf("Failed to create backup: %v", err)
		}
	}

	// Create the cases, inhabitants, and custom directories
	casesPath := filepath.Join(absDataPath, "cases")
	inhabitantsPath := filepath.Join(absDataPath, "inhabitants")
	customPath := filepath.Join(absDataPath, "custom")

	dirsToCreate := []string{casesPath, inhabitantsPath, customPath}
	for _, dir := range dirsToCreate {
		if !*dryRun {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
		log.Printf("Directory: %s (exists or created)", dir)
	}

	// Pattern to match case-XXX-YY folders
	pattern := regexp.MustCompile(caseFolderPattern)

	// Read all entries in the data directory
	entries, err := os.ReadDir(absDataPath)
	if err != nil {
		log.Fatalf("Failed to read data directory: %v", err)
	}

	var foldersToMigrate []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		// Skip the entity type folders we just created
		if folderName == "cases" || folderName == "inhabitants" || folderName == "custom" {
			continue
		}

		// Check if it matches the case-XXX-YY pattern
		if pattern.MatchString(folderName) {
			foldersToMigrate = append(foldersToMigrate, folderName)
		}
	}

	if len(foldersToMigrate) == 0 {
		log.Println("No case folders found to migrate")
		return
	}

	log.Printf("Found %d case folders to migrate: %v", len(foldersToMigrate), foldersToMigrate)

	// Migrate each folder
	for _, folderName := range foldersToMigrate {
		srcPath := filepath.Join(absDataPath, folderName)
		dstPath := filepath.Join(casesPath, folderName)

		if *dryRun {
			log.Printf("[DRY-RUN] Would move: %s -> %s", srcPath, dstPath)
		} else {
			if err := os.Rename(srcPath, dstPath); err != nil {
				log.Printf("WARNING: Failed to move %s: %v", folderName, err)
				continue
			}
			log.Printf("Migrated: %s -> %s", folderName, dstPath)
		}
	}

	// Validate migration
	if !*dryRun {
		if err := validateMigration(absDataPath, casesPath); err != nil {
			log.Fatalf("Migration validation failed: %v", err)
		}
		log.Println("Migration completed successfully")
	} else {
		log.Println("[DRY-RUN] No changes made")
	}
}

func createBackup(dataPath string) error {
	timestamp := strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf("%s", time.Now()), ":", "-"), " ", "_")
	backupPath := filepath.Join(filepath.Dir(dataPath), fmt.Sprintf("data_backup_%s", timestamp))

	log.Printf("Creating backup: %s -> %s", dataPath, backupPath)

	return filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dataPath, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(backupPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func validateMigration(dataPath, casesPath string) error {
	// Check that cases folder has the migrated folders
	casesEntries, err := os.ReadDir(casesPath)
	if err != nil {
		return fmt.Errorf("failed to read cases directory: %v", err)
	}

	if len(casesEntries) == 0 {
		return fmt.Errorf("cases directory is empty after migration")
	}

	// Check that the old case folders are no longer in the root
	dataEntries, err := os.ReadDir(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %v", err)
	}

	pattern := regexp.MustCompile(caseFolderPattern)
	for _, entry := range dataEntries {
		if entry.IsDir() && pattern.MatchString(entry.Name()) {
			return fmt.Errorf("old case folder still exists: %s", entry.Name())
		}
	}

	log.Printf("Validation passed: %d case folders migrated to %s", len(casesEntries), casesPath)
	return nil
}
