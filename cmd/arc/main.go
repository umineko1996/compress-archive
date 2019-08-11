package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/umineko1996/compress-archive/archive"
)

const (
	success = iota
	failed
)

func main() {
	os.Exit(Run())
}

func Run() int {
	srcPath, dstPath := getArg()

	arcFile, err := archive.Create(dstPath)
	if err != nil {
		fmt.Printf("archive file create error: %s\n", err.Error())
		return failed
	}
	defer arcFile.Close()

	offsetPath := filepath.Dir(srcPath)
	if err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("src filepath walk error: %s\n", err.Error())
			return err
		}
		if info.IsDir() {
			return nil
		}
		dstName, err := filepath.Rel(offsetPath, path)
		if err != nil {
			fmt.Printf("dstname relative error: %s\n", err.Error())
			return err
		}
		return arcFile.Append(dstName, path)
	}); err != nil {
		fmt.Printf("archive file append src file error: %s\n", err.Error())
		return failed
	}

	return success
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
