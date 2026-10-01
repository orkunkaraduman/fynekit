package fynekit

/*
#cgo CFLAGS: -xobjective-c -fvisibility=hidden
*/
import "C"

var (
	R  = Walk
	W  = Wrap
	WM = WrapWithMinSize
)
