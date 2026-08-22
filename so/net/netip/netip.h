#include "so/builtin/builtin.h"

#if defined(so_build_hosted) && defined(so_build_windows)

// if_nametoindex requires iphlpapi. Link the program with -liphlpapi,
// because a so:link directive cannot depend on the target.
#include <winsock2.h>

#include <iphlpapi.h>

#elif defined(so_build_hosted) && !defined(so_build_wasm)

#include <net/if.h>

#else

// A freestanding or WASM environment has no network interfaces, so an interface
// name can't be resolved to an index. Returning 0 matches the behavior in hosted
// environments when a name doesn't match any interface.
static inline unsigned int if_nametoindex(const char* ifname) {
    (void)ifname;
    return 0;
}

#endif  // so_build_hosted
