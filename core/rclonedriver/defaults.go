package rclonedriver

func ConnectDefaultRclone() string {
	SetRcloneConf("alist", "type", "webdav")
	SetRcloneConf("alist", "url", "http://8.142.131.147:5244/dav")
	SetRcloneConf("alist", "vendor", "nextcloud")
	//rclonedriver.SetRcloneConf("alist", "headers", "Referer:")
	SetRcloneAuth("alist", "admin", "DtyelQR1")

	return "http://8.142.131.147:5244/d/liyunweb"
}
