package helpers

import (
	"net/http"

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
