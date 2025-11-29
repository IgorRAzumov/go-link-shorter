package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

const (
	ContentEncoding = "Content-Encoding"
	AcceptEncoding  = "Accept-Encoding"
	HeaderCode      = "gzip"
)

type compressWriter struct {
	writer     http.ResponseWriter
	GZIPWriter *gzip.Writer
}

func newCompressWriter(writer http.ResponseWriter) *compressWriter {
	return &compressWriter{
		writer:     writer,
		GZIPWriter: gzip.NewWriter(writer),
	}
}

func (compressWriter *compressWriter) Header() http.Header {
	return compressWriter.writer.Header()
}

func (compressWriter *compressWriter) Write(p []byte) (int, error) {
	return compressWriter.GZIPWriter.Write(p)
}

func (compressWriter *compressWriter) WriteHeader(statusCode int) {
	// Устанавливаем Content-Encoding только для успешных кодов (200-299)
	if statusCode >= 200 && statusCode < 300 {
		compressWriter.writer.Header().Set(ContentEncoding, HeaderCode)
	}
	compressWriter.writer.WriteHeader(statusCode)
}

func (compressWriter *compressWriter) Close() error {
	return compressWriter.GZIPWriter.Close()
}

type compressReader struct {
	readCloser io.ReadCloser
	GZIPReader *gzip.Reader
}

func newCompressReader(readCloser io.ReadCloser) (*compressReader, error) {
	GZIPReader, err := gzip.NewReader(readCloser)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		readCloser: readCloser,
		GZIPReader: GZIPReader,
	}, nil
}

func (compressReader *compressReader) Read(bytes []byte) (count int, err error) {
	return compressReader.GZIPReader.Read(bytes)
}

func (compressReader *compressReader) Close() error {
	if err := compressReader.readCloser.Close(); err != nil {
		return err
	}
	return compressReader.GZIPReader.Close()
}
