package control

var builtinBuilder []func(*Control) error

func AddBuiltinApplication(builder func(*Control) error) bool {
	if builder != nil {
		builtinBuilder = append(builtinBuilder, builder)
	}

	return true
}

func GetBuiltinApplications() []func(*Control) error {
	return builtinBuilder
}
