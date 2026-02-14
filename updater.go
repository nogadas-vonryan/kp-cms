package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/inconshreveable/go-update"
)

const currentVersion = "0.1.3"
const repoOwner = "nogadas-vonryan"
const repoName = "kp-cms"

type GithubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func selfUpdate() {
	fmt.Printf("kpcms v%s (%s/%s)\n", currentVersion, runtime.GOOS, runtime.GOARCH)

	latest, err := getLatestFromList()
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		return
	}

	vCurrent, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Printf("Error parsing local version: %v\n", err)
		return
	}

	vLatest, err := semver.NewVersion(latest.TagName)
	if err != nil {
		fmt.Printf("Error parsing remote version (%s): %v\n", latest.TagName, err)
		return
	}

	if !vLatest.GreaterThan(vCurrent) {
		fmt.Println("No new updates found. You are up to date.")
		return
	}

	status := "Stable"
	if latest.Prerelease {
		status = "Pre-release"
	}
	fmt.Printf("Found newer %s version: %s\n", status, vLatest.String())

	downloadURL := ""
	for _, asset := range latest.Assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(runtime.GOOS)) {
			if runtime.GOOS != "windows" && !strings.Contains(asset.Name, runtime.GOARCH) {
				continue
			}
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		fmt.Println("No compatible binary found for your operating system.")
		return
	}

	fmt.Println("Downloading and applying update...")
	if err := runUpdate(downloadURL); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return
	}

	fmt.Println("Successfully updated! Please restart kpcms to apply changes.")
	os.Exit(0)
}

func getLatestFromList() (*GithubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var releases []GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}

	// Index 0 is the most recent release (by date)
	return &releases[0], nil
}

func runUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var binaryReader io.Reader
	var extractErr error

	if strings.HasSuffix(url, ".tar.gz") {
		binaryReader, extractErr = extractFromTarGz(resp.Body)
	} else if strings.HasSuffix(url, ".zip") {
		binaryReader, extractErr = extractFromZip(resp.Body)
	} else {
		return fmt.Errorf("unsupported file extension: %s", url)
	}

	if extractErr != nil {
		return extractErr
	}

	return update.Apply(binaryReader, update.Options{})
}

func extractFromTarGz(r io.Reader) (io.Reader, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag == tar.TypeReg && (header.Name == "kpcms" || strings.HasSuffix(header.Name, ".exe")) {
			return tr, nil
		}
	}
	return nil, fmt.Errorf("executable not found in tar.gz")
}

func extractFromZip(r io.Reader) (io.Reader, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	readerAt := bytes.NewReader(buf)
	zr, err := zip.NewReader(readerAt, int64(len(buf)))
	if err != nil {
		return nil, err
	}

	for _, f := range zr.File {
		if f.Name == "kpcms.exe" || f.Name == "kpcms" {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("executable not found in zip")
}
