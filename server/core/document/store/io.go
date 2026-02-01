package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kpcms/server/core/document"
	"os"
	"path/filepath"
	"strings"

	"github.com/hymkor/trash-go"
)

func (r *Store) AddFileMetadata(ctx context.Context, uuid string, file document.File) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := r.getDocumentPath(doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []document.File
	data, err := os.ReadFile(filesJSONPath)
	if err == nil {
		if err := json.Unmarshal(data, &files); err != nil {
			return fmt.Errorf("decode files metadata: %w", err)
		}
	}

	for i, f := range files {
		if f.FileName == file.FileName {
			files[i] = file
			return writeFilesMetadata(filesJSONPath, files)
		}
	}

	files = append(files, file)
	return writeFilesMetadata(filesJSONPath, files)
}

func (r *Store) UpdateFileMetadata(ctx context.Context, uuid string, fileName string, updates document.FileMetadataUpdate) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := r.getDocumentPath(doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []document.File
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

func (r *Store) scanPhysicalFolder(folderPath string) ([]document.File, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var files []document.File
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "meta.json" || entry.Name() == "files.json" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, document.File{
			FileName:  entry.Name(),
			Type:      filepath.Ext(entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}
	return files, nil
}

func (r *Store) readFiles(folderName string) ([]document.File, error) {
	folderPath := r.getDocumentPath(folderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	physicalFiles, err := r.scanPhysicalFolder(folderPath)
	if err != nil {
		return nil, fmt.Errorf("scanning physical folder: %w", err)
	}

	var metadataMap = make(map[string]document.File)
	jsonData, err := os.ReadFile(filesJSONPath)
	if err == nil {
		var metadataList []document.File
		if err := json.Unmarshal(jsonData, &metadataList); err == nil {
			for _, f := range metadataList {
				metadataMap[f.FileName] = f
			}
		}
	}

	syncedFiles := make([]document.File, 0, len(physicalFiles))
	newFilesFound := false
	for _, physFile := range physicalFiles {
		if meta, exists := metadataMap[physFile.FileName]; exists {
			meta.Size = physFile.Size
			meta.CreatedAt = physFile.CreatedAt
			meta.Type = physFile.Type
			syncedFiles = append(syncedFiles, meta)
		} else {
			syncedFiles = append(syncedFiles, physFile)
			newFilesFound = true
		}
	}

	// If we found new files, automatically sync them to files.json
	if newFilesFound {
		if err := writeFilesMetadata(filesJSONPath, syncedFiles); err != nil {
			// Log the error but don't fail the read operation
			// This allows the system to continue working even if sync fails
			fmt.Fprintf(os.Stderr, "warning: failed to sync new files to %s: %v\n", filesJSONPath, err)
		}
	}

	return syncedFiles, nil
}

func (r *Store) DownloadFile(ctx context.Context, uuid string, fileName string) (io.ReadCloser, error) {
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

func (r *Store) UploadFile(ctx context.Context, uuid string, fileName string, content io.Reader) error {
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

	folderPath := r.getDocumentPath(doc.FolderName)
	finalName, err := r.uniqueFileName(folderPath, safeName)
	if err != nil {
		return fmt.Errorf("determine unique filename: %w", err)
	}

	filePath := r.getDocumentFilePath(doc.FolderName, finalName)

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

	file := document.File{
		FileName:  finalName,
		Type:      filepath.Ext(finalName),
		Size:      writtenBytes,
		CreatedAt: fileInfo.ModTime(),
	}

	return r.AddFileMetadata(ctx, uuid, file)
}

func (r *Store) DeleteFile(ctx context.Context, uuid string, fileName string) error {
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

	if err := trash.Throw(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("move file to trash: %w", err)
	}

	filesJSONPath := filepath.Join(folderPath, "files.json")
	var files []document.File
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

// uniqueFileName finds a free filename in folderPath by appending " (Copy)" / " (Copy N)"
func (r *Store) uniqueFileName(folderPath, baseName string) (string, error) {
	if baseName == "" {
		return "", errors.New("empty base name")
	}

	try := baseName
	name := baseName
	ext := filepath.Ext(baseName)
	prefix := strings.TrimSuffix(baseName, ext)

	i := 0
	for {
		fullPath := filepath.Join(folderPath, try)
		_, err := os.Stat(fullPath)
		if errors.Is(err, os.ErrNotExist) {
			return try, nil
		}
		if err != nil {
			return "", err
		}
		// exists -> prepare next
		i++
		if i == 1 {
			name = fmt.Sprintf("%s (Copy)%s", prefix, ext)
		} else {
			name = fmt.Sprintf("%s (Copy %d)%s", prefix, i-1, ext)
		}
		try = name
	}
}

// UpdateFileContents overwrites an existing file's content and updates its metadata (preserves Description/Note/Tags)
func (r *Store) UpdateFileContents(ctx context.Context, uuid string, fileName string, content io.Reader) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if fileName == "" {
		return errors.New("file name is required")
	}
	if content == nil {
		return errors.New("file content is required")
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

	filePath := r.getDocumentFilePath(doc.FolderName, safeName)

	// Overwrite file (must already exist or we create it — preserve metadata if present)
	outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("open file for update: %w", err)
	}
	defer outFile.Close()

	writtenBytes, err := io.Copy(outFile, content)
	if err != nil {
		return fmt.Errorf("write updated content: %w", err)
	}

	info, err := outFile.Stat()
	if err != nil {
		return fmt.Errorf("stat updated file: %w", err)
	}

	// Update files.json (preserve custom fields)
	folderPath := r.getDocumentPath(doc.FolderName)
	filesJSONPath := filepath.Join(folderPath, "files.json")

	var files []document.File
	data, err := os.ReadFile(filesJSONPath)
	if err == nil {
		_ = json.Unmarshal(data, &files) // ignore unmarshal error and proceed
	}

	found := false
	for i := range files {
		if files[i].FileName == safeName {
			files[i].Size = writtenBytes
			files[i].CreatedAt = info.ModTime()
			files[i].Type = filepath.Ext(safeName)
			found = true
			break
		}
	}
	if !found {
		files = append(files, document.File{
			FileName:  safeName,
			Type:      filepath.Ext(safeName),
			Size:      writtenBytes,
			CreatedAt: info.ModTime(),
		})
	}

	return writeFilesMetadata(filesJSONPath, files)
}

// RenameFile renames a file on disk and updates files.json (preserves Description/Note/Tags)
func (r *Store) RenameFile(ctx context.Context, uuid string, oldName string, newName string) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if oldName == "" || newName == "" {
		return errors.New("old and new file names are required")
	}

	oldSafe := filepath.Base(oldName)
	newSafe := filepath.Base(newName)

	hasPathTraversal := oldSafe != oldName || newSafe != newName ||
		oldSafe == "." || oldSafe == ".." || newSafe == "." || newSafe == ".." ||
		filepath.IsAbs(oldName) || filepath.IsAbs(newName) ||
		strings.ContainsAny(oldName, `/\`) || strings.ContainsAny(newName, `/\`)

	if hasPathTraversal {
		return errors.New("invalid file name: path traversal detected")
	}

	doc, err := r.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	folderPath := r.getDocumentPath(doc.FolderName)
	oldPath := filepath.Join(folderPath, oldSafe)
	// ensure source exists
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("source file does not exist: %w", err)
	}

	// Resolve target name if conflict
	finalName, err := r.uniqueFileName(folderPath, newSafe)
	if err != nil {
		return fmt.Errorf("determine target filename: %w", err)
	}
	newPath := filepath.Join(folderPath, finalName)

	// Perform rename
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename file: %w", err)
	}

	// Update metadata
	filesJSONPath := filepath.Join(folderPath, "files.json")
	var files []document.File
	data, err := os.ReadFile(filesJSONPath)
	if err == nil {
		_ = json.Unmarshal(data, &files) // ignore unmarshal error and rebuild if needed
	}

	info, err := os.Stat(newPath)
	if err != nil {
		return fmt.Errorf("stat renamed file: %w", err)
	}

	// Find old metadata entry, move/persist it under new name
	found := false
	for i := range files {
		if files[i].FileName == oldSafe {
			files[i].FileName = finalName
			files[i].Type = filepath.Ext(finalName)
			if info != nil {
				files[i].Size = info.Size()
				files[i].CreatedAt = info.ModTime()
			}
			found = true
			break
		}
	}
	if !found {
		// create new metadata entry if none existed
		f := document.File{
			FileName: finalName,
			Type:     filepath.Ext(finalName),
		}
		if info != nil {
			f.Size = info.Size()
			f.CreatedAt = info.ModTime()
		}
		files = append(files, f)
	}

	return writeFilesMetadata(filesJSONPath, files)
}
