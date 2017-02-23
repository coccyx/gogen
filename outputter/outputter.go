package outputter

import (
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"time"
	"sync/atomic"
	"bytes"

	"github.com/coccyx/go-s2s/s2s"
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/template"
)

var (
	eventsWritten int64
	bytesWritten  int64
	lastTS        time.Time
	rotchan       chan *config.OutputStats
	gout          [config.MaxOutputThreads]config.Outputter
)

// ROT starts the Read Out Thread which will log statistics about what's being output
// ROT is intended to be started as a goroutine which will log output every c.
func ROT(c *config.Config) {
	rotchan = make(chan *config.OutputStats)
	go readStats()

	lastEventsWritten := eventsWritten
	lastBytesWritten := bytesWritten
	var gbday, eventssec, kbytessec float64
	var tempEW, tempBW int64
	lastTS = time.Now()
	for {
		timer := time.NewTimer(time.Duration(c.Global.ROTInterval) * time.Second)
		<-timer.C
		n := time.Now()
		tempEW = eventsWritten
		tempBW = bytesWritten
		eventssec = float64(tempEW-lastEventsWritten) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval)
		kbytessec = float64(tempBW-lastBytesWritten) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval) / 1024
		gbday = (kbytessec * 60 * 60 * 24) / 1024 / 1024
		log.Infof("Events/Sec: %.2f Kilobytes/Sec: %.2f GB/Day: %.2f", eventssec, kbytessec, gbday)
		lastTS = n
		lastEventsWritten = tempEW
		lastBytesWritten = tempBW
	}
}

func readStats() {
	for {
		select {
		case os := <-rotchan:
			eventsWritten += os.EventsWritten
			bytesWritten += os.BytesWritten
		}
	}
}

// Account sends eventsWritten and bytesWritten to the readStats() thread
func Account(eventsWritten int64, bytesWritten int64) {
	os := new(config.OutputStats)
	os.EventsWritten = eventsWritten
	os.BytesWritten = bytesWritten
	rotchan <- os
}

// A hacky way to prepend every* tag to raw event.
func prependSearchTag(event string, numEvents uint64) string {
	var buffer bytes.Buffer
	if math.Mod(float64(numEvents), 10) == 0 {
		buffer.WriteString("every10")
	}
	if math.Mod(float64(numEvents), 100) == 0 {
		buffer.WriteString(" every100")
	}
	if math.Mod(float64(numEvents), 1000) == 0 {
		buffer.WriteString(" every1K")
	}
	if math.Mod(float64(numEvents), 10000) == 0 {
		buffer.WriteString(" every10K")
	}
	if math.Mod(float64(numEvents), 100000) == 0 {
		buffer.WriteString(" every100K")
	}
	if math.Mod(float64(numEvents), 1000000) == 0 {
		buffer.WriteString(" every1M")
	}
	if math.Mod(float64(numEvents), 10000000) == 0 {
		buffer.WriteString(" every10M")
	}
	buffer.WriteString(" ")
	buffer.WriteString(event)
	return buffer.String()
}

// Start starts an output thread and runs until notified to shut down
func Start(oq chan *config.OutQueueItem, oqs chan int, num int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	var lastS *config.Sample
	var out config.Outputter
	var duration float64
	numBytes := make(chan int64, 1)
	var numEvents uint64 = 0

	for {
		item, ok := <-oq
		if !ok {
			if lastS != nil {
				log.Infof("Closing output for sample '%s'", lastS.Name)
				out.Close()
				gout[num] = nil
			}
			oqs <- 1
			break
		}
		out = setup(generator, item, num)
		startTime := time.Now()

		if len(item.Events) > 0 {
			go func() {
				var bytes int64
				defer item.IO.W.Close()
				switch item.S.Output.OutputTemplate {
				case "raw", "json", "splunktcp":
					for _, line := range item.Events {
						var tempbytes int
						var err error
						if item.S.Output.Outputter != "devnull" {
							switch item.S.Output.OutputTemplate {
							case "raw":
								atomic.AddUint64(&numEvents, 1)
								l := prependSearchTag(line["_raw"], numEvents)
								tempbytes, err = io.WriteString(item.IO.W, l)
								if err != nil {
									log.Errorf("Error writing to IO Buffer: %s", err)
								}
							case "json":
								jb, err := json.Marshal(line)
								if err != nil {
									log.Errorf("Error marshaling json: %s", err)
								}
								tempbytes, err = item.IO.W.Write(jb)
								if err != nil {
									log.Errorf("Error writing to IO Buffer: %s", err)
								}
							case "splunktcp":
								tempbytes, err = item.IO.W.Write(s2s.EncodeEvent(line).Bytes())
								if err != nil {
									log.Errorf("Error writing to IO Buffer: %s", err)
								}
							}
						} else {
							tempbytes = len(line["_raw"])
						}
						bytes += int64(tempbytes) + 1
						if item.S.Output.Outputter != "devnull" {
							_, err = io.WriteString(item.IO.W, "\n")
							if err != nil {
								log.Errorf("Error writing to IO Buffer: %s", err)
							}
						}
					}
				default:
					if !template.Exists(item.S.Output.OutputTemplate + "_row") {
						log.Errorf("Template %s does not exist, skipping output", item.S.Output.OutputTemplate)
						return
					}
					// We'll crash on empty events, but don't do that!
					bytes += int64(getLine("header", item.S, item.Events[0], item.IO.W))
					// log.Debugf("Out Queue Item %#v", item)
					var last int
					for i, line := range item.Events {
						atomic.AddUint64(&numEvents, 1)
						line["_raw"] = prependSearchTag(line["_raw"], numEvents)
						bytes += int64(getLine("row", item.S, line, item.IO.W))
						last = i
					}
					bytes += int64(getLine("footer", item.S, item.Events[last], item.IO.W))
				}
				if item.S.KBps != 0 {
					numBytes <- bytes
				}
				Account(int64(len(item.Events)), bytes)

			}()

			err := out.Send(item)
			if err != nil {
				log.Errorf("Error with Send(): %s", err)
			}

			if item.S.KBps != 0 {
				expectedDuration := float64(<-numBytes) / 1024.0 / float64(item.S.KBps)
				duration = time.Since(startTime).Seconds()
				if expectedDuration > duration {
					log.Debugf("expectedDuration: %.2f, duration: %.2f, sleep: %.2f",
						expectedDuration, duration, expectedDuration - duration)
					time.Sleep(time.Duration((expectedDuration - duration) * 1000) * time.Millisecond)
				}
			}

		}
		lastS = item.S
	}
}

func getLine(templatename string, s *config.Sample, line map[string]string, w io.Writer) (bytes int) {
	if template.Exists(s.Output.OutputTemplate + "_" + templatename) {
		linestr, err := template.Exec(s.Output.OutputTemplate+"_"+templatename, line)
		if err != nil {
			log.Errorf("Error from sample '%s' in template execution: %v", s.Name, err)
		}
		// log.Debugf("Outputting line %s", linestr)
		bytes, err = w.Write([]byte(linestr))
		_, err = w.Write([]byte("\n"))
		if err != nil {
			log.Errorf("Error sending event for sample '%s' to outputter '%s': %s", s.Name, s.Output.Outputter, err)
		}
	}
	return bytes
}

func setup(generator *rand.Rand, item *config.OutQueueItem, num int) config.Outputter {
	item.Rand = generator
	item.IO = config.NewOutputIO()

	if gout[num] == nil {
		log.Infof("Setting sample '%s' to outputter '%s'", item.S.Name, item.S.Output.Outputter)
		switch item.S.Output.Outputter {
		case "stdout":
			gout[num] = new(stdout)
		case "devnull":
			gout[num] = new(devnull)
		case "file":
			gout[num] = new(file)
		case "http":
			gout[num] = new(httpout)
		case "buf":
			gout[num] = new(buf)
		case "splunktcp":
			gout[num] = new(splunktcp)
		default:
			gout[num] = new(stdout)
		}
	}
	return gout[num]
}
