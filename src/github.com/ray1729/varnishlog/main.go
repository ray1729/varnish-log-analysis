package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "regexp"
    "sort"
    "strconv"
    "time"
)

var window_duration = time.Duration(5) * time.Minute

var log_line_rx = regexp.MustCompile("^[^|]+\\|-\\|[^|]+\\|\\[([^]]+)\\]\\|.+\\|(\\d+)\\|([^|]+)\\s*$")

var wanted_backend_rx = regexp.MustCompile("^live_wanda_\\d+_cantor$")

const time_layout = "2/Jan/2006:15:04:05 -0700"

type LogEntry struct {
    time time.Time
    duration int
    backend string
}

type AccumulatorEntry struct {
    num_requests int
    total_duration int
    max_duration int
    min_duration int
}

func parse_line(line string) (*LogEntry, error) {
    components := log_line_rx.FindStringSubmatch(line)
    if len(components) != 4 {
        return nil, fmt.Errorf("Failed to parse line %s", line)
    }
    if !wanted_backend_rx.MatchString(components[3]) {
        return nil, nil
    }
    time, err := time.Parse(time_layout, components[1])
    if err != nil {
        return nil, fmt.Errorf("Failed to parse request time %s", components[1])
    }
    duration, err := strconv.Atoi(components[2])
    if err != nil {
        return nil, fmt.Errorf("Failed to parse request duration %s", components[2])
    }
    entry := LogEntry{time, duration, components[3]}
    return &entry, nil
}

func min(x,y int) int {
    if x < y {
        return x
    }
    return y
}

func max(x,y int) int {
    if x > y {
        return x
    }
    return y
}

func accumulate(accumulator map[time.Time]AccumulatorEntry, entry *LogEntry) {
    k := entry.time.Round(window_duration)
    v, ok := accumulator[k]
    if ok {
        v.num_requests++
        v.total_duration += entry.duration
        v.max_duration = max(v.max_duration, entry.duration)
        v.min_duration = min(v.min_duration, entry.duration)
    } else {
        v.num_requests = 1
        v.total_duration = entry.duration
        v.max_duration = entry.duration
        v.min_duration = entry.duration
    }
    accumulator[k] = v
}

type ByTime []time.Time

func (a ByTime) Len() int {
    return len(a)
}

func (a ByTime) Swap(i,j int) {
    a[i], a[j] = a[j], a[i]
}

func (a ByTime) Less(i,j int) bool {
    return a[i].Before(a[j])
}

func print_summary (accumulator map[time.Time]AccumulatorEntry) {
    var keys []time.Time
    for k := range accumulator {
        keys = append(keys, k)
    }
    sort.Sort(ByTime(keys))
    for _, k := range keys {
        entry := accumulator[k]
        fmt.Printf("%v % 8d % 10d % 10d % 10d\n", k,
            entry.num_requests,
            entry.min_duration,
            entry.max_duration,
            entry.total_duration/entry.num_requests)
    }
}

func process_file(accumulator map[time.Time]AccumulatorEntry, filename string) {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        entry, err := parse_line(scanner.Text())
        if err != nil {
            log.Printf("Reading %s: %v", filename, err)
            continue
        }
        if entry != nil {
            accumulate(accumulator, entry)
        }
    }
    if err := scanner.Err(); err != nil {
        log.Printf("Reading %s: %v", filename, err)
    }
}

func main() {
    accumulator := make(map[time.Time]AccumulatorEntry)
    for _, filename := range os.Args[1:] {
        process_file(accumulator, filename)
    }
    print_summary(accumulator)
}
