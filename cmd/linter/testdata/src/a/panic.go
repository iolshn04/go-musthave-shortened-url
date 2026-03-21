package a

func panicFunc() {
	panic("boom") // want "panic call is forbidden"
}
