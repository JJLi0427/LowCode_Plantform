package appplugin

import (
	"net/http"
)

type AppPlugin interface {
	GetAppName() string
	GetUrlRoot() string
	GetAccessableUrls() []string

	GetEnvs() []string
	GetControlFile() string
	Changed() bool
	GetHttpHandler() (string, http.HandlerFunc, error)
}
