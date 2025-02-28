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
const defaultRandomizeEvents = false
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

var (
	defaultCSVTemplate       *Template
	defaultJSONTemplate      *Template
	defaultSplunkHECTemplate *Template
	defaultRawTemplate       *Template

	defaultRaterConfig       *RaterConfig
	defaultConfigRaterConfig *RaterConfig
)

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
			"HourOfDay": map[int]float64{
				0:  1.0,
				1:  1.0,
				2:  1.0,
				3:  1.0,
				4:  1.0,
				5:  1.0,
				6:  1.0,
				7:  1.0,
				8:  1.0,
				9:  1.0,
				10: 1.0,
				11: 1.0,
				12: 1.0,
				13: 1.0,
				14: 1.0,
				15: 1.0,
				16: 1.0,
				17: 1.0,
				18: 1.0,
				19: 1.0,
				20: 1.0,
				21: 1.0,
				22: 1.0,
				23: 1.0,
			},
			"DayOfWeek": map[int]float64{
				0: 1.0,
				1: 1.0,
				2: 1.0,
				3: 1.0,
				4: 1.0,
				5: 1.0,
				6: 1.0,
			},
			"MinuteOfHour": map[int]float64{
				0:  1.0,
				1:  1.0,
				2:  1.0,
				3:  1.0,
				4:  1.0,
				5:  1.0,
				6:  1.0,
				7:  1.0,
				8:  1.0,
				9:  1.0,
				10: 1.0,
				11: 1.0,
				12: 1.0,
				13: 1.0,
				14: 1.0,
				15: 1.0,
				16: 1.0,
				17: 1.0,
				18: 1.0,
				19: 1.0,
				20: 1.0,
				21: 1.0,
				22: 1.0,
				23: 1.0,
				24: 1.0,
				25: 1.0,
				26: 1.0,
				27: 1.0,
				28: 1.0,
				29: 1.0,
				30: 1.0,
				31: 1.0,
				32: 1.0,
				33: 1.0,
				34: 1.0,
				35: 1.0,
				36: 1.0,
				37: 1.0,
				38: 1.0,
				39: 1.0,
				40: 1.0,
				41: 1.0,
				42: 1.0,
				43: 1.0,
				44: 1.0,
				45: 1.0,
				46: 1.0,
				47: 1.0,
				48: 1.0,
				49: 1.0,
				50: 1.0,
				51: 1.0,
				52: 1.0,
				53: 1.0,
				54: 1.0,
				55: 1.0,
				56: 1.0,
				57: 1.0,
				58: 1.0,
				59: 1.0,
			},
		},
	}
}
