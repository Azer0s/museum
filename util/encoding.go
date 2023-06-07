package util

import (
	"bytes"
	"compress/gzip"
	"io"
)

func DecodeBody(body []byte, encoding string) (b []byte, err error) {
	switch encoding {
	case "gzip":
		return decodeGzip(body)
	case "deflate":
		panic("deflate not implemented")
	case "compress":
		panic("compress not implemented")
	case "br":
		panic("br not implemented")
	default:
		return body, nil
	}
}

func decodeGzip(body []byte) (b []byte, err error) {
	bufferedReader := bytes.NewReader(body)

	var gzipReader *gzip.Reader
	gzipReader, err = gzip.NewReader(bufferedReader)
	if err != nil {
		return nil, err
	}
	defer func(gzipReader *gzip.Reader) {
		err = gzipReader.Close()
	}(gzipReader)

	return io.ReadAll(gzipReader)
}

func EncodeBody(body []byte, encoding string) (b []byte, err error) {
	switch encoding {
	case "gzip":
		return encodeGzip(body)
	default:
		return body, nil
	}
}

func encodeGzip(body []byte) (b []byte, err error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)

	_, err = gzipWriter.Write(body)
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
