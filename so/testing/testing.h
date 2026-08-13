// msgSize is the size of the buffer that Errorf and Fatalf format into.
// A longer message is truncated.
#define testing_msgSize 1024

// Errorf formats its arguments like fmt.Sprintf,
// logs the result, and marks the test failed.
void testing_T_Errorf(void* self, so_String format, ...);

// Fatalf is like Errorf but marks the test as fatally failed.
// The test function must return right after calling it.
void testing_T_Fatalf(void* self, so_String format, ...);
