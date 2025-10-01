package main

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type FileItem struct {
	Path    string
	Name    string
	IsDir   bool
	Size    int64
	Keep    bool
	Decided bool
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
		
		item := FileItem{
			Path:    path,
			Name:    d.Name(),
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			Keep:    false,
			Decided: false,
		}
		
		items = append(items, item)
		
		if d.IsDir() {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return items, err
}
