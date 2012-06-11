package main

/*
#cgo pkg-config: wayland-client
#include <unistd.h>
#include <sys/mman.h>
#include <wayland-client.h>
#include "os-compatibility.h"
#include "bridge.h"
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

//export goGlobalHandler
func goGlobalHandler(id C.uint32_t, face *C.char,
	version C.uint32_t, data unsafe.Pointer) {

	disp := (*display)(data)
	goFace := C.GoString(face)
	switch goFace {
	case "wl_compositor":
		disp.compositor = (*C.struct_wl_compositor)(
			C.wl_display_bind(disp.display, id,
				(*C.struct_wl_interface)(&C.wl_compositor_interface)))
	case "wl_shell":
		disp.shell = (*C.struct_wl_shell)(
			C.wl_display_bind(disp.display, id,
				(*C.struct_wl_interface)(&C.wl_shell_interface)))
	case "wl_shm":
		disp.shm = (*C.struct_wl_shm)(
			C.wl_display_bind(disp.display, id,
				(*C.struct_wl_interface)(&C.wl_shm_interface)))
		C.wl_shm_add_listener(disp.shm, &C._shm_listener,
			unsafe.Pointer(disp))
	}
}

//export goShmListener
func goShmListener(data unsafe.Pointer, format C.uint32_t) {
	disp := (*display)(data)
	disp.formats |= (1 << uint32(format))
}

//export goEventMaskUpdate
func goEventMaskUpdate(mask C.uint32_t, data unsafe.Pointer) C.int {
	disp := (*display)(data)
	disp.mask = uint32(mask)
	return 0
}

//export goHandlePing
func goHandlePing(data unsafe.Pointer, serial C.uint32_t) {
	win := (*window)(data)
	C.wl_shell_surface_pong(win.shellSurface, serial)
}

//export goHandleConfigure
func goHandleConfigure(data unsafe.Pointer, edges C.uint32_t,
	width, height C.int32_t) {

	println("confgure", width, height)
}

//export goHandlePopupDone
func goHandlePopupDone(data unsafe.Pointer) {
	println("popup done")
}

type display struct {
	display *C.struct_wl_display
	compositor *C.struct_wl_compositor
	shell *C.struct_wl_shell
	shm *C.struct_wl_shm
	formats uint32
	mask uint32
}

func newDisplay() *display {
	disp := &display{}
	disp.display = C.wl_display_connect(nil)
	fmt.Printf("Type of wl_display value: %T\n", disp.display)

	if disp.display == nil {
		panic("Could not connect to Wayland.")
	}

	C._wl_display_add_global_listener(disp.display, unsafe.Pointer(disp))
	C.wl_display_iterate(disp.display, C.WL_DISPLAY_READABLE)
	C.wl_display_roundtrip(disp.display)

	if disp.formats & (1 << C.WL_SHM_FORMAT_XRGB8888) == 0 {
		panic("WL_SHM_FORMAT_XRGB32 not available")
	}

	C._wl_display_get_fd(disp.display, unsafe.Pointer(disp))

	return disp
}

type window struct {
	display *display
	width, height int
	surface *C.struct_wl_surface
	shellSurface *C.struct_wl_shell_surface
	buffer *C.struct_wl_buffer
	shmData unsafe.Pointer
	callback *C.struct_wl_callback
}

func newWindow(disp *display, width, height int) *window {
	win := &window{
		display: disp,
		width: width,
		height: height,
		callback: nil,
	}
	win.surface = C.wl_compositor_create_surface(disp.compositor)
	win.shellSurface = C.wl_shell_get_shell_surface(disp.shell, win.surface)
	win.buffer = win.newBuffer(width, height, uint32(C.WL_SHM_FORMAT_XRGB8888))

	if win.buffer == nil {
		panic("Could not create window buffer")
	}

	if win.shellSurface != nil {
		C._wl_shell_surface_add_listener(win.shellSurface,
			unsafe.Pointer(win))
	}
	C.wl_shell_surface_set_toplevel(win.shellSurface)

	return win
}

func (w *window) newBuffer(width, height int,
	format uint32) *C.struct_wl_buffer {

	stride := width * 4
	size := stride * height

	fd := C.os_create_anonymous_file(C.off_t(size))
	if fd < 0 {
		panic("Could not create buffer file.")
	}

	data := C.mmap(nil, C.size_t(size), C.PROT_READ | C.PROT_WRITE,
		C.MAP_SHARED, C.int(fd), 0)
	if *(*int)(data) == -1 {
		panic("mmap failed")
		C.close(fd)
		return nil
	}

	pool := C.wl_shm_create_pool(w.display.shm,
		C.int32_t(fd), C.int32_t(size))
	buffer := C.wl_shm_pool_create_buffer(pool, 0,
		C.int32_t(width), C.int32_t(height),
		C.int32_t(stride), C.uint32_t(format))
	C.wl_shm_pool_destroy(pool)
	C.close(fd)

	w.shmData = data
	return buffer
}

func (w *window) paintPixels(time uint32) {
	var or int
	halfw, halfh := w.width / 2, w.height / 2
	if halfw < halfh {
		or = halfw - 8
	} else {
		or = halfh - 8
	}
	ir := or - 32
	or *= or
	ir *= ir
	data := w.shmData

	for y := 0; y < w.height; y++ {
		y2 := (y - halfh) * (y - halfh)
		for x := 0; x < w.width; x++ {
			var v int
			r2 := (x - halfw) * (x - halfw) + y2
			if r2 < ir {
				v = (r2 / 32 + int(time) / 64) * 0x0080401
			} else if r2 < or {
				v = (y + int(time) / 32) * 0x0080401
			} else {
				v = (x + int(time) / 16) * 0x0080401
			}
			v &= 0x00ffffff

			*(*int)(data) = v
			data = unsafe.Pointer(uintptr(data) + 4)
		}
	}
}

func redraw(data unsafe.Pointer, callback *C.struct_wl_callback,
	time C.uint32_t) {

	win := (*window)(data)
	win.paintPixels(uint32(time))
	C.wl_surface_attach(win.surface, win.buffer, 0, 0)
	C.wl_surface_damage(win.surface, 0, 0,
		C.int32_t(win.width), C.int32_t(win.height))
}

func main() {
	disp := newDisplay()
	win := newWindow(disp, 250, 250)

	redraw(unsafe.Pointer(win), nil, 0)

	C.wl_display_iterate(disp.display, C.uint32_t(disp.mask))

	time.Sleep(5 * time.Second)
}

