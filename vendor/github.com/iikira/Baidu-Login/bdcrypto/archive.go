package bdcrypto

import (
	"compress/gzip"
	"io"
	"os"
)

// GZIPCompress GZIP 压缩
func GZIPCompress(src io.Reader, writeTo io.Writer) (err error) {
	w := gzip.NewWriter(writeTo)
	_, err = io.Copy(w, src)
	if err != nil {
		return
	}

	w.Flush()
	return w.Close()
}

// GZIPUncompress GZIP 解压缩
func GZIPUncompress(src io.Reader, writeTo io.Writer) (err error) {
	unReader, err := gzip.NewReader(src)
	if err != nil {
		return err
	}

	_, err = io.Copy(writeTo, unReader)
	if err != nil {
		return
	}

	return unReader.Close()
}

// GZIPCompressFile GZIP 压缩文件
func GZIPCompressFile(filePath string) (err error) {
	return gzipCompressFile("en", filePath)
}

// GZIPUnompressFile GZIP 解压缩文件
func GZIPUnompressFile(filePath string) (err error) {
	return gzipCompressFile("de", filePath)
}

func gzipCompressFile(op, filePath string) (err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return
	}

	f2, err := os.Create(filePath + ".gzip.tmp")
	if err != nil {
		f.Close()
		return
	}

	defer f2.Close()

	switch op {
	case "en":
		err = GZIPCompress(f, f2)
	case "de":
		err = GZIPUncompress(f, f2)
	default:
		panic("unknown op" + op)
	}

	if err != nil {
		os.Remove(f2.Name())
		return
	}

	f.Close()
	os.Remove(filePath)
	return os.Rename(filePath+".gzip.tmp", filePath)
}
