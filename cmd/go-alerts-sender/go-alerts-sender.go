package goAlertsSender

import (
	"crypto/rand"
	"errors"
	"github.com/hallazzang/go-windows-programming/pkg/win"
	"golang.org/x/sys/windows"
	"log"
	"unsafe"
)

const notifyIconMsg = win.WM_APP + 1

var errShellNotifyIcon = errors.New("Shell_NotifyIcon error")

type notifyIcon struct {
	hwnd uintptr
	guid win.GUID
}

func newNotifyIcon(hwnd uintptr) (*notifyIcon, error) {
	ni := &notifyIcon{
		hwnd: hwnd,
		guid: newGUID(),
	}
	data := ni.newData()
	data.UFlags |= win.NIF_MESSAGE
	data.UCallbackMessage = notifyIconMsg
	if win.Shell_NotifyIcon(win.NIM_ADD, data) == win.FALSE {
		return nil, errShellNotifyIcon
	}
	return ni, nil
}

func (ni *notifyIcon) Dispose() {
	win.Shell_NotifyIcon(win.NIM_DELETE, ni.newData())
}

func (ni *notifyIcon) SetTooltip(tooltip string) error {
	data := ni.newData()
	data.UFlags |= win.NIF_TIP
	copy(data.SzTip[:], windows.StringToUTF16(tooltip))
	if win.Shell_NotifyIcon(win.NIM_MODIFY, data) == win.FALSE {
		return errShellNotifyIcon
	}
	return nil
}

func (ni *notifyIcon) SetIcon(hIcon uintptr) error {
	data := ni.newData()
	data.UFlags |= win.NIF_ICON
	data.HIcon = hIcon
	if win.Shell_NotifyIcon(win.NIM_MODIFY, data) == win.FALSE {
		return errShellNotifyIcon
	}
	return nil
}

func (ni *notifyIcon) ShowNotification(title, text string) error {
	data := ni.newData()
	data.UFlags |= win.NIF_INFO
	copy(data.SzInfoTitle[:], windows.StringToUTF16(title))
	copy(data.SzInfo[:], windows.StringToUTF16(text))
	if win.Shell_NotifyIcon(win.NIM_MODIFY, data) == win.FALSE {
		return errShellNotifyIcon
	}
	return nil
}

func (ni *notifyIcon) ShowNotificationWithIcon(title, text string, hIcon uintptr) error {
	data := ni.newData()
	data.UFlags |= win.NIF_INFO
	copy(data.SzInfoTitle[:], windows.StringToUTF16(title))
	copy(data.SzInfo[:], windows.StringToUTF16(text))
	data.DwInfoFlags = win.NIIF_USER | win.NIIF_LARGE_ICON
	if win.Shell_NotifyIcon(win.NIM_MODIFY, data) == win.FALSE {
		return errShellNotifyIcon
	}
	return nil
}

func (ni *notifyIcon) newData() *win.NOTIFYICONDATA {
	var nid win.NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.UFlags = win.NIF_GUID
	nid.HWnd = ni.hwnd
	nid.GuidItem = ni.guid
	return &nid
}
func wndProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case notifyIconMsg:
		switch nmsg := win.LOWORD(uint32(lParam)); nmsg {
		case win.NIN_BALLOONUSERCLICK:
			log.Print("User has clicked the balloon message")
		case win.WM_LBUTTONDOWN:
			clickHandler()
		}
	case win.WM_DESTROY:
		win.PostQuitMessage(0)
	default:
		return win.DefWindowProc(hWnd, msg, wParam, lParam)
	}
	return 0
}
func newGUID() win.GUID {
	var buf [16]byte
	rand.Read(buf[:])
	return *(*win.GUID)(unsafe.Pointer(&buf[0]))
}
func createMainWindow() (uintptr, error) {
	hInstance := win.GetModuleHandle(nil)

	wndClass := windows.StringToUTF16Ptr("MyWindow")

	var wcex win.WNDCLASSEX
	wcex.CbSize = uint32(unsafe.Sizeof(wcex))
	wcex.Style = win.CS_HREDRAW | win.CS_VREDRAW
	wcex.LpfnWndProc = windows.NewCallback(wndProc)
	wcex.HInstance = hInstance
	wcex.HCursor = win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	wcex.HbrBackground = win.COLOR_WINDOW + 1
	wcex.LpszClassName = wndClass
	if win.RegisterClassEx(&wcex) == 0 {
		return 0, win.GetLastError()
	}

	hwnd := win.CreateWindowEx(0, wndClass, windows.StringToUTF16Ptr("NotifyIcon Example"), win.WS_THICKFRAME, -100, -100, 0, 0, 0, 0, hInstance, nil)
	if hwnd == win.NULL {
		return 0, win.GetLastError()
	}
	win.ShowWindow(hwnd, win.SW_SHOW)

	return hwnd, nil
}
func loadIconFromResource(id uintptr) (uintptr, error) {
	hIcon := win.LoadImage(
		win.GetModuleHandle(nil),
		win.MAKEINTRESOURCE(id),
		win.IMAGE_ICON,
		0, 0,
		win.LR_DEFAULTSIZE)
	if hIcon == win.NULL {
		return 0, win.GetLastError()
	}

	return hIcon, nil
}
func loadIconFromFile(name string) (uintptr, error) {
	hIcon := win.LoadImage(
		win.NULL,
		windows.StringToUTF16Ptr(name),
		win.IMAGE_ICON,
		0, 0,
		win.LR_DEFAULTSIZE|win.LR_LOADFROMFILE)
	if hIcon == win.NULL {
		return 0, win.GetLastError()
	}

	return hIcon, nil
}
func clickHandler() {
	log.Print("User has clicked the notify icon")
}
func sendAlertWithResourceIcon(title string, message string, resourceId uintptr) {
	hIcon, err := loadIconFromResource(resourceId)
	hwnd, err := createMainWindow()
	ni, err := newNotifyIcon(hwnd)

	if err != nil {
		panic(err)
	}
	defer ni.Dispose()

	//ni.SetIcon(hIcon)
	ni.SetTooltip(title)

	ni.ShowNotificationWithIcon("QR", message, hIcon)
}
