package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileItem struct {
	Path     string
	Name     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	Preview  string
	Keep     bool
	Decided  bool
	Skipped  bool
}

func scanDirectory(dir string) ([]FileItem, error) {
	var items []FileItem
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if path == dir {
			return nil
		}
		
		info, err := d.Info()
		if err != nil {
			return err
		}
		
		relPath, _ := filepath.Rel(dir, path)
		if strings.HasPrefix(relPath, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		
		preview := ""
		if !d.IsDir() && info.Size() < 10240 { // Only preview files < 10KB
			preview = getFilePreview(path)
		}

		item := FileItem{
			Path:    path,
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Preview: preview,
			Keep:    false,
			Decided: false,
			Skipped: false,
		}
		
		items = append(items, item)
		
		if d.IsDir() {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return items, err
}

func getFilePreview(path string) string {
	if !isTextFile(path) {
		return ""
	}

	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	lineCount := 0

	for scanner.Scan() && lineCount < 3 {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
			lineCount++
		}
	}

	if len(lines) == 0 {
		return ""
	}

	preview := strings.Join(lines, "\n")
	if len(preview) > 150 {
		preview = preview[:147] + "..."
	}

	return preview
}

func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := []string{
		".txt", ".md", ".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h",
		".css", ".html", ".xml", ".json", ".yaml", ".yml", ".toml", ".ini",
		".sh", ".bat", ".ps1", ".sql", ".r", ".php", ".rb", ".scala", ".kt",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}

	return false
}
