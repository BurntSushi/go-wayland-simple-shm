struct wl_shm_listener _shm_listener;
const struct wl_shell_surface_listener _shell_surface_listener;

struct wl_global_listener *
_wl_display_add_global_listener(struct wl_display *display, void *data);

int
_wl_display_get_fd(struct wl_display *display, void *data);

int
_wl_shell_surface_add_listener(struct wl_shell_surface *wl_shell_surface,
    void *data);

