package main

import (
	"flag"
	"fmt"
	"lowcode/core"
	"time"
)

func main() {
	apppath := flag.String("apppath", "apps", "input directory containing applications")
	debug := flag.String("debug", "", "input appname, the app's index.html will be update automatly, if anything modified")
	port := flag.String("port", "", "[http] listening port")
	tlsport := flag.String("tlsport", "", "[https] listening port")
	httpRedirect := flag.Bool("httpRedirect", false, "Redirect http request to https")
	flag.Parse()

	if len(*debug) > 0 {
		fmt.Println("now staring as DEBUG mode...")
	}

	startTime := time.Now()
	//set default http port
	if len(*port) == 0 && len(*tlsport) == 0 {
		*port = ":8088"
	}

	httpd := &core.Httpd{Port: *port,
		TlsPort:             *tlsport,
		AppsControlRootPath: *apppath,
		DebugAppname:        *debug,
		SiteDomainName:      "",
		CACert:              false,
		HttpRedirect:        *httpRedirect,
	}

	if err := httpd.Init(); err != nil {
		fmt.Println(err)
	}
	
	initTime := time.Since(startTime)
    fmt.Printf("Server initialized in %v\n", initTime)

	if err := httpd.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
