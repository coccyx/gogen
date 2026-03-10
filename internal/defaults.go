package internal

import "time"

// ProfileOn determines whether we should run a CPU profiler for perf optimization
const ProfileOn = false

// DefaultField is the default field to replace a token in
const DefaultField = "_raw"

// Default global values
const defaultGeneratorWorkers = 1
const defaultOutputWorkers = 1
const defaultROTInterval = 1
const defaultOutputter = "stdout"
const defaultOutputTemplate = "raw"

// Default Sample values
const defaultGenerator = "sample"
const defaultEarliest = "now"
const defaultLatest = "now"

const defaultRater = "default"

// Default file output values
const defaultFileName = "/tmp/test.log"
const defaultMaxBytes = 10485760
const defaultBackupFiles = 5
const defaultTopic = "defaultTopic"

// Default buffer size
const defaultBufferBytes = 4096

// Default timeout for network connections
const defaultTimeout = time.Duration(10 * time.Second)

// MaxOutputThreads defines how large an array we'll define for output threads
const MaxOutputThreads = 100

// defaultGenQueueLength defines how many items can be in the Generator queue at a given time
const defaultGenQueueLength = 50

// defaultOutQueueLength defines how many items can be in the Output queue at a given time
const defaultOutQueueLength = 10

// defaultSyslogPriority defines the default value for the priority field in syslog
// This is the user facilitiy (1 << 3 == 8) at INFO (6) level, (8+6)
const defaultSyslogPriority = 14

// configExtensions defines the file extensions accepted for YAML/JSON config files
var configExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}

var (
	defaultCSVTemplate       *Template
	defaultJSONTemplate      *Template
	defaultSplunkHECTemplate *Template
	defaultRawTemplate       *Template

	defaultRaterConfig       *RaterConfig
	defaultConfigRaterConfig *RaterConfig
)

// uniformRateMap returns a map[int]float64 with keys 0..n-1 all set to 1.0.
func uniformRateMap(n int) map[int]float64 {
	m := make(map[int]float64, n)
	for i := 0; i < n; i++ {
		m[i] = 1.0
	}
	return m
}

func init() {
	defaultCSVTemplate = &Template{
		Name:   "csv",
		Header: `{{ keys . | join "," }}`,
		Row:    `{{ values . | join "," }}`,
		Footer: "",
	}
	defaultJSONTemplate = &Template{
		Name:   "json",
		Header: "",
		Row:    `{{ json . | printf "%s" }}`,
		Footer: "",
	}
	defaultSplunkHECTemplate = &Template{
		Name:   "splunkhec",
		Header: "",
		Row:    `{{ splunkhec . | printf "%s" }}`,
		Footer: "",
	}
	defaultRawTemplate = &Template{
		Name:   "raw",
		Header: "",
		Row:    `{{ ._raw }}`,
		Footer: "",
	}

	defaultRaterConfig = &RaterConfig{
		Name: "default",
		Type: "native",
	}
	defaultConfigRaterConfig = &RaterConfig{
		Name: "config",
		Type: "config",
		Options: map[string]interface{}{
			"HourOfDay":    uniformRateMap(24),
			"DayOfWeek":    uniformRateMap(7),
			"MinuteOfHour": uniformRateMap(60),
		},
	}
}
