package greedy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	cfg "github.com/rogierlommers/quick-note/backend/config"
)

func createScreenshot(s string) (string, error) {

	target := fmt.Sprintf("https://screenshot.abstractapi.com/v1/?api_key=%s&url=%s", cfg.Settings.ScreenshotAPIToken, url.QueryEscape(s))

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
