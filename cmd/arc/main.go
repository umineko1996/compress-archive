package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/umineko1996/compress-archive/archive"
)

func main() {
	srcPath, dstPath := getArg()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fileList, err := getFileCh(ctx, srcPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	arcFile, err := archive.Create(dstPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer arcFile.Close()

	offsetPath := filepath.Dir(srcPath)
	for srcFilePath := range fileList {
		dstFilePath, err := filepath.Rel(offsetPath, srcFilePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := arcFile.Append(dstFilePath, srcFilePath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func getArg() (src, dst string) {
	zipFlag := flag.Bool("zip", false, "archive zip flag")
	tarFlag := flag.Bool("tar", false, "archive tar flag")
	dstPath := flag.String("o", "", "out filename")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("usage: archive [-zip | -tar] {-o dst} src")
		os.Exit(1)
	}
	zf := *zipFlag
	tf := *tarFlag
	switch {
	case zf && tf:
		// デフォルト動作に従う
	case zf:
		archive.Mode = archive.ZIP
	case tf:
		archive.Mode = archive.TAR
	}
	args := flag.Args()
	src = args[0]
	dst = *dstPath
	if dst == "" {
		dst = filepath.Base(src)
	}
	return src, dst
}
