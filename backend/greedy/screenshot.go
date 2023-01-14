package greedy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func createScreenshot(s string) (string, error) {

	target := fmt.Sprintf("https://screenshot.abstractapi.com/v1/?api_key=4ab7e5e662114f27b8c4c7ce9669d3d8&url=%s", url.QueryEscape(s))

	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	str := base64.StdEncoding.EncodeToString(body)
	return str, nil
}
