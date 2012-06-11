package main

/*
#cgo pkg-config: wayland-client
#include <wayland-client.h>
#include "bridge.h"
*/
import "C"

import (
	"fmt"
)

//export goGlobalHandler
func goGlobalHandler(display *C.struct_wl_display, id C.uint32_t,
	face *C.char, version C.uint32_t, data unsafe.Pointer) {

	println(C.GoString(face))
}

func main() {
	var disp *C.struct_wl_display
	disp = wl_display_connect(nil)
	fmt.Printf("Type: %T\n", disp)

	C._wl_display_add_global_listener(disp, handleGlobal, nil)
}

