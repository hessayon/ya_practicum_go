package compressing

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func isGzipContentType(contentType string) bool{
	return contentType == "application/json" || contentType == "text/html"
}
// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (compressWr *compressWriter) Header() http.Header {
	return compressWr.w.Header()
}

func (compressWr *compressWriter) Write(p []byte) (int, error) {
	if isGzipContentType(compressWr.Header().Get("Content-Type")) {
		return compressWr.zw.Write(p)
	}
	return compressWr.w.Write(p)
}

func (compressWr *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 400  && isGzipContentType(compressWr.Header().Get("Content-Type")){
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

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
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

func checkSupportOfGzip(encodingList []string) bool {
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

func GzipCompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		currentWriter := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		supportsGzip := checkSupportOfGzip(r.Header.Values("Accept-Encoding"))

		if isGzipContentType(r.Header.Get("Content-Type")) && supportsGzip{
			compressWr := newCompressWriter(w)
			currentWriter = compressWr
			defer compressWr.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		sendsGzip := checkSupportOfGzip(r.Header.Values("Content-Encoding"))
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(currentWriter, r)
	}
}
