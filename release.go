//go:build release

package webkitgtk

func init() {
	_RELEASE = true
	LogWriter = nil
}
