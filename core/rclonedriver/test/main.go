package main

import (
	"context"
	"fmt"
	"onlinetools/core/rclonedriver"
)

func main() {
	rclonedriver.SetRcloneConf("alist", "type", "webdav")
	rclonedriver.SetRcloneConf("alist", "url", "http://8.142.131.147:5244/dav")
	rclonedriver.SetRcloneConf("alist", "vendor", "nextcloud")
	rclonedriver.SetRcloneAuth("alist", "admin", "DtyelQR1")

	f, err := rclonedriver.GetCloudDiskFileSystemByPath("alist:liyun/music")
	if err != nil {
		fmt.Println(err)
	}

	entries, err := f.List(context.Background(), "")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(entries)
	for i, o := range entries {
		fmt.Printf("object[%i] = (%T) %v\n", i, o, o)
	}
}
