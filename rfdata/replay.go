package rfdata

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/jcw/flow/flow"
)

func init() {
	flow.Registry["LogReader"] = func() flow.Worker { return &LogReader{} }
}

// LogReader opens a (possibly compressed) log file and sends out its entries.
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
		if err != nil {
			panic(err)
		}
		if path.Ext(name) == ".gz" {
			file, err = gzip.NewReader(file)
			if err != nil {
				panic(err)
			}
		}
		// Mon Jan 2 15:04:05 -0700 MST 2006
		day, err := time.Parse("20060102", path.Base(name)[:8])
		if err != nil {
			panic(err)
		}
		yr, mo, dy := day.Date()
		lastDev := ""
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			m := re.FindStringSubmatch(scanner.Text())
			// Mon Jan 2 15:04:05 -0700 MST 2006
			tod, err := time.Parse("15:04:05.000", m[1])
			if err != nil {
				panic(err)
			}
			w.Out.Send(tod.AddDate(yr, int(mo)-1, dy-1))
			if m[2] != lastDev {
				w.Out.Send("<" + m[2] + ">")
				lastDev = m[2]
			}
			w.Out.Send(m[3])
		}
	}
}
