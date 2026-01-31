package document

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hymkor/trash-go"
)

type BackupFile struct {
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *FileDocumentRepository) GetBackupPath() string {
	return r.backupPath
}

func (r *FileDocumentRepository) ListBackups(ctx context.Context) ([]BackupFile, error) {
	entries, err := os.ReadDir(r.backupPath)
	if err != nil {
		return nil, fmt.Errorf("read backup directory: %w", err)
	}

	backups := make([]BackupFile, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".zip") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			backups = append(backups, BackupFile{
				FileName:  entry.Name(),
				Size:      info.Size(),
				CreatedAt: info.ModTime(),
			})
		}
	}

	slices.SortFunc(backups, func(a, b BackupFile) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		return 0
	})

	return backups, nil
}

func (r *FileDocumentRepository) CreateBackup(ctx context.Context) (string, error) {
	r.mu.RLock()
	sourceDir := r.basePath
	backupDir := r.backupPath
	r.mu.RUnlock()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("backup_%s.zip", timestamp)
	fullOutputPath := filepath.Join(backupDir, fileName)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	outFile, err := os.Create(fullOutputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)

	resolvedSource, err := filepath.EvalSymlinks(sourceDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve source path: %w", err)
	}

	err = filepath.Walk(resolvedSource, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(resolvedSource, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
			_, err = zw.CreateHeader(header)
			return err
		}

		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		zw.Close()
		return "", fmt.Errorf("walk error during backup: %w", err)
	}

	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("failed to close zip writer: %w", err)
	}

	return fullOutputPath, nil
}

func (r *FileDocumentRepository) RestoreFromLocalPath(ctx context.Context, fileName string, overwrite bool, onProgress func(float64)) error {
	safeName := filepath.Base(fileName)
	filePath := filepath.Join(r.backupPath, safeName)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", safeName)
	}
	if info.IsDir() {
		return fmt.Errorf("requested path is a directory, not a backup file")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open backup file: %w", err)
	}
	defer file.Close()

	return r.ImportBackup(ctx, file, overwrite, onProgress)
}

// ProgressWriter tracks bytes written and executes a callback
type ProgressWriter struct {
	Total      int64
	Written    int64
	OnProgress func(float64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Written += int64(n)
	if pw.Total > 0 && pw.OnProgress != nil {
		percentage := (float64(pw.Written) / float64(pw.Total)) * 100
		pw.OnProgress(percentage)
	}
	return n, nil
}

func (r *FileDocumentRepository) ImportBackup(ctx context.Context, reader io.Reader, overwrite bool, onProgress func(float64)) error {
	// 1. Create a temporary file to store the incoming stream
	// We need a real file (not just a pipe) because zip.NewReader requires ReaderAt
	tempFile, err := os.CreateTemp("", "upload-backup-*.zip")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 2. Save the stream to the temp file
	size, err := io.Copy(tempFile, reader)
	if err != nil {
		return fmt.Errorf("save backup to disk: %w", err)
	}

	// 3. Open the ZIP reader
	zr, err := zip.NewReader(tempFile, size)
	if err != nil {
		return fmt.Errorf("open zip reader: %w", err)
	}

	// 4. Calculate total uncompressed size for accurate progress tracking
	var totalSize int64
	for _, f := range zr.File {
		if !f.FileInfo().IsDir() {
			totalSize += int64(f.UncompressedSize64)
		}
	}

	tracker := &ProgressWriter{
		Total:      totalSize,
		OnProgress: onProgress,
	}

	// 5. Extraction Phase
	// We wrap this in a function literal to ensure r.mu.Unlock() runs
	// BEFORE we trigger ReloadCache to prevent deadlocks.
	err = func() error {
		r.mu.Lock()
		defer r.mu.Unlock()

		if overwrite {
			// SAFETY CHECK: Ensure we aren't trashing the root directory
			if r.basePath != "" && r.basePath != "/" {
				if _, err := os.Stat(r.basePath); err == nil {
					// Assuming 'trash' is your package for safe deletion
					if err := trash.Throw(r.basePath); err != nil {
						return fmt.Errorf("failed to move current data to trash: %w", err)
					}
				}

				// Recreate the directory structure after trashing
				if err := os.MkdirAll(r.basePath, 0755); err != nil {
					return fmt.Errorf("failed to recreate base path: %w", err)
				}
			}
		}

		for _, f := range zr.File {
			// Security check: ZipSlip vulnerability prevention
			if strings.Contains(f.Name, "..") {
				continue
			}

			targetPath := filepath.Join(r.basePath, f.Name)

			if f.FileInfo().IsDir() {
				os.MkdirAll(targetPath, 0755)
				continue
			}

			// Ensure the parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			// Extract the individual file
			err := func() error {
				dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return err
				}
				defer dst.Close()

				src, err := f.Open()
				if err != nil {
					return err
				}
				defer src.Close()

				// io.MultiWriter writes to the file AND the tracker simultaneously
				multi := io.MultiWriter(dst, tracker)
				_, err = io.Copy(multi, src)
				return err
			}()

			if err != nil {
				return fmt.Errorf("failed to extract %s: %w", f.Name, err)
			}
		}
		return nil
	}()

	if err != nil {
		return err
	}

	// 6. Reload Cache
	// This is now safe to call because the Mutex has been released.
	_, err = r.ReloadCache(ctx)
	if err != nil {
		return fmt.Errorf("extraction succeeded but cache reload failed: %w", err)
	}

	return nil
}
