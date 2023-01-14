package greedy

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"
	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/sirupsen/logrus"
)

func createScreenshot(s string) (string, error) {

	// first download image using screenshot api
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

	logrus.Infof("raw screenshot size: %s, api response: %s", humanize.Bytes(uint64(len(body))), resp.Status)

	// resize image
	img, err := imaging.Decode(bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	dstImage := imaging.Resize(img, 600, 0, imaging.Lanczos)

	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, dstImage, imaging.JPEG); err != nil {
		return "", err
	}

	resizedBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}

	// base64 encode
	// https://github.com/disintegration/imaging/issues/141

	str := base64.StdEncoding.EncodeToString(resizedBytes)
	return str, nil
}
