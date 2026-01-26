package document

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (r *FileDocumentRepository) AddFileMetadata(ctx context.Context, uuid string, file File) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := r.getDocumentPath(doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []File
	data, err := os.ReadFile(filesJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &files); err != nil {
			return fmt.Errorf("decode files metadata: %w", err)
		}
	}

	// Check if file already exists
	for i, f := range files {
		if f.FileName == file.FileName {
			// Update existing file metadata
			files[i] = file
			return writeFilesMetadata(filesJSONPath, files)
		}
	}

	files = append(files, file)
	return writeFilesMetadata(filesJSONPath, files)
}

func (r *FileDocumentRepository) UpdateFileMetadata(ctx context.Context, uuid string, fileName string, updates FileMetadataUpdate) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := r.getDocumentPath(doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	// Read existing metadata
	var files []File
	data, err := os.ReadFile(filesJSONPath)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}
	if err := json.Unmarshal(data, &files); err != nil {
		return err
	}

	// Find and update
	found := false
	for i := range files {
		if files[i].FileName == fileName {
			if updates.Description != nil {
				files[i].Description = *updates.Description
			}
			if updates.Note != nil {
				files[i].Note = *updates.Note
			}
			if updates.Tags != nil {
				files[i].Tags = *updates.Tags
			}
			found = true
			break
		}
	}

	if !found {
		return errors.New("file metadata not found")
	}

	return writeFilesMetadata(filesJSONPath, files)
}

func (r *FileDocumentRepository) scanPhysicalFolder(folderPath string) ([]File, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var files []File
	for _, entry := range entries {
		// Ignore directories and the metadata file itself
		if entry.IsDir() || entry.Name() == "meta.json" || entry.Name() == "files.json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, File{
			FileName:  entry.Name(),
			Type:      filepath.Ext(entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			// Description and Note will be empty because we are recovering from raw files
		})
	}
	return files, nil
}

func (r *FileDocumentRepository) readFiles(folderName string) ([]File, error) {
	folderPath := r.getDocumentPath(folderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	physicalFiles, err := r.scanPhysicalFolder(folderPath)
	if err != nil {
		return nil, fmt.Errorf("scanning physical folder: %w", err)
	}

	// Try to read the existing metadata
	var metadataMap = make(map[string]File)
	jsonData, err := os.ReadFile(filesJSONPath)
	if err == nil {
		var metadataList []File
		if err := json.Unmarshal(jsonData, &metadataList); err == nil {
			for _, f := range metadataList {
				metadataMap[f.FileName] = f
			}
		}
	}

	// Loop through physical files and attach metadata if it exists
	syncedFiles := make([]File, 0, len(physicalFiles))
	for _, physFile := range physicalFiles {
		if meta, exists := metadataMap[physFile.FileName]; exists {
			// Keep existing metadata (like custom tags/names)
			// but update physical stats from the scan
			meta.Size = physFile.Size
			meta.CreatedAt = physFile.CreatedAt
			meta.Type = physFile.Type
			syncedFiles = append(syncedFiles, meta)
		} else {
			// It's a brand new file found on disk
			syncedFiles = append(syncedFiles, physFile)
		}
	}

	return syncedFiles, nil
}

func (r *FileDocumentRepository) DownloadFile(ctx context.Context, uuid string, fileName string) (io.ReadCloser, error) {
	if err := ctxErr(ctx); err != nil {
		return nil, err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// Path traversal protection is handled by getDocumentFilePath using filepath.Base
	filePath := r.getDocumentFilePath(doc.FolderName, filepath.Base(fileName))

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file for download: %w", err)
	}

	return file, nil
}

func (r *FileDocumentRepository) UploadFile(ctx context.Context, uuid string, fileName string, content io.Reader) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if fileName == "" {
		return errors.New("file name is required")
	}
	if content == nil {
		return errors.New("file content is required")
	}

	// Sanitize filename to prevent path traversal attacks
	safeName := filepath.Base(fileName)
	// Check for path traversal attempts (both Unix and Windows styles)
	hasPathTraversal := safeName != fileName ||
		safeName == "." || safeName == ".." ||
		filepath.IsAbs(fileName) ||
		strings.ContainsAny(fileName, `/\`)

	if hasPathTraversal {
		return errors.New("invalid file name: path traversal detected")
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("fetching document: %w", err)
	}

	filePath := r.getDocumentFilePath(doc.FolderName, safeName)

	outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer outFile.Close()

	writtenBytes, err := io.Copy(outFile, content)
	if err != nil {
		return fmt.Errorf("write file content: %w", err)
	}

	fileInfo, err := outFile.Stat()
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	file := File{
		FileName:  safeName,
		Type:      filepath.Ext(safeName),
		Size:      writtenBytes,
		CreatedAt: fileInfo.ModTime(),
	}

	return r.AddFileMetadata(ctx, uuid, file)
}

func (r *FileDocumentRepository) DeleteFile(ctx context.Context, uuid string, fileName string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	safeName := filepath.Base(fileName)
	hasPathTraversal := safeName != fileName ||
		safeName == "." || safeName == ".." ||
		filepath.IsAbs(fileName) ||
		strings.ContainsAny(fileName, `/\`)

	if hasPathTraversal {
		return errors.New("invalid file name: path traversal detected")
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := r.getDocumentPath(doc.FolderName)
	filePath := filepath.Join(folderPath, safeName)

	// Delete physical file
	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete physical file: %w", err)
	}

	// Update files.json
	filesJSONPath := filepath.Join(folderPath, "files.json")
	var files []File
	data, err := os.ReadFile(filesJSONPath)

	// Parse metadata if it exists
	if err == nil {
		if err := json.Unmarshal(data, &files); err != nil {
			return fmt.Errorf("decode files metadata: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// If there's a real error (not just missing file), return it
		return fmt.Errorf("read files metadata: %w", err)
	}
	// If file doesn't exist, continue with empty files slice

	// Remove file from metadata
	for i, f := range files {
		if f.FileName == safeName {
			files = append(files[:i], files[i+1:]...)
			break
		}
	}

	// Always write back files.json (even if file wasn't in metadata)
	return writeFilesMetadata(filesJSONPath, files)
}
