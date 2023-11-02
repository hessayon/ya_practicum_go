package compressing

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}


func IsGzipContentType(contentType string) bool{
	return contentType == "application/json" || contentType == "text/html"
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (compressWr *compressWriter) Header() http.Header {
	return compressWr.w.Header()
}

func (compressWr *compressWriter) Write(p []byte) (int, error) {
	if IsGzipContentType(compressWr.Header().Get("Content-Type")) {
		return compressWr.zw.Write(p)
	}
	return compressWr.w.Write(p)
}

func (compressWr *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 400  && IsGzipContentType(compressWr.Header().Get("Content-Type")){
		compressWr.w.Header().Set("Content-Encoding", "gzip")
	}
	compressWr.w.WriteHeader(statusCode)
}


func (compressWr *compressWriter) Close() error {
	return compressWr.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (compressR compressReader) Read(p []byte) (n int, err error) {
	return compressR.zr.Read(p)
}

func (compressR *compressReader) Close() error {
	if err := compressR.r.Close(); err != nil {
		return err
	}
	return compressR.zr.Close()
}

func CheckSupportOfGzip(encodingList []string) bool {
	for _, value := range encodingList {
		acceptEncodings := strings.Split(value, ",")
		for _, encoding := range acceptEncodings {
			if strings.Contains(strings.Split(encoding, ";")[0], "gzip") {
				return true
			}
		}
	}
	return false
}

