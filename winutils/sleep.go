package winutils

import (
	"syscall"
)

// Windows API constants for SetThreadExecutionState
const (
	ES_CONTINUOUS       = 0x80000000
	ES_DISPLAY_REQUIRED = 0x00000002
	ES_SYSTEM_REQUIRED  = 0x00000001
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")

// preventSleep tells Windows not to sleep or turn off display while app is running
func PreventSleep() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS | ES_DISPLAY_REQUIRED | ES_SYSTEM_REQUIRED))
}

// allowSleep restores normal Windows sleep behavior
func AllowSleep() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
}
