#include "so/builtin/builtin.h"

#define newObj(T) (alloca(sizeof(T)))
#define freeObj(T, ptr) ((void)(ptr))
#define newMap(K, V, size) ((main_Map){.len = (size)})
#define main_Map_Len(K, V, m) ((m)->len)

typedef struct {
    int len;
} main_Map;
