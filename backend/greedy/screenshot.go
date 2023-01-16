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

func createScreenshot(s string) (string, error, []string) {

	var d []string // used for debugging afterwards

	// first download image using screenshot api
	target := fmt.Sprintf("https://screenshot.abstractapi.com/v1/?api_key=%s&url=%s", cfg.Settings.ScreenshotAPIToken, url.QueryEscape(s))
	d = logAndCollect(d, target, nil)

	resp, err := http.Get(target)
	if err != nil {
		return "", err, d
	}

	d = logAndCollect(d, fmt.Sprintf("api response code: %d", resp.StatusCode), nil)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err, d
	}

	d = logAndCollect(d, fmt.Sprintf("raw screenshot size: %s, api response: %s", humanize.Bytes(uint64(len(body))), resp.Status), nil)

	// resize image
	img, err := imaging.Decode(bytes.NewReader(body))
	if err != nil {
		return "", err, d
	}

	logrus.Info("start resize")
	dstImage := imaging.Resize(img, 600, 0, imaging.Lanczos)

	logrus.Info("start cropanchor")
	croppedImage := imaging.CropAnchor(dstImage, 600, 600, imaging.TopLeft)

	buf := new(bytes.Buffer)
	if err := imaging.Encode(buf, croppedImage, imaging.JPEG); err != nil {
		return "", err, d
	}

	resizedBytes, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err, d
	}

	// base64 encode
	// https://github.com/disintegration/imaging/issues/141

	str := base64.StdEncoding.EncodeToString(resizedBytes)
	d = logAndCollect(d, fmt.Sprintf("thumb string size: %s", humanize.Bytes(uint64(len(str)))), nil)

	return str, nil, d
}

// func logAndCollect(collection []string, s string, err error) []string {

// 	if len(s) != 0 {
// 		collection = append(collection, s)
// 		logrus.Info(s)
// 	}

// 	if err != nil {
// 		collection = append(collection, err.Error())
// 		logrus.Error(err)
// 	}

// 	logrus.Info
// 	return collection
// }

func logAndCollect(c []string, s string, e error) []string {

	if len(s) != 0 {
		c = append(c, s)
	}

	if e != nil {
		c = append(c, e.Error())
	}

	return c
}
