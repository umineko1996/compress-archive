package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getFileCh(ctx context.Context, path string) (<-chan string, error) {
	file, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !file.IsDir() {
		ch := make(chan string, 1)
		defer close(ch)
		ch <- path
		return ch, nil
	}

	fileCh := make(chan string, 10)
	go func() {
		defer close(fileCh)
		dirwalk(ctx, fileCh, path)
	}()
	return fileCh, nil
}

func dirwalk(ctx context.Context, fileCh chan<- string, srcPath string) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		path := filepath.Join(srcPath, file.Name())
		if file.IsDir() {
			dirwalk(ctx, fileCh, path)
			continue
		}
		fileCh <- path
	}
}
