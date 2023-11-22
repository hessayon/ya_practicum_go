package middleware

import (
	"net/http"

	"github.com/hessayon/ya_practicum_go/internal/compressing"

)



func GzipCompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		currentWriter := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		supportsGzip := compressing.CheckSupportOfGzip(r.Header.Values("Accept-Encoding"))

		if compressing.IsGzipContentType(r.Header.Get("Content-Type")) && supportsGzip{
			compressWr := compressing.NewCompressWriter(w)
			currentWriter = compressWr
			defer compressWr.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		sendsGzip := compressing.CheckSupportOfGzip(r.Header.Values("Content-Encoding"))
		if sendsGzip {
			cr, err := compressing.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h(currentWriter, r)
	}
}


