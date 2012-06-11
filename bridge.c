#include <wayland-client.h>

#include "bridge.h"
#include "_cgo_export.h"

void
_go_global_handler(struct wl_display *display, uint32_t id,
                   char *interface, uint32_t version, void *data)
{
    goGlobalHandler(id, interface, version, data);
}

struct wl_global_listener *
_wl_display_add_global_listener(struct wl_display *display, void *data)
{
    return wl_display_add_global_listener(display,
            (wl_display_global_func_t)_go_global_handler, data);
}

void
_go_shm_listener(void *data, struct wl_shm *wl_shm, uint32_t format)
{
    goShmListener(data, format);
}

struct wl_shm_listener _shm_listener = {
    _go_shm_listener
};

int
_go_event_mask_update(uint32_t mask, void *data)
{
    return goEventMaskUpdate(mask, data);
}

int
_wl_display_get_fd(struct wl_display *display, void *data)
{
    return wl_display_get_fd(display, _go_event_mask_update, data);
}


void
_go_handle_ping(void *data, struct wl_shell_surface *wl_shell_surface,
    uint32_t serial)
{
    goHandlePing(data, serial);
}

void _go_handle_configure(void *data, struct wl_shell_surface *wl_shell_surface,
    uint32_t edges, int32_t width, int32_t height)
{
    goHandleConfigure(data, edges, width, height);
}

void _go_handle_popup_done(void *data,
    struct wl_shell_surface *wl_shell_surface)
{
    goHandlePopupDone(data);
}

const struct wl_shell_surface_listener _shell_surface_listener = {
    _go_handle_ping,
    _go_handle_configure,
    _go_handle_popup_done,
};

int
_wl_shell_surface_add_listener(struct wl_shell_surface *wl_shell_surface,
    void *data)
{
    wl_shell_surface_add_listener(wl_shell_surface, &_shell_surface_listener,
        data);
}

