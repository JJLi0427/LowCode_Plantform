package downloader

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"os"
	"net/http"
	"path"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func DownloadResource(url, saveto string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("sss", err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("sss")
		return "", fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	var respBody bytes.Buffer
	respBody.ReadFrom(resp.Body)

	mtype := mimetype.Detect(respBody.Bytes())
	h := md5.New()
	h.Write([]byte(url))
	fname := hex.EncodeToString(h.Sum(nil)) + mtype.Extension()

	os.WriteFile(path.Join(saveto, fname), respBody.Bytes(), 0644)

	return fname, nil
}
