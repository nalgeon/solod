#include "main.h"

// -- Implementation --

int main(void) {
    {
        // Current time.
        time_TimeT now = 0;
        now = time_Time(&now);
        if (now <= 0) {
            so_panic("want now > 0");
        }
    }
    {
        // Clock.
        so_int ticks = time_Clock();
        (void)ticks;
    }
    {
        // ClocksPerSec.
        so_int cps = time_ClocksPerSec;
        if (cps <= 0) {
            so_panic("want ClocksPerSec > 0");
        }
    }
    {
        // Difftime.
        time_TimeT t1 = 0;
        t1 = time_Time(&t1);
        time_TimeT t2 = 0;
        t2 = time_Time(&t2);
        double diff = time_Difftime(t2, t1);
        if (diff < 0.0) {
            so_panic("want diff >= 0");
        }
    }
    {
        // Gmtime.
        time_TimeT ts = 0;
        ts = 0;
        time_Tm tm = time_Gmtime(&ts);
        // Unix epoch: 1970-01-01 00:00:00 UTC.
        if (tm.Year != 70) {
            so_panic("want Year == 70");
        }
        if (tm.Mon != 0) {
            so_panic("want Mon == 0");
        }
        if (tm.Mday != 1) {
            so_panic("want Mday == 1");
        }
    }
    {
        // Mktime.
        time_Tm tm = (time_Tm){.Sec = 0, .Min = 0, .Hour = 0, .Mday = 1, .Mon = 0, .Year = 70, .Isdst = -1};
        time_TimeT ts = time_Mktime(&tm);
        // Should normalize and return a valid timestamp.
        (void)ts;
    }
    {
        // Strftime.
        so_byte buf[64] = {0};
        time_TimeT ts = 0;
        ts = 0;
        time_Tm tm = time_Gmtime(&ts);
        so_int n = time_Strftime(&buf[0], 64, "%Y-%m-%d", &tm);
        if (n == 0) {
            so_panic("strftime failed");
        }
        stdio_Printf("%s\n", &buf[0]);
    }
}
