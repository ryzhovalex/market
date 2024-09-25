package assert

func Run(condition bool) {
	if !condition {
		panic("panic:: Assertion error")
	}
}
