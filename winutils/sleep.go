package winutils

import (
	"syscall"
	"unsafe"
)

// Windows API constants for SetThreadExecutionState
const (
	ES_CONTINUOUS       = 0x80000000
	ES_DISPLAY_REQUIRED = 0x00000002
	ES_SYSTEM_REQUIRED  = 0x00000001
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
var user32 = syscall.NewLazyDLL("user32.dll")
var getForegroundWindow = user32.NewProc("GetForegroundWindow")
var getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
var getCurrentProcessId = kernel32.NewProc("GetCurrentProcessId")

// preventSleep tells Windows not to sleep or turn off display while the app is focused.
func PreventSleep() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS | ES_DISPLAY_REQUIRED | ES_SYSTEM_REQUIRED))
}

// allowSleep restores normal Windows sleep behavior.
func AllowSleep() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
}

func UpdateSleepPreventionState(currentState bool, focused bool, prevent func(), allow func()) bool {
	if focused == currentState {
		return currentState
	}

	if focused {
		prevent()
		return true
	}

	allow()
	return false
}

func IsAppInForeground() bool {
	foregroundWindow, _, _ := getForegroundWindow.Call()
	if foregroundWindow == 0 {
		return false
	}

	currentProcessID, _, _ := getCurrentProcessId.Call()
	var windowProcessID uint32
	getWindowThreadProcessId.Call(foregroundWindow, uintptr(unsafe.Pointer(&windowProcessID)))

	return uintptr(windowProcessID) == currentProcessID
}
