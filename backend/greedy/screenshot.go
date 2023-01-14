package greedy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dustin/go-humanize"
	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/sirupsen/logrus"
)

func createScreenshot(s string) (string, error) {

	target := fmt.Sprintf("https://screenshot.abstractapi.com/v1/?api_key=%s&url=%s", cfg.Settings.ScreenshotAPIToken, url.QueryEscape(s))
	logrus.Infof("target: %s", target)

	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}

	logrus.Infof("api response code: %d", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	logrus.Infof("screenshot size: %s, api response: %d", humanize.Bytes(uint64(len(body))), resp.Status)
	str := base64.StdEncoding.EncodeToString(body)
	return str, nil
}
