package helpers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func DoGetGoQuery(client *http.Client, api_url string) (*goquery.Document, error) {
	resp, err := client.Get(api_url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, qerr := goquery.NewDocumentFromReader(resp.Body)

	if qerr != nil {
		return nil, qerr
	}

	return doc, nil
}

func CompressString(s string) (string, error) {
	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write([]byte(s)); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func CreateFolders() error {
	return os.MkdirAll(".cache", os.ModePerm)
}
