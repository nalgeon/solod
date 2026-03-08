#include <string.h>

#define c_Bytes(ptr, n) ({   \
    void* _p = (void*)(ptr); \
    _p ? unsafe_Slice(_p, n) : (so_Slice){NULL, 0, 0}; \
})

#define c_String(ptr) ({            \
    const char* _p = (const char*)(ptr); \
    _p ? unsafe_String(_p, strlen(_p)) : (so_String){NULL, 0}; \
})

#define c_CharPtr(ptr) ((char*)(ptr))
