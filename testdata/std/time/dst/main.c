#include "main.h"

// -- Implementation --

int main(void) {
    {
        // time.Date and time.Time properties.
        time_Time t = time_Date(2021, time_May, 10, 12, 33, 44, 777888999, time_UTC);
        if (time_Time_Year(t) != 2021) {
            so_panic("unexpected Time.Year");
        }
        if (time_Time_Month(t) != time_May) {
            so_panic("unexpected Time.Month");
        }
        if (time_Time_Day(t) != 10) {
            so_panic("unexpected Time.Day");
        }
        if (time_Time_Hour(t) != 12) {
            so_panic("unexpected Time.Hour");
        }
        if (time_Time_Minute(t) != 33) {
            so_panic("unexpected Time.Minute");
        }
        if (time_Time_Second(t) != 44) {
            so_panic("unexpected Time.Second");
        }
        if (time_Time_Nanosecond(t) != 777888999) {
            so_panic("unexpected Time.Nanosecond");
        }
        so_println("%" PRId64 " %" PRId64 " %" PRId64 " %" PRId64 " %" PRId64 " %" PRId64 " %" PRId64, time_Time_Year(t), time_Time_Month(t), time_Time_Day(t), time_Time_Hour(t), time_Time_Minute(t), time_Time_Second(t), time_Time_Nanosecond(t));
    }
    {
        // time.Time.Format and time.Time.String.
        time_Time t = time_Date(2024, time_March, 15, 14, 30, 45, 0, time_UTC);
        so_byte buf[64] = {0};
        so_String s = time_Time_Format(t, so_str("%Y-%m-%d"), time_UTC, so_array_slice(so_byte, buf, 0, 64, 64));
        if (so_string_ne(s, so_str("2024-03-15"))) {
            so_panic("unexpected Format");
        }
        s = time_Time_String(t, so_array_slice(so_byte, buf, 0, 64, 64));
        if (so_string_ne(s, so_str("2024-03-15T14:30:45Z"))) {
            so_panic("unexpected String");
        }
    }
    {
        // time.Parse.
        time_TimeResult _res1 = time_Parse(so_str("%Y-%m-%d %H:%M:%S"), so_str("2024-03-15 14:30:45"), time_UTC);
        time_Time t = _res1.val;
        so_Error err = _res1.err;
        if (err != NULL) {
            so_panic("unexpected Parse error");
        }
        time_CalDate date = time_Time_Date(t, time_UTC);
        time_CalClock clock = time_Time_Clock(t, time_UTC);
        if (date.Year != 2024 || date.Month != time_March || date.Day != 15) {
            so_panic("unexpected Parse date");
        }
        if (clock.Hour != 14 || clock.Minute != 30 || clock.Second != 45) {
            so_panic("unexpected Parse clock");
        }
    }
    {
        // time.Parse error.
        time_TimeResult _res2 = time_Parse(so_str("%Y-%m-%d"), so_str("not-a-date"), time_UTC);
        so_Error err = _res2.err;
        if (err == NULL) {
            so_panic("expected Parse error");
        }
    }
    {
        // Time.Now.
        time_Time t = time_Now();
        if (time_Time_IsZero(t)) {
            so_panic("unexpected Time.IsZero");
        }
        so_println("%s %.*s", "UTC:", time_Time_String(t, so_make_slice(so_byte, 64, 64)).len, time_Time_String(t, so_make_slice(so_byte, 64, 64)).ptr);
        time_Offset utc5 = (time_Offset)(5 * 3600);
        so_String layout = so_str("%Y-%m-%dT%H:%M:%S%z");
        so_println("%s %.*s", "UTC+5:", time_Time_Format(t, layout, utc5, so_make_slice(so_byte, 64, 64)).len, time_Time_Format(t, layout, utc5, so_make_slice(so_byte, 64, 64)).ptr);
    }
}
