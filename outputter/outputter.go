package outputter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/template"
)

var (
	EventsWritten map[string]int64
	BytesWritten  map[string]int64
	Mutex         sync.RWMutex
	lastTS        time.Time
	rotchan       chan *config.OutputStats
	gout          [config.MaxOutputThreads]config.Outputter
	lasterr       [config.MaxOutputThreads]lastError
	rotInterval   int
	cacheBufs     map[string]*bytes.Buffer
	cacheMutex    sync.RWMutex
)

type lastError struct {
	when  time.Time
	err   error
	count int64
}

func init() {
	EventsWritten = make(map[string]int64)
	BytesWritten = make(map[string]int64)
	cacheBufs = make(map[string]*bytes.Buffer)
}

// ROT starts the Read Out Thread which will log statistics about what's being output
// ROT is intended to be started as a goroutine which will log output every c.
func ROT(c *config.Config) {
	rotInterval = c.Global.ROTInterval
	rotchan = make(chan *config.OutputStats)
	go readStats()

	lastEventsWritten := make(map[string]int64)
	lastBytesWritten := make(map[string]int64)
	var gbday, eventssec, kbytessec float64
	var tempEW, tempBW int64
	lastTS = time.Now()
	for {
		timer := time.NewTimer(time.Duration(rotInterval) * time.Second)
		<-timer.C
		n := time.Now()
		eventssec = 0
		kbytessec = 0
		Mutex.RLock()
		for k := range BytesWritten {
			tempEW = EventsWritten[k]
			tempBW = BytesWritten[k]
			eventssec += float64(tempEW-lastEventsWritten[k]) / float64(int(n.Sub(lastTS))/int(time.Second)/rotInterval)
			kbytessec += float64(tempBW-lastBytesWritten[k]) / float64(int(n.Sub(lastTS))/int(time.Second)/rotInterval) / 1024
			gbday = (kbytessec * 60 * 60 * 24) / 1024 / 1024
			lastEventsWritten[k] = tempEW
			lastBytesWritten[k] = tempBW
		}
		Mutex.RUnlock()
		log.WithFields(log.Fields{
			"eventsSec": eventssec,
			"kbytesSec": kbytessec,
			"gbDay":     gbday,
		}).Infof("Events/Sec: %.2f Kilobytes/Sec: %.2f GB/Day: %.2f", eventssec, kbytessec, gbday)
		lastTS = n
	}
}

// ReadFinal outputs final statistics about our run
func ReadFinal() {
	totalEvents := int64(0)
	totalBytes := int64(0)
	Mutex.RLock()
	for k := range BytesWritten {
		totalEvents += EventsWritten[k]
		totalBytes += BytesWritten[k]
	}
	totalGBytes := float64(totalBytes / 1024 / 1024 / 1024)
	Mutex.RUnlock()
	log.WithField("totalEvents", totalEvents).Infof("Total Events Written: %d", totalEvents)
	log.WithField("totalBytes", totalBytes).Infof("Total Bytes Written: %d", totalBytes)
	log.WithField("totalGBytes", totalGBytes).Infof("Total Gigabytes Written: %.2f", totalGBytes)
}

func readStats() {
	for {
		select {
		case os := <-rotchan:
			Mutex.Lock()
			BytesWritten[os.SampleName] += os.BytesWritten
			EventsWritten[os.SampleName] += os.EventsWritten
			Mutex.Unlock()
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

func write(item *config.OutQueueItem) {
	var bytesCounter int64
	var w io.Writer
	cacheBuf, cacheBufOk := cacheBufs[item.S.Name]
	useCache := item.Cache.UseCache && cacheBufOk // if we aren't in the cache yet, just output cached generated events
	if item.Cache.UseCache && !useCache {
		log.Infof("cache miss")
	}
	if item.Cache.SetCache {
		if !cacheBufOk {
			cacheBuf = &bytes.Buffer{}
			log.Infof("Setting cache")
			cacheBufs[item.S.Name] = cacheBuf
		}
		cacheBuf.Reset()
		w = cacheBuf
	} else if !useCache {
		w = item.IO.W
	}
	defer item.IO.W.Close()
	if !useCache {
		item.Cache.RLock()
		switch item.S.Output.OutputTemplate {
		case "raw", "json", "splunkhec", "rfc3164", "rfc5424", "elasticsearch":
			for _, line := range item.Events {
				var tempbytes int
				var err error
				if item.S.Output.Outputter != "devnull" {
					switch item.S.Output.OutputTemplate {
					case "raw":
						tempbytes, err = io.WriteString(w, line["_raw"])
					case "json":
						jb, err := json.Marshal(line)
						if err != nil {
							log.Errorf("Error marshaling json: %s", err)
						}
						tempbytes, err = w.Write(jb)
					case "splunkhec":
						if _, ok := line["_raw"]; ok {
							line["event"] = line["_raw"]
							delete(line, "_raw")
						}
						if _, ok := line["_time"]; ok {
							line["time"] = line["_time"]
							delete(line, "_time")
						}
						// TODO Refactor to avoid copy pasta, being lazy for now
						jb, err := json.Marshal(line)
						if err != nil {
							log.Errorf("Error marshaling json: %s", err)
						}
						tempbytes, err = w.Write(jb)
					case "rfc3164":
						tempbytes, err = io.WriteString(w, fmt.Sprintf("<%s>%s %s %s[%s]: %s", line["priority"], line["_time"], line["host"], line["tag"], line["pid"], line["_raw"]))
					case "rfc5424":
						kv := "-"
						for k, v := range line {
							if k != "_raw" && k != "_time" && k != "priority" && k != "host" && k != "appName" && k != "pid" && k != "tag" {
								kv = kv + fmt.Sprintf("%s=\"%s\" ", k, v)
							}
						}
						if len(kv) != 1 {
							kv = fmt.Sprintf("[meta %s]", kv[1:len(kv)-1])
						}
						tempbytes, err = io.WriteString(w, fmt.Sprintf("<%s>%d %s %s %s %s - %s %s", line["priority"], 1, line["_time"], line["host"], line["appName"], line["pid"], kv, line["_raw"]))
					case "elasticsearch":
						_, err := io.WriteString(w, fmt.Sprintf("{ \"index\": { \"_index\": \"%s\", \"_type\": \"doc\" } }\n", line["index"]))
						if err != nil {
							break
						}
						if _, ok := line["_raw"]; ok {
							line["message"] = line["_raw"]
							delete(line, "_raw")
						}
						jb, err := json.Marshal(line)
						if err != nil {
							break
						}
						tempbytes, err = w.Write(jb)
					}
					if err != nil {
						log.Errorf("Error writing to IO Buffer: %s", err)
					}
				} else {
					tempbytes = len(line["_raw"])
				}
				bytesCounter += int64(tempbytes) + 1
				if item.S.Output.Outputter != "devnull" && item.S.Output.Outputter != "kafka" {
					_, err = io.WriteString(w, "\n")
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
			bytesCounter += int64(getLine("header", item.S, item.Events[0], w))
			// log.Debugf("Out Queue Item %#v", item)
			var last int
			for i, line := range item.Events {
				bytesCounter += int64(getLine("row", item.S, line, w))
				last = i
			}
			bytesCounter += int64(getLine("footer", item.S, item.Events[last], w))
		}
		item.Cache.RUnlock()
	}
	if useCache || item.Cache.SetCache {
		tempBytes, err := item.IO.W.Write(cacheBufs[item.S.Name].Bytes())
		if err != nil {
			log.Errorf("Error reading from cache buffer: %s", err)
		}
		bytesCounter = int64(tempBytes)
		// log.Infof("Used cache, sent %d events and %d bytes", len(item.Events), bytesCounter)
	}
	Account(int64(len(item.Events)), bytesCounter, item.S.Name)
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
				err := out.Close()
				if err != nil {
					log.Errorf("Error closing output for sample '%s': %s", lastS.Name, err)
				}
				gout[num] = nil
			}
			oqs <- 1
			break
		}
		out = setup(generator, item, num)
		if len(item.Events) > 0 {
			go write(item)
			err := out.Send(item)
			if err != nil {
				logErr := false
				if lasterr[num].err == nil {
					lasterr[num].err = err
					lasterr[num].when = time.Now()
					lasterr[num].count = 1
					logErr = true
				} else if time.Since(lasterr[num].when) > time.Duration(int64(rotInterval))*time.Second {
					lasterr[num].when = time.Now()
					logErr = true
				} else {
					lasterr[num].count++
				}
				if logErr {
					log.Errorf("Error with Send(): %s. %d errors in the last %d second. Closing Output.", err, lasterr[num].count, rotInterval)
					err = out.Close()
					if err != nil {
						log.Errorf("Error closing output: %s", err)
					}
					lasterr[num].count = 0
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
		log.Infof("Setting outputter %d to outputter '%s'", num, item.S.Output.Outputter)
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
		case "network":
			gout[num] = new(network)
		case "kafka":
			gout[num] = new(kafkaout)
		default:
			gout[num] = new(stdout)
		}
	}
	return gout[num]
}
