package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type LogEntry struct {
    time_received time.Time
    duration      time.Duration
    backend       string
}

// 28/Mar/2016:06:25:21 +0100

const time_layout = "[2/Jan/2006:15:04:05 -0700]"

var wanted_backend_rx = regexp.MustCompile("^live_wanda_\\d+_cantor\\s*$")

var window = time.Duration(5) * time.Minute

func same_window (t time.Time, y *LogEntry) bool {
    return y.time_received.Round(window) == t
}

func parse_line(line string) (*LogEntry, error) {
    components := strings.Split(line, "|")
    if len(components) < 12 {
        return nil, fmt.Errorf("Log line has too few components (%d): %s", len(components), line)
    }
    time_ix := 3
    duration_ix := len(components) - 2
    backend_ix := len(components) - 1
    request_time, err := time.Parse(time_layout, components[time_ix])
    if err != nil {
        return nil, fmt.Errorf("Failed to parse request time %s: %v", components[time_ix], err)
    }
    duration, err := strconv.Atoi(components[duration_ix])
    if err != nil {
        return nil, fmt.Errorf("Failed to parse duration %s: %v", components[duration_ix], err)
    }
    entry := LogEntry{request_time, time.Duration(duration)*time.Microsecond, components[backend_ix]}
    return &entry, nil
}

func is_wanted(entry *LogEntry) bool {
    return wanted_backend_rx.MatchString(entry.backend)
}

func next_entry (in *bufio.Reader) *LogEntry {
    line, err := in.ReadString('\n')
    if err == io.EOF {
        return nil
    }
    if err != nil {
        log.Println(err)
        return next_entry(in)
    }
    entry, err := parse_line(line)
    if err != nil {
        log.Println(err)
        return next_entry(in)
    }
    if is_wanted(entry) {
        return entry
    }
    return next_entry(in)
}

func process_file(filename string) {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    reader := bufio.NewReader(file)
    x := next_entry(reader)
    for x != nil {
        t := x.time_received.Round(window)
        num_requests := 0
        total_duration := time.Duration(0)
        for x != nil && same_window(t, x) {
            num_requests++
            total_duration += x.duration
            x = next_entry(reader)
        }
        fmt.Printf("%v % 6d %8.3f\n", t, num_requests, total_duration.Seconds()/float64(num_requests))
    }
}

func main() {
    for _, filename := range os.Args[1:] {
        process_file(filename)
    }
}
