package core

func Shutdown() {
	evalBackgroundRewriteAof()
}
