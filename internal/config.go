package internal

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/template"
	"github.com/coccyx/timeparser"
	"github.com/kr/pretty"
)

// Config is a struct representing a Singleton which contains a copy of the running config
// across all processes.  Should mirror the structure of $GOGEN_HOME/configs/default/global.yml
type Config struct {
	Global      Global             `json:"global,omitempty" yaml:"global,omitempty"`
	Samples     []*Sample          `json:"samples" yaml:"samples"`
	Mix         []*Mix             `json:"mix" yaml:"mix"`
	Templates   []*Template        `json:"templates,omitempty" yaml:"templates,omitempty"`
	Raters      []*RaterConfig     `json:"raters,omitempty" yaml:"raters,omitempty"`
	Generators  []*GeneratorConfig `json:"generators,omitempty" yaml:"generators,omitempty"`
	initialized bool
	cc          ConfigConfig

	// Exported but internal use variables
	Timezone *time.Location `json:"-" yaml:"-"`
	Buf      bytes.Buffer   `json:"-" yaml:"-"`
}

// Global represents global configuration options which apply to all of gogen
type Global struct {
	UTC                  bool     `json:"utc,omitempty" yaml:"utc,omitempty"`
	Debug                bool     `json:"debug,omitempty" yaml:"debug,omitempty"`
	Verbose              bool     `json:"verbose,omitempty" yaml:"verbose,omitempty"`
	GeneratorWorkers     int      `json:"generatorWorkers,omitempty" yaml:"generatorWorkers,omitempty"`
	OutputWorkers        int      `json:"outputWorkers,omitempty" yaml:"outputWorkers,omitempty"`
	GeneratorQueueLength int      `json:"generatorQueueLength,omitempty" yaml:"generatorQueueLength,omitempty"`
	OutputQueueLength    int      `json:"outputQueueLength,omitempty" yaml:"outputQueueLength,omitempty"`
	ROTInterval          int      `json:"rotInterval,omitempty" yaml:"rotInterval,omitempty"`
	Output               Output   `json:"output,omitempty" yaml:"output,omitempty"`
	SamplesDir           []string `json:"samplesDir,omitempty" yaml:"samplesDir,omitempty"`
	AddTime              bool     `json:"addTime,omitempty" yaml:"addTime,omitempty"`
	CacheIntervals       int      `json:"cacheIntervals,omitempty" yaml:"cacheIntervals,omitempty"`
}

// Output represents configuration for outputting data
type Output struct {
	FileName       string            `json:"fileName,omitempty" yaml:"fileName,omitempty"`
	MaxBytes       int64             `json:"maxBytes,omitempty" yaml:"maxBytes,omitempty"`
	BackupFiles    int               `json:"backupFiles,omitempty" yaml:"backupFiles,omitempty"`
	BufferBytes    int               `json:"bufferBytes,omitempty" yaml:"bufferBytes,omitempty"`
	Outputter      string            `json:"outputter,omitempty" yaml:"outputter,omitempty"`
	OutputTemplate string            `json:"outputTemplate,omitempty" yaml:"outputTemplate,omitempty"`
	Endpoints      []string          `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Protocol       string            `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Timeout        time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Topic          string            `json:"topic,omitempty" yaml:"topic,omitempty"`

	// Used for S2S Outputter to maintain state of unique host, source, sourcetype combos
	channelIdx int
	channelMap map[string]int
}

// ConfigConfig represents options to pass to NewConfig
type ConfigConfig struct {
	Home       string
	GlobalFile string
	ConfigDir  string
	SamplesDir string
	FullConfig string
	Export     bool
}

// Share allows accessing the share module from Config without a circular dependency
type Share interface {
	PullFile(gogen string, filename string)
}

var instance *Config
var share Share

// setDefault sets *ptr to defaultVal if *ptr is the zero value for its type.
func setDefault[T comparable](ptr *T, defaultVal T) {
	var zero T
	if *ptr == zero {
		*ptr = defaultVal
	}
}

func getConfig() *Config {
	if instance == nil {
		instance = &Config{initialized: false}
	}
	return instance
}

// ResetConfig will delete any current running config
func ResetConfig() {
	log.Debugf("Resetting config to fresh config")
	instance = nil
}

// NewConfig is a singleton constructor which will return a pointer to a global instance of Config
// Environment variables will impact the function of how we configure ourselves
// GOGEN_HOME: Change home directory where we will search for configs
// GOGEN_SAMPLES_DIR: Change where we will look for samples
// GOGEN_ALWAYS_REFRESH: Do not use singleton pattern and reparse configs
// GOGEN_FULLCONFIG: The reference is to a full exported config, so don't resolve or validate
// GOGEN_EXPORT: Don't set defaults for export
func NewConfig() *Config {
	var cc ConfigConfig

	cc.Home = os.Getenv("GOGEN_HOME")
	if len(cc.Home) == 0 {
		log.Debug("GOGEN_HOME not set, setting to '.'")
		cc.Home = "."
		os.Setenv("GOGEN_HOME", ".")
	}
	log.Debugf("Home: %v", cc.Home)

	if os.Getenv("GOGEN_ALWAYS_REFRESH") != "1" {
		c := getConfig()
		if c.initialized {
			return c
		}
	} else {
		log.Debugf("Always refresh on, using fresh config")
	}

	cc.ConfigDir = os.Getenv("GOGEN_CONFIG_DIR")
	if len(cc.ConfigDir) == 0 {
		cc.ConfigDir = filepath.Join(cc.Home, "config")
		log.Debugf("GOGEN_CONFIG_DIR not set, setting to '%s'", cc.ConfigDir)
	}

	cc.SamplesDir = os.Getenv("GOGEN_SAMPLES_DIR")
	if len(cc.SamplesDir) == 0 {
		cc.SamplesDir = filepath.Join(cc.ConfigDir, "samples")
		log.Debugf("GOGEN_SAMPLES_DIR not set, setting to '%s'", cc.SamplesDir)
	}

	cc.FullConfig = os.Getenv("GOGEN_FULLCONFIG")
	cc.GlobalFile = os.Getenv("GOGEN_GLOBAL")

	if os.Getenv("GOGEN_EXPORT") == "1" {
		cc.Export = true
	}
	instance = BuildConfig(cc)
	return instance
}

// BuildConfig builds a new config object from the passed ConfigConfig
func BuildConfig(cc ConfigConfig) *Config {
	c := &Config{initialized: false, cc: cc}

	// Setup timezone
	c.Timezone, _ = time.LoadLocation("Local")

	if len(cc.FullConfig) > 0 {
		cc.FullConfig = os.ExpandEnv(cc.FullConfig)
		if strings.HasPrefix(cc.FullConfig, "http") {
			log.Infof("Fetching config from '%s'", cc.FullConfig)
			if err := c.parseWebConfig(&c, cc.FullConfig); err != nil {
				log.Panic(err)
			}
		} else {
			_, err := os.Stat(cc.FullConfig)
			if err != nil {
				log.Fatalf("Cannot stat file %s", cc.FullConfig)
			}
			if err := c.parseFileConfig(&c, cc.FullConfig); err != nil {
				log.Panic(err)
			}
			if !strings.Contains(cc.FullConfig, "tests") {
				c.Global.SamplesDir = append(c.Global.SamplesDir, filepath.Dir(cc.FullConfig))
			}
		}
		for i := 0; i < len(c.Samples); i++ {
			c.Samples[i].realSample = true
		}
	} else {
		if len(cc.GlobalFile) > 0 {
			if err := c.parseFileConfig(&c.Global, cc.GlobalFile); err != nil {
				log.Panic(err)
			}
		}
	}
	setDefault(&c.Global.ROTInterval, defaultROTInterval)
	// Don't set defaults if we're exporting
	if !cc.Export {
		// Setup defaults for global
		setDefault(&c.Global.GeneratorWorkers, defaultGeneratorWorkers)
		setDefault(&c.Global.OutputWorkers, defaultOutputWorkers)
		setDefault(&c.Global.GeneratorQueueLength, defaultGenQueueLength)
		setDefault(&c.Global.OutputQueueLength, defaultOutQueueLength)
		setDefault(&c.Global.Output.Outputter, defaultOutputter)
		setDefault(&c.Global.Output.OutputTemplate, defaultOutputTemplate)

		// Setup defaults for outputs
		setDefault(&c.Global.Output.FileName, defaultFileName)
		setDefault(&c.Global.Output.BackupFiles, defaultBackupFiles)
		setDefault(&c.Global.Output.MaxBytes, defaultMaxBytes)
		setDefault(&c.Global.Output.BufferBytes, defaultBufferBytes)
		setDefault(&c.Global.Output.Timeout, defaultTimeout)
		setDefault(&c.Global.Output.Topic, defaultTopic)
		if len(c.Global.Output.Headers) == 0 {
			c.Global.Output.Headers = map[string]string{
				"Content-Type": "application/json",
			}
		}
		// CacheIntervals cannot be negative, just override to zero if someone's being silly
		if c.Global.CacheIntervals < 0 {
			c.Global.CacheIntervals = 0
		}

		c.Global.Output.channelIdx = 0
		c.Global.Output.channelMap = make(map[string]int)

		// Add default templates
		templates := []*Template{defaultCSVTemplate, defaultJSONTemplate, defaultSplunkHECTemplate, defaultRawTemplate}
		c.Templates = append(c.Templates, templates...)
		for _, t := range c.Templates {
			if len(t.Header) > 0 {
				if err := template.New(t.Name+"_header", t.Header); err != nil {
					log.Errorf("Error creating header template for '%s': %v", t.Name, err)
				}
			}
			if err := template.New(t.Name+"_row", t.Row); err != nil {
				log.Errorf("Error creating row template for '%s': %v", t.Name, err)
			}
			if len(t.Footer) > 0 {
				if err := template.New(t.Name+"_footer", t.Footer); err != nil {
					log.Errorf("Error creating footer template for '%s': %v", t.Name, err)
				}
			}
		}
	}

	if len(cc.FullConfig) == 0 {
		loadConfigDir(c, cc.ConfigDir, "templates", &c.Templates)
		loadConfigDir(c, cc.ConfigDir, "raters", &c.Raters)
		loadConfigDir(c, cc.ConfigDir, "generators", &c.Generators)

		c.readSamplesDir(cc.SamplesDir)
	}

	// Configuration allows for finding additional samples directories and reading them
	for _, sd := range c.Global.SamplesDir {
		log.Debugf("Reading samplesDir from Global SamplesDir: %s", sd)
		c.readSamplesDir(sd)
	}

	// Add a clause to allow copying from other samples
	for i := 0; i < len(c.Samples); i++ {
		if len(c.Samples[i].FromSample) > 0 {
			for j := 0; j < len(c.Samples); j++ {
				if c.Samples[j].Name == c.Samples[i].FromSample {
					log.Debugf("Copying sample '%s' to sample '%s' because fromSample set", c.Samples[j].Name, c.Samples[i].Name)
					ext := filepath.Ext(c.Samples[j].Name)
					if ext == ".csv" || ext == ".sample" {
						c.Samples[i].Lines = c.Samples[j].Lines
					} else {
						tempname := c.Samples[i].Name
						tempcount := c.Samples[i].Count
						tempinterval := c.Samples[i].Interval
						tempendintervals := c.Samples[i].EndIntervals
						tempbegin := c.Samples[i].Begin
						tempend := c.Samples[i].End
						temp := *c.Samples[j]
						c.Samples[i] = &temp
						c.Samples[i].Disabled = false
						c.Samples[i].Name = tempname
						c.Samples[i].FromSample = ""
						if tempcount > 0 {
							c.Samples[i].Count = tempcount
						}
						if tempinterval > 0 {
							c.Samples[i].Interval = tempinterval
						}
						if tempendintervals > 0 {
							c.Samples[i].EndIntervals = tempendintervals
						}
						if len(tempbegin) > 0 {
							c.Samples[i].Begin = tempbegin
						}
						if len(tempend) > 0 {
							c.Samples[i].End = tempend
						}
						break
					}
				}
			}
		}
	}

	// Raters brought in from config will be typed wrong, validate and fixes
	for i := 0; i < len(c.Raters); i++ {
		c.validateRater(c.Raters[i])
	}

	// Allow bringing in generator scripts from a file
	for i := 0; i < len(c.Generators); i++ {
		if c.Generators[i].FileName != "" && c.Generators[i].Script == "" {
			err := c.readGenerator(cc.ConfigDir, c.Generators[i])
			if err != nil {
				log.Fatalf("Error reading generator file: %s", err)
			}
		}
	}

	// Due to data structure differences, we append default raters later in the startup process
	if !cc.Export {
		raters := []*RaterConfig{defaultRaterConfig, defaultConfigRaterConfig}
		for _, r := range raters {
			c.Raters = append(c.Raters, r)
		}
	}

	// Setup time and facility
	c.SetupSystemTokens()

	// There area references from tokens to samples, need to resolve those references
	for i := 0; i < len(c.Samples); i++ {
		c.validate(c.Samples[i])
	}

	c.Clean()

	// Add support for the mix statements
	if !cc.Export {
		for _, m := range c.Mix {
			cc := ConfigConfig{FullConfig: m.Sample, Export: false}
			var nc *Config
			mixExtensions := map[string]bool{".yml": true, ".yaml": true, ".json": true, ".sample": true, ".csv": true}
			if _, ok := mixExtensions[filepath.Ext(m.Sample)]; ok {
				nc = BuildConfig(cc)
				c.mergeMixConfig(nc, m)
			} else {
				PullFile(m.Sample, ".tmp.yml")
				cc = ConfigConfig{FullConfig: ".tmp.yml"}
				nc = BuildConfig(cc)
				c.mergeMixConfig(nc, m)
				os.Remove(".tmp.yml")
			}
		}
	}

	c.initialized = true
	return c
}

func (c *Config) mergeMixConfig(nc *Config, m *Mix) {
	for i := range nc.Samples {
		log.Debugf("Merging config for sample '%s' from mix: %# v", nc.Samples[i].Name, pretty.Formatter(m))
		if c.FindSampleByName(nc.Samples[i].Name) != nil {
			log.Errorf("Sample name '%s' already exists, not adding from mix", nc.Samples[i].Name)
			continue
		}
		if m.Count != 0 {
			nc.Samples[i].Count = m.Count
		}
		if m.Interval != 0 {
			nc.Samples[i].Interval = m.Interval
		}
		if m.Begin != "" {
			nc.Samples[i].Begin = m.Begin
		}
		if m.End != "" {
			nc.Samples[i].End = m.End
		}
		if m.EndIntervals != 0 {
			nc.Samples[i].EndIntervals = m.EndIntervals
		}
		ParseBeginEnd(nc.Samples[i])
		if m.Realtime {
			nc.Samples[i].EndIntervals = 0
			nc.Samples[i].Realtime = true
		}
		log.Debugf("Adding Sample '%s' from mix", nc.Samples[i].Name)
		c.Samples = append(c.Samples, nc.Samples[i])
	}
	for i := range nc.Generators {
		c.Generators = append(c.Generators, nc.Generators[i])
	}
	for i := range nc.Raters {
		c.Raters = append(c.Raters, nc.Raters[i])
	}
}

func (c *Config) readSamplesDir(samplesDir string) {
	// Read all flat file samples
	acceptableExtensions := map[string]bool{".sample": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
		log.Debugf("Loading file sample '%s'", innerPath)
		s := new(Sample)
		s.Name = filepath.Base(innerPath)
		s.Disabled = true

		file, err := os.Open(innerPath)
		if err != nil {
			log.Errorf("Error reading sample file '%s': %s", innerPath, err)
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s.Lines = append(s.Lines, map[string]string{"_raw": scanner.Text()})
		}
		c.Samples = append(c.Samples, s)
		return nil
	})

	// Read all csv file samples
	acceptableExtensions = map[string]bool{".csv": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
		log.Debugf("Loading CSV sample '%s'", innerPath)
		s := new(Sample)
		s.Name = filepath.Base(innerPath)
		s.Disabled = true

		var (
			fields []string
			rows   [][]string
			err    error
		)

		file, err := os.Open(innerPath)
		if err != nil {
			log.Errorf("Error reading sample file '%s': %s", innerPath, err)
			return nil
		}
		defer file.Close()

		reader := csv.NewReader(file)
		if fields, err = reader.Read(); err != nil {
			log.Errorf("Error parsing header row of sample file '%s' as csv: %s", innerPath, err)
			return nil
		}
		if rows, err = reader.ReadAll(); err != nil {
			log.Errorf("Error parsing sample file '%s' as csv: %s", innerPath, err)
			return nil
		}
		for _, row := range rows {
			fieldsmap := map[string]string{}
			for i := 0; i < len(fields); i++ {
				fieldsmap[fields[i]] = row[i]
			}
			s.Lines = append(s.Lines, fieldsmap)
		}
		c.Samples = append(c.Samples, s)
		return nil
	})

	// Read all YAML & JSON samples in $GOGEN_HOME/config/samples directory
	c.walkPath(samplesDir, configExtensions, func(innerPath string) error {
		if c.cc.FullConfig != innerPath {
			log.Debugf("Loading YAML sample '%s'", innerPath)
			s := Sample{}
			if err := c.parseFileConfig(&s, innerPath); err != nil {
				log.Errorf("Error parsing config %s: %s", innerPath, err)
				return nil
			}
			s.realSample = true

			c.Samples = append(c.Samples, &s)
			return nil
		}
		return nil
	})
}

// Brings in a Generator script from a file
func (c *Config) readGenerator(configDir string, g *GeneratorConfig) error {
	// First try to find the file by absolute path
	fullPath := os.ExpandEnv(g.FileName)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		fullPath = os.ExpandEnv(filepath.Join(configDir, "generators", g.FileName))
		_, err = os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("Cannot find generator file for generator '%s'", g.Name)
		}
	} else if err != nil {
		return err
	}
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	g.Script = string(contents)
	return nil
}

// FindRater returns a RaterConfig matched by the passed name
func (c *Config) FindRater(name string) *RaterConfig {
	for _, findr := range c.Raters {
		if findr.Name == name {
			return findr
		}
	}
	log.Errorf("Rater '%s' not found", name)
	return nil
}

// ParseBeginEnd parses the Begin and End settings for a sample
func ParseBeginEnd(s *Sample) {
	// EndIntervals overrides begin and end
	if s.EndIntervals > 0 {
		if s.Interval == 0 {
			s.Interval = 1
		}
		s.Begin = "-" + strconv.Itoa(s.EndIntervals*s.Interval) + "s"
		s.End = "now"
		log.Infof("EndIntervals set at %d, setting Begin to '%s' and end to '%s'", s.EndIntervals, s.Begin, s.End)
	}
	// Setup Begin & End
	// If End is not set, then we're intended to always run in realtime
	if s.End == "" {
		s.Realtime = true
	}
	if s.Begin != "" && s.EndIntervals > 0 {
		s.Realtime = false
	}
	// Cache a time so we can get a delta for parsed begin, end, earliest and latest
	n := time.Now()
	now := func() time.Time {
		return n
	}
	var err error
	if len(s.Begin) > 0 {
		if s.BeginParsed, err = timeparser.TimeParserNow(s.Begin, now); err != nil {
			log.Errorf("Error parsing Begin for sample %s: %v", s.Name, err)
		} else {
			s.Current = s.BeginParsed
			s.Realtime = false
		}
	}
	if len(s.End) > 0 {
		if s.EndParsed, err = timeparser.TimeParserNow(s.End, now); err != nil {
			log.Errorf("Error parsing End for sample %s: %v", s.Name, err)
		}
	} else {
		s.EndParsed = time.Time{}
	}
	log.Infof("Beginning generation at %s; Ending at %s; Realtime: %v", s.BeginParsed, s.EndParsed, s.Realtime)
}

// FindSampleByName finds and returns a pointer to a sample referenced by the passed name
func (c Config) FindSampleByName(name string) *Sample {
	for i := 0; i < len(c.Samples); i++ {
		if c.Samples[i].Name == name {
			return c.Samples[i]
		}
	}
	return nil
}

// convertUTC sets time local to UTC if configured as UTC
func convertUTC(t time.Time) time.Time {
	if instance != nil {
		if instance.Global.UTC {
			return t.UTC()
		}
	}
	return t
}

// Clean frees memory from disabled samples
func (c *Config) Clean() {
	// Clean up disabled and informational samples
	samples := make([]*Sample, 0)
	for i := 0; i < len(c.Samples); i++ {
		if c.Samples[i].realSample && !c.Samples[i].Disabled {
			samples = append(samples, c.Samples[i])
		}
	}
	c.Samples = samples
	debug.FreeOSMemory()
}

// WriteTempConfigFileFromString writes a configuration string to a temporary file and returns the filename
func WriteTempConfigFileFromString(config string) string {
	tmpfile, err := os.CreateTemp("", "gogen-test-*.yml")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write([]byte(config)); err != nil {
		tmpfile.Close()
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}

func SetupFromFile(filename string) {
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", filename)
}

func SetupFromString(configStr string) {
	configFile := WriteTempConfigFileFromString(configStr)
	os.Setenv("GOGEN_HOME", "..")
	os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
	os.Setenv("GOGEN_FULLCONFIG", configFile)
}

func CleanupConfigAndEnvironment() {
	os.Unsetenv("GOGEN_HOME")
	os.Unsetenv("GOGEN_ALWAYS_REFRESH")
	if strings.Contains(os.Getenv("GOGEN_FULLCONFIG"), "gogen-test-") {
		os.Remove(os.Getenv("GOGEN_FULLCONFIG"))
	}
	os.Unsetenv("GOGEN_FULLCONFIG")
}
