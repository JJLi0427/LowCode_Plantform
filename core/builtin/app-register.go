package builtin

import (
	"onlinetools/core/builtin/apps"
	"onlinetools/core/control"
)

func RegisterApps() {
	//static resource page
	control.AddBuiltinApplication(apps.BuildStaticResource)

	//home page
	control.AddBuiltinApplication(apps.BuildHomePage)

}
