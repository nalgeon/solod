#define newObj(T) (alloca(sizeof(T)))
#define freeObj(T, ptr) ((void)(ptr))
