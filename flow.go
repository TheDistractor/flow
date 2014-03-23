package flow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/golang/glog"
)

// Version of this package.
var Version = "0.9.0"

// The registry is the factory for all known types of gadgets.
var Registry = map[string]func() Circuitry{}

// Config stores configuration settings for general use.
var Config = map[string]string{}

// Messages are the generic type sent to, between, and from gadgets.
type Message interface{}

// A tag allows adding a descriptive string to a message.
type Tag struct {
	Tag string
	Msg Message
}

// Input pins are used to receive messages.
type Input <-chan Message

// Output pins are used to send messages elsewhere.
type Output interface {
	Send(v Message) // Send a message through an output pin.
	Disconnect()    // Disconnect the pin, close channel when last one is gone.
}

// Circuitry is the collective name for circuits and gadgets.
type Circuitry interface {
	Run()

	initGadget(Circuitry, string, *Circuit) *Gadget
}

// A transformer processes each message through a supplied function.
func Transformer(f func(Message) Message) Circuitry {
	return &transformer{fun: f}
}

type transformer struct {
	Gadget
	In  Input
	Out Output

	fun func(Message) Message
}

func (g *transformer) Run() {
	for m := range g.In {
		// if m, ok := <-g.In; ok {
		g.Out.Send(g.fun(m))
	}
}

// A wire is a ref-counted Input, it's closed when the count drops to 0.
type wire struct {
	channel  chan Message
	senders  int
	capacity int
	dest     *Gadget
}

func (c *wire) Send(v Message) {
	c.dest.sendTo(c, v)
}

func (c *wire) Disconnect() {
	c.senders--
	if c.senders == 0 && c.channel != nil {
		close(c.channel)
	}
}

// Use a fake sink for every output pin not connected to anything else.
type fakeSink struct{}

func (c *fakeSink) Send(m Message) {
	fmt.Printf("Lost %T: %v\n", m, m)
}

func (c *fakeSink) Disconnect() {}

// extract "a" from "a.b", panics if there's no dot in the string
func gadgetPart(s string) string {
	n := strings.IndexRune(s, '.')
	return s[:n]
}

// extract "b" from "a.b", also works if only "b" is given
func pinPart(s string) string {
	n := strings.IndexRune(s, '.')
	return s[n+1:]
}

// Utility to check for errors, report as fatal error if the arg is not nil.
func Check(err interface{}) {
	if err != nil {
		glog.Fatal(err)
	}
}

// Call this as "defer flow.DontPanic()" for a concise stack trace on panics.
func DontPanic() {
	// generate a nice stack trace, see https://code.google.com/p/gonicetrace/
	if e := recover(); e != nil {
		fmt.Fprintf(os.Stderr, "\nPANIC: %v\n", e)
		for skip := 1; skip < 20; skip++ {
			pc, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			if strings.HasSuffix(file, ".go") {
				name := runtime.FuncForPC(pc).Name()
				name = name[strings.LastIndex(name, "/")+1:]
				fmt.Fprintf(os.Stderr, "%s:%d %s()\n", file, line, name)
			}
		}
		glog.Error("EXIT")
	}
}

// AddToRegistry adds circuit definitions from a JSON file to the registry.
func AddToRegistry(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var definitions map[string]json.RawMessage
	err = json.Unmarshal(data, &definitions)
	if err != nil {
		return err
	}
	for name, def := range definitions {
		registerCircuit(name, def)
	}
	return nil
}

func registerCircuit(name string, def []byte) {
	Registry[name] = func() Circuitry {
		g := NewCircuit()
		err := g.LoadJSON(def)
		Check(err)
		return g
	}
}

// Print a compact list of the registry entries on standard output.
func PrintRegistry() {
	keys := []string{}
	for k := range Registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := " "
	for _, k := range keys {
		if len(s)+len(k) > 78 {
			println(s)
			s = " "
		}
		s += " " + k
	}
	if len(s) > 1 {
		println(s)
	}
}

// LoadConfig parses a configuration file, if it exists, to set up some basic
// application settings, such as where the app/ and data/ directories are.
// Settings can be overridden through environment variables with the same name.
func LoadConfig(defaults, filename string) map[string]string {
	parseSettingsLine := func(line string) {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			fields := strings.SplitN(line, "=", 2)
			if len(fields) != 2 {
				panic(fmt.Errorf("cannot parse configuration: %s", line))
			}
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])
			env := os.Getenv(key)
			if env != "" {
				value = env
			}
			glog.Infoln("config", key, "=", value)
			Config[key] = value
		}
	}

	for _, s := range strings.Split(defaults, "\n") {
		parseSettingsLine(s)
	}

	if fd, err := os.Open(filename); err == nil {
		defer fd.Close()
		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			parseSettingsLine(scanner.Text())
		}
	}

	return Config
}
