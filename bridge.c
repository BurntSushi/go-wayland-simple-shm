#include <wayland-client.h>

#include "bridge.h"
#include "_cgo_export.h"

struct wl_global_listener *
_wl_display_add_global_listener(struct wl_display *display, void *data) {
	return wl_display_add_global_listener(display,
            (wl_display_global_func_t)goGlobalHandler, data);
}

