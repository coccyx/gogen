package outputter

import (
	"encoding/json"
	"io"
	"math/rand"
	"time"

	"github.com/coccyx/go-s2s/s2s"
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/template"
)

var (
	EventsWritten map[string]int64
	BytesWritten  map[string]int64
	lastTS        time.Time
	rotchan       chan *config.OutputStats
	gout          [config.MaxOutputThreads]config.Outputter
)

func init() {
	EventsWritten = make(map[string]int64)
	BytesWritten = make(map[string]int64)
}

// ROT starts the Read Out Thread which will log statistics about what's being output
// ROT is intended to be started as a goroutine which will log output every c.
func ROT(c *config.Config, statsChan chan config.OutputStats) {
	rotchan = make(chan *config.OutputStats)
	go readStats()

	lastEventsWritten := make(map[string]int64)
	lastBytesWritten := make(map[string]int64)
	var gbday, eventssec, kbytessec float64
	var tempEW, tempBW int64
	lastTS = time.Now()
	for {
		timer := time.NewTimer(time.Duration(c.Global.ROTInterval) * time.Second)
		<-timer.C
		n := time.Now()
		eventssec = 0
		kbytessec = 0
		for k := range BytesWritten {
			tempEW = EventsWritten[k]
			tempBW = BytesWritten[k]
			if c.Global.Web {
				statsChan <- config.OutputStats{
					EventsWritten: tempEW - lastEventsWritten[k],
					BytesWritten:  tempBW - lastBytesWritten[k],
					SampleName:    k,
				}
			}
			eventssec += float64(tempEW-lastEventsWritten[k]) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval)
			kbytessec += float64(tempBW-lastBytesWritten[k]) / float64(int(n.Sub(lastTS))/int(time.Second)/c.Global.ROTInterval) / 1024
			gbday = (kbytessec * 60 * 60 * 24) / 1024 / 1024
			lastEventsWritten[k] = tempEW
			lastBytesWritten[k] = tempBW
		}
		log.Infof("Events/Sec: %.2f Kilobytes/Sec: %.2f GB/Day: %.2f", eventssec, kbytessec, gbday)
		lastTS = n
	}
}

func readStats() {
	for {
		select {
		case os := <-rotchan:
			BytesWritten[os.SampleName] += os.BytesWritten
			EventsWritten[os.SampleName] += os.EventsWritten
		}
	}
}

// Account sends eventsWritten and bytesWritten to the readStats() thread
func Account(eventsWritten int64, bytesWritten int64, sampleName string) {
	os := new(config.OutputStats)
	os.EventsWritten = eventsWritten
	os.BytesWritten = bytesWritten
	os.SampleName = sampleName
	rotchan <- os
}

// Start starts an output thread and runs until notified to shut down
func Start(oq chan *config.OutQueueItem, oqs chan int, num int) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	var lastS *config.Sample
	var out config.Outputter
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
								tempbytes, err = io.WriteString(item.IO.W, line["_raw"])
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
						bytes += int64(getLine("row", item.S, line, item.IO.W))
						last = i
					}
					bytes += int64(getLine("footer", item.S, item.Events[last], item.IO.W))
				}
				Account(int64(len(item.Events)), bytes, item.S.Name)
			}()
			err := out.Send(item)
			if err != nil {
				log.Errorf("Error with Send(): %s", err)
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
