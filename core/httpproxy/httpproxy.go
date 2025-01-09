package httpproxy

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const BaseHttpRedirectProxyURL string = "/gapplications/commonhproxy/link"

var notFollowRedirectClient *http.Client
var followRedirectClient *http.Client

func init() {
	notFollowRedirectClient = newHttpClient(false)
	followRedirectClient = newHttpClient(true)
}

func newHttpClient(followRedirect bool) *http.Client {
	nofollowRedirect := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	if followRedirect {
		nofollowRedirect = nil
	}

	client := &http.Client{CheckRedirect: nofollowRedirect}
	return client
}

func GenerateThirdParyHttpRedirectProxyURL(targetUrl string, followRedirect bool) string {
	if strings.HasPrefix(targetUrl, "http") || strings.HasPrefix(targetUrl, "//") {
		targetUrl = base64.URLEncoding.EncodeToString([]byte(targetUrl))
		if followRedirect {
			return fmt.Sprintf("%s?url=%s&code=%d&redirect=%v", BaseHttpRedirectProxyURL, targetUrl, len(targetUrl)*75%1000, followRedirect)
		} else {
			return fmt.Sprintf("%s?url=%s&code=%d", BaseHttpRedirectProxyURL, targetUrl, len(targetUrl)*75%1000)
		}
	} else {
		return targetUrl
	}
}

func ParserThirdParyHttpRedirectProxyURL(gurl *url.URL) (string, bool) {
	eurl, code := gurl.Query().Get("url"), gurl.Query().Get("code")
	if code != strconv.Itoa(len(eurl)*75%1000) {
		return "", false
	}

	if proxyurl, err := base64.URLEncoding.DecodeString(eurl); err == nil && len(proxyurl) > 0 {
		return string(proxyurl), true
	} else {
		return "", false
	}
}

func IsThirdParyHttpRedirectProxyURLRequest(gurl *url.URL) bool {
	if gurl == nil {
		return false
	}

	return gurl.Path == BaseHttpRedirectProxyURL
}

func HttpProxy(w http.ResponseWriter, r *http.Request, targetUrl string) {
	if r.Method == http.MethodGet || r.Method == http.MethodPost {
		if req, err := http.NewRequest(r.Method, targetUrl, r.Body); err == nil {
			for k, v := range r.Header {
				if k == "Accept" ||
					k == "Accept-Encoding" ||
					k == "Accept-Language" ||
					k == "Cache-Control" ||
					k == "Authorization" ||
					k == "User-Agent" {
					req.Header[k] = v
				}
			}

			client := notFollowRedirectClient
			if r.URL.Query().Get("redirect") == "true" {
				client = followRedirectClient
			}

			if res, err := client.Do(req); err == nil {
				for k, v := range res.Header {
					if k != "Content-Length" {
						w.Header().Set(k, strings.Join(v, ","))
					}
				}
				w.WriteHeader(res.StatusCode)
				if res.Body != nil {
					defer res.Body.Close()
					io.Copy(w, res.Body)
				}
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}
