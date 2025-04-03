package builtin

import (
	"lowcode/core/builtin/apps"
	"lowcode/core/control"
)

func RegisterApps() {
	//static resource page
	control.AddBuiltinApplication(apps.BuildStaticResource)

	//home page
	control.AddBuiltinApplication(apps.BuildHomePage)

}
