package rfdata

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"time"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["LogReader"] = func() flow.Worker { return &LogReader{} }
	flow.Registry["LogReplayer"] = func() flow.Worker { return &LogReplayer{} }
}

// LogReader opens a (possibly compressed) log file and sends out its entries.
// Registers as "LogReader".
type LogReader struct {
	flow.Work
	In   flow.Input
	Name flow.Input
	Out  flow.Output
}

// Start opening a logfile and generating timestamped entries.
func (w *LogReader) Run() {
	// Parse lines of the form "L 02:16:56.749 usb-A40117UK OK 14 25 0 54"
	var re = regexp.MustCompile(`^L (\d\d:\d\d:\d\d\.\d\d\d) (\S+) (\S.*)`)

	if m, ok := <-w.Name; ok {
		name := m.(string)
		var file io.Reader
		file, err := os.Open(name)
		flow.Check(err)
		if path.Ext(name) == ".gz" {
			file, err = gzip.NewReader(file)
			flow.Check(err)
		}
		// Mon Jan 2 15:04:05 -0700 MST 2006
		day, err := time.Parse("20060102", path.Base(name)[:8])
		flow.Check(err)
		yr, mo, dy := day.Date()
		lastDev := ""

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			m := re.FindStringSubmatch(scanner.Text())
			// Mon Jan 2 15:04:05 -0700 MST 2006
			tod, err := time.Parse("15:04:05.000", m[1])
			flow.Check(err)
			w.Out.Send(tod.AddDate(yr, int(mo)-1, dy-1))
			if m[2] != lastDev {
				w.Out.Send("<" + m[2] + ">")
				lastDev = m[2]
			}
			w.Out.Send(m[3])
		}
	}
}

// A LogReplayer generates new entries from existing ones in simulated time.
// Registers as "LogReplayer".
type LogReplayer struct {
	flow.Work
	In  flow.Input
	Out flow.Output
}

// Start reading log entries, and replay them as if they happened today.
func (w *LogReplayer) Run() {
	const MS_PER_DAY = 86400000
	// collect all log entries into two arrays (easier to search that way)
	var times []int
	var texts []string
	var t time.Time
	for m := range w.In {
		switch v := m.(type) {
		case time.Time:
			t = v
		case string:
			times = append(times, int((t.UnixNano()/1000000)%MS_PER_DAY))
			texts = append(texts, v)
		}
	}
	// add a final entry to wrap back to the beginning on each new day
	// this avoids having to special-case searching past the last entry
	times = append(times, times[0]+MS_PER_DAY)
	for {
		// find best candidate using binary search
		ms := int((time.Now().UnixNano() / 1000000) % MS_PER_DAY)
		i := sort.SearchInts(times, ms)
		// sleep until it's time for that next entry
		time.Sleep(time.Duration(times[i]-ms) * time.Millisecond)
		// wrap around at the end of the day
		if i >= len(texts) {
			i = 0
		}
		// replay the string to the output port
		w.Out.Send(texts[i])
		// advance time so that the next search will find a different entry
		time.Sleep(time.Millisecond)
	}
}
