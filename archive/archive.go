package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type archiveType int

var Mode archiveType

const (
	UNKNOWN archiveType = iota
	ZIP
	TAR
)

func init() {
	switch runtime.GOOS {
	case "windows":
		Mode = ZIP
	case "linux":
		Mode = TAR
	default:
		Mode = UNKNOWN
	}
}

type File struct {
	w arcWriter
}

func Create(filename string) (*File, error) {
	var arcName string
	var factory arcWriterFactory
	switch Mode {
	case ZIP:
		arcName = filename + ".zip"
		factory = arcWriterFactoryFunc(newZipWriter)
	case TAR:
		arcName = filename + ".tar.gz"
		factory = arcWriterFactoryFunc(newTarWriter)
	default:
		return nil, errors.New("unknown mode selected")
	}

	f, err := os.Create(arcName)
	if err != nil {
		return nil, err
	}

	writer := factory.NewWriter(f)

	return &File{w: writer}, nil
}

func (f File) Append(dstFileName, srcFilePath string) error {
	fileInfo, err := os.Stat(srcFilePath)
	if err != nil {
		return err
	}
	w, err := f.w.Create(dstFileName, fileInfo)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return nil
	}
	r, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer r.Close()

	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return nil
}

func (f File) Close() error {
	return f.w.Close()
}

type arcWriterFactoryFunc func(w io.Writer) arcWriter

func (f arcWriterFactoryFunc) NewWriter(w io.Writer) arcWriter {
	return f(w)
}

type arcWriterFactory interface {
	NewWriter(w io.Writer) arcWriter
}

type arcWriter interface {
	Create(filepath string, fileInfo os.FileInfo) (io.Writer, error)
	Close() error
}

type zipWriter struct {
	*zip.Writer
}

func newZipWriter(w io.Writer) arcWriter {
	zw := zip.NewWriter(w)
	return &zipWriter{Writer: zw}
}

func (zw zipWriter) Create(path string, fileInfo os.FileInfo) (io.Writer, error) {
	if fileInfo.IsDir() {
		return nil, nil
	}

	path = filepath.ToSlash(path)
	fmt.Println(path)

	fh, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return nil, err
	}

	return zw.Writer.CreateHeader(fh)
}

type tarWriter struct {
	*tar.Writer
	gzw *gzip.Writer
}

func newTarWriter(w io.Writer) arcWriter {
	gw := gzip.NewWriter(w)
	tw := tar.NewWriter(gw)
	return &tarWriter{Writer: tw, gzw: gw}
}

func (tw tarWriter) Create(path string, fileInfo os.FileInfo) (io.Writer, error) {
	path = filepath.ToSlash(path)
	fmt.Println(path)

	hdr, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		hdr.Name = path + "/"
	} else {
		hdr.Name = path
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	return tw.Writer, nil
}

func (tw tarWriter) Close() error {
	if err := tw.Writer.Close(); err != nil {
		return err
	}
	return tw.gzw.Close()
}
