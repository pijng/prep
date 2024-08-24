package prep

func Comptime[FN any](fn FN) FN {
	return fn
}
