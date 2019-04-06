package internal

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/template"
	"github.com/coccyx/timeparser"
	"github.com/kr/pretty"
	lua "github.com/yuin/gopher-lua"
	yaml "gopkg.in/yaml.v2"
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
var once sync.Once
var share Share

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
		if cc.FullConfig[0:4] == "http" {
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
			// if filepath.Dir(cc.FullConfig) != "." && !strings.Contains(cc.FullConfig, "tests") {
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
	if c.Global.ROTInterval == 0 {
		c.Global.ROTInterval = defaultROTInterval
	}
	// Don't set defaults if we're exporting
	if !cc.Export {
		//
		// Setup defaults for global
		//
		if c.Global.GeneratorWorkers == 0 {
			c.Global.GeneratorWorkers = defaultGeneratorWorkers
		}
		if c.Global.OutputWorkers == 0 {
			c.Global.OutputWorkers = defaultOutputWorkers
		}
		if c.Global.GeneratorQueueLength == 0 {
			c.Global.GeneratorQueueLength = defaultGenQueueLength
		}
		if c.Global.OutputQueueLength == 0 {
			c.Global.OutputQueueLength = defaultOutQueueLength
		}
		if c.Global.Output.Outputter == "" {
			c.Global.Output.Outputter = defaultOutputter
		}
		if c.Global.Output.OutputTemplate == "" {
			c.Global.Output.OutputTemplate = defaultOutputTemplate
		}

		//
		// Setup defaults for outputs
		//
		if c.Global.Output.FileName == "" {
			c.Global.Output.FileName = defaultFileName
		}
		if c.Global.Output.BackupFiles == 0 {
			c.Global.Output.BackupFiles = defaultBackupFiles
		}
		if c.Global.Output.MaxBytes == 0 {
			c.Global.Output.MaxBytes = defaultMaxBytes
		}
		if c.Global.Output.BufferBytes == 0 {
			c.Global.Output.BufferBytes = defaultBufferBytes
		}
		if c.Global.Output.Timeout == time.Duration(0) {
			c.Global.Output.Timeout = defaultTimeout
		}

		// Add default templates
		templates := []*Template{defaultCSVTemplate, defaultJSONTemplate, defaultSplunkHECTemplate, defaultRawTemplate, defaultModinputTemplate}
		c.Templates = append(c.Templates, templates...)
		for _, t := range c.Templates {
			if len(t.Header) > 0 {
				_ = template.New(t.Name+"_header", t.Header)
			}
			_ = template.New(t.Name+"_row", t.Row)
			if len(t.Footer) > 0 {
				_ = template.New(t.Name+"_footer", t.Footer)
			}
		}
	}

	if len(cc.FullConfig) == 0 {
		// Read all templates in $GOGEN_HOME/config/templates
		fullPath := filepath.Join(cc.ConfigDir, "templates")
		acceptableExtensions := map[string]bool{".yml": true, ".yaml": true, ".json": true}
		c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
			t := new(Template)

			if err := c.parseFileConfig(&t, innerPath); err != nil {
				log.Errorf("Error parsing config %s: %s", innerPath, err)
				return err
			}

			c.Templates = append(c.Templates, t)
			return nil
		})

		// Read all raters in $GOGEN_HOME/config/raters
		fullPath = filepath.Join(cc.ConfigDir, "raters")
		acceptableExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}
		c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
			var r RaterConfig

			if err := c.parseFileConfig(&r, innerPath); err != nil {
				log.Errorf("Error parsing config %s: %s", innerPath, err)
				return err
			}

			c.Raters = append(c.Raters, &r)
			return nil
		})

		// Read all generators in $GOGEN_HOME/config/generators
		fullPath = filepath.Join(cc.ConfigDir, "generators")
		acceptableExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}
		c.walkPath(fullPath, acceptableExtensions, func(innerPath string) error {
			var g GeneratorConfig

			if err := c.parseFileConfig(&g, innerPath); err != nil {
				log.Errorf("Error parsing config %s: %s", innerPath, err)
				return err
			}

			c.Generators = append(c.Generators, &g)
			return nil
		})

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

	// Clean up disabled and informational samples
	samples := make([]*Sample, 0, len(c.Samples))
	for i := 0; i < len(c.Samples); i++ {
		if c.Samples[i].realSample && !c.Samples[i].Disabled {
			samples = append(samples, c.Samples[i])
		}
	}
	c.Samples = samples

	// Add support for the mix statements
	if !cc.Export {
		for _, m := range c.Mix {
			cc := ConfigConfig{FullConfig: m.Sample, Export: false}
			var nc *Config
			acceptableExtensions := map[string]bool{".yml": true, ".yaml": true, ".json": true, ".sample": true, ".csv": true}
			if _, ok := acceptableExtensions[filepath.Ext(m.Sample)]; ok {
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
	acceptableExtensions = map[string]bool{".yml": true, ".yaml": true, ".json": true}
	c.walkPath(samplesDir, acceptableExtensions, func(innerPath string) error {
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

// validate takes a sample and checks against any rules which may cause the configuration to be invalid.
// This hopefully centralizes logic for valid configs, disabling any samples which are not valid and
// preventing this logic from sprawling all over the code base.
// Also finds any references from tokens to other samples and
// updates the token to point to the sample data
// Also fixes up any additional things which are needed, like weighted choice string
// string map to the randutil Choice struct
func (c *Config) validate(s *Sample) {
	if s.realSample {
		s.Buf = &c.Buf
		if s.Generator == "" {
			s.Generator = defaultGenerator
		}
		if len(s.Name) == 0 {
			s.Disabled = true
			s.realSample = false
		} else if len(s.Lines) == 0 && (s.Generator == "sample" || s.Generator == "replay") {
			s.Disabled = true
			s.realSample = false
			log.Errorf("Disabling sample '%s', no lines in sample", s.Name)
		} else {
			s.realSample = true
		}

		// Put the output into the sample for convenience
		s.Output = &c.Global.Output

		// Setup defaults
		if s.Earliest == "" {
			s.Earliest = defaultEarliest
		}
		if s.Latest == "" {
			s.Latest = defaultLatest
		}
		if s.RandomizeEvents == false {
			s.RandomizeEvents = defaultRandomizeEvents
		}
		if s.Field == "" {
			s.Field = DefaultField
		}
		if s.RaterString == "" {
			s.RaterString = defaultRater
		}

		ParseBeginEnd(s)

		//
		// Parse earliest and latest as relative times
		//
		// Cache a time so we can get a delta for parsed begin, end, earliest and latest
		n := time.Now()
		now := func() time.Time {
			return n
		}
		if p, err := timeparser.TimeParserNow(s.Earliest, now); err != nil {
			log.Errorf("Error parsing earliest time '%s' for sample '%s', using Now", s.Earliest, s.Name)
			s.EarliestParsed = time.Duration(0)
		} else {
			s.EarliestParsed = n.Sub(p) * -1
		}
		if p, err := timeparser.TimeParserNow(s.Latest, now); err != nil {
			log.Errorf("Error parsing latest time '%s' for sample '%s', using Now", s.Latest, s.Name)
			s.LatestParsed = time.Duration(0)
		} else {
			s.LatestParsed = n.Sub(p) * -1
		}

		// log.Debugf("Resolving '%s'", s.Name)
		for i := 0; i < len(s.Tokens); i++ {
			if s.Tokens[i].Type == "rated" && s.Tokens[i].RaterString == "" {
				s.Tokens[i].RaterString = "default"
			}
			if s.Tokens[i].Field == "" {
				s.Tokens[i].Field = s.Field
			}
			// If format is template, then create a default token of $tokenname$
			if s.Tokens[i].Format == "template" && s.Tokens[i].Token == "" {
				s.Tokens[i].Token = "$" + s.Tokens[i].Name + "$"
			}
			// log.Debugf("Resolving token '%s' for sample '%s'", s.Tokens[i].Name, s.Name)
			for j := 0; j < len(c.Samples); j++ {
				s.Tokens[i].Parent = s
				s.Tokens[i].luaState = new(lua.LTable)
				if s.Tokens[i].SampleString != "" && s.Tokens[i].SampleString == c.Samples[j].Name {
					log.Debugf("Resolving sample '%s' for token '%s'", c.Samples[j].Name, s.Tokens[i].Name)
					s.Tokens[i].Sample = c.Samples[j]
					// See if a field exists other than _raw, if so, FieldChoice
					otherfield := false
					if len(c.Samples[j].Lines) > 0 {
						for k := range c.Samples[j].Lines[0] {
							if k != "_raw" {
								otherfield = true
								break
							}
						}
					}
					if otherfield {
						// If we're a structured sample and we contain the field "_weight", then we create a weighted choice struct
						// Otherwise we're a fieldChoice
						_, ok := c.Samples[j].Lines[0]["_weight"]
						_, ok2 := c.Samples[j].Lines[0][s.Tokens[i].SrcField]
						if ok && ok2 {
							for _, line := range c.Samples[j].Lines {
								weight, err := strconv.Atoi(line["_weight"])
								if err != nil {
									weight = 0
								}
								s.Tokens[i].WeightedChoice = append(s.Tokens[i].WeightedChoice, WeightedChoice{Weight: weight, Choice: line[s.Tokens[i].SrcField]})
							}
						} else {
							s.Tokens[i].FieldChoice = c.Samples[j].Lines
						}
					} else {
						// s.Tokens[i].WeightedChoice = c.Samples[j].Lines
						temp := make([]string, 0, len(c.Samples[j].Lines))
						for _, line := range c.Samples[j].Lines {
							if _, ok := line["_raw"]; ok {
								if len(line["_raw"]) > 0 {
									temp = append(temp, line["_raw"])
								}
							}
						}
						s.Tokens[i].Choice = temp
					}
					break
				}
			}
		}

		// Begin Validation logic
		if s.EarliestParsed > s.LatestParsed {
			log.Errorf("Earliest time cannot be greater than latest for sample '%s', disabling Sample", s.Name)
			s.Disabled = true
			return
		}
		// If no interval is set, generate one time and exit
		if s.Interval == 0 && s.Generator != "replay" {
			log.Infof("No interval set for sample '%s', setting endIntervals to 1", s.Name)
			s.EndIntervals = 1
		}
		for i, t := range s.Tokens {
			switch t.Type {
			case "random", "rated":
				if t.Replacement == "int" || t.Replacement == "float" {
					if t.Lower > t.Upper {
						log.Errorf("Lower cannot be greater than Upper for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					} else if t.Upper == 0 {
						log.Errorf("Upper cannot be zero for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					}
				} else if t.Replacement == "string" || t.Replacement == "hex" {
					if t.Length == 0 {
						log.Errorf("Length cannot be zero for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
						s.Disabled = true
					}
				} else {
					if t.Replacement != "guid" && t.Replacement != "ipv4" && t.Replacement != "ipv6" {
						log.Errorf("Replacement '%s' is invalid for token '%s' in sample '%s'", t.Replacement, t.Name, s.Name)
						s.Disabled = true
					}
				}
			case "choice":
				if len(t.Choice) == 0 || t.Choice == nil {
					log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
			case "weightedChoice":
				if len(t.WeightedChoice) == 0 || t.WeightedChoice == nil {
					log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
			case "fieldChoice":
				if len(t.FieldChoice) == 0 || t.FieldChoice == nil {
					log.Errorf("Zero choice items for token '%s' in sample '%s', disabling Sample", t.Name, s.Name)
					s.Disabled = true
				}
				for _, choice := range t.FieldChoice {
					if _, ok := choice[t.SrcField]; !ok {
						log.Errorf("Source field '%s' does not exist for token '%s' in row '%#v' in sample '%s', disabling Sample", t.SrcField, t.Name, choice, s.Name)
						s.Disabled = true
						break
					}
				}
			case "script":
				s.Tokens[i].mutex = &sync.Mutex{}
				for k, v := range t.Init {
					vAsNum, err := strconv.ParseFloat(v, 64)
					if err != nil {
						t.luaState.RawSet(lua.LString(k), lua.LNumber(vAsNum))
					} else {
						t.luaState.RawSet(lua.LString(k), lua.LString(v))
					}
				}
			}
		}

		// Check if we are able to do singlepass on this sample by looping through all lines
		// and ensuring we can match all the tokens on each line
		if !s.Disabled {
			s.SinglePass = true

			var tlines []map[string]tokenspos

		outer:
			for _, l := range s.Lines {
				tp := make(map[string]tokenspos)
				for j, t := range s.Tokens {
					// tokenpos 0 first char, 1 last char, 2 token #
					var pos tokenpos
					var err error
					offsets, err := t.GetReplacementOffsets(l[t.Field])
					if err != nil || len(offsets) == 0 {
						log.Infof("Error getting replacements for token '%s' in event '%s', disabling SinglePass", t.Name, l[t.Field])
						s.SinglePass = false
						break outer
					}
					for _, offset := range offsets {
						pos1 := offset[0]
						pos2 := offset[1]
						if pos1 < 0 || pos2 < 0 {
							log.Infof("Token '%s' not found in event '%s', disabling SinglePass", t.Name, l)
							s.SinglePass = false
							break outer
						}
						pos.Pos1 = pos1
						pos.Pos2 = pos2
						pos.Token = j
						tp[t.Field] = append(tp[t.Field], pos)
					}
				}

				// Ensure we don't have any tokens overlapping one another for singlepass
				for _, v := range tp {
					sort.Sort(v)

					lastpos := 0
					lasttoken := ""
					maxpos := 0
					for _, pos := range v {
						// Does the beginning of this token overlap with the end of the last?
						if lastpos > pos.Pos1 {
							log.Infof("Token '%s' extends beyond beginning of token '%s', disabling SinglePass", lasttoken, s.Tokens[pos.Token].Name)
							s.SinglePass = false
							break outer
						}
						// Does the beginning of this token happen before the max we've seen a token before?
						if maxpos > pos.Pos1 {
							log.Infof("Some former token extends beyond the beginning of token '%s', disabling SinglePass", s.Tokens[pos.Token].Name)
							s.SinglePass = false
							break outer
						}
						if pos.Pos2 > maxpos {
							maxpos = pos.Pos2
						}
						lastpos = pos.Pos2
						lasttoken = s.Tokens[pos.Token].Name
					}
				}
				tlines = append(tlines, tp)
			}

			if s.SinglePass {

				// Now loop through each line and each field, breaking it up according to the positions of the tokens
				for i, line := range s.Lines {
					if len(tlines) >= i && len(tlines) > 0 {
						bline := make(map[string][]StringOrToken)
						for field := range line {
							var bfield []StringOrToken
							// Field doesn't exist because no tokens hit that field
							if _, ok := tlines[i][field]; !ok {
								bf := StringOrToken{T: nil, S: line[field]}
								bfield = append(bfield, bf)
							} else {
								lastpos := 0
								// Here, we need to iterate through all the tokens and add StringOrToken for each match
								// Make sure we check for a token a pos 0, we'll put a token first
								for _, tp := range tlines[i][field] {
									if tp.Pos1 == 0 {
										bf := StringOrToken{T: &s.Tokens[tp.Token], S: ""}
										bfield = append(bfield, bf)
										lastpos = tp.Pos2
									} else {
										// Add string from end of last token to the beginning of this one
										bf := StringOrToken{T: nil, S: s.Lines[i][field][lastpos:tp.Pos1]}
										bfield = append(bfield, bf)
										// Add this token
										bf = StringOrToken{T: &s.Tokens[tp.Token], S: ""}
										bfield = append(bfield, bf)
										lastpos = tp.Pos2
									}
								}
								// Add the last string if the last token didn't cover to the end of the string
								if lastpos < len(s.Lines[i][field]) {
									bf := StringOrToken{T: nil, S: s.Lines[i][field][lastpos:]}
									bfield = append(bfield, bf)
								}
							}
							bline[field] = bfield
						}
						s.BrokenLines = append(s.BrokenLines, bline)
					}
				}
			}
		}

		if s.Generator == "replay" {
			// For replay, loop through all events, attempt to find a timestamp in each row, store sleep times in a data structure
			s.ReplayOffsets = make([]time.Duration, len(s.Lines))
			var lastts time.Time
			var avgOffset time.Duration
		outer2:
			for i := 0; i < len(s.Lines); i++ {
			inner2:
				for _, t := range s.Tokens {
					if t.Type == "timestamp" || t.Type == "gotimestamp" || t.Type == "epochtimestamp" {
						offsets, err := t.GetReplacementOffsets(s.Lines[i][t.Field])
						if err != nil || len(offsets) == 0 {
							log.WithFields(log.Fields{
								"token":  t.Name,
								"sample": s.Name,
								"err":    err,
							}).Errorf("Error getting timestamp offsets, disabling sample")
							s.Disabled = true
							break outer2
						}
						pos1 := offsets[0][0]
						pos2 := offsets[0][1]
						ts, err := t.ParseTimestamp(s.Lines[i][t.Field][pos1:pos2])
						if err != nil {
							log.WithFields(log.Fields{
								"token":  t.Name,
								"sample": s.Name,
								"err":    err,
								"event":  s.Lines[0][t.Field],
							}).Errorf("Error parsing timestamp, disabling sample")
							s.Disabled = true
							break outer2
						}
						if i == 0 {
							s.ReplayOffsets[0] = time.Duration(0)
						} else {
							s.ReplayOffsets[i] = lastts.Sub(ts) * -1
							avgOffset = (avgOffset + s.ReplayOffsets[i]) / 2
						}
						lastts = ts
						break inner2
					}
				}
				s.ReplayOffsets[0] = avgOffset
			}
		} else if s.Generator != "sample" {
			for _, g := range c.Generators {
				// TODO If not single threaded, we won't establish state in the sample object
				if g.Name == s.Generator {
					s.LuaMutex = &sync.Mutex{}
					s.CustomGenerator = g
					if g.SingleThreaded {
						s.GeneratorState = NewGeneratorState(s)
					}
				}
			}
			if s.CustomGenerator == nil {
				log.Errorf("Generator '%s' not found for sample '%s', disabling sample", s.Generator, s.Name)
				s.Disabled = true
			}
		}
	}
}

// Returns a copy of the rater with the Options properly cast
func (c *Config) validateRater(r *RaterConfig) {
	configRaterKeys := map[string]bool{
		"HourOfDay":    true,
		"MinuteOfHour": true,
		"DayOfWeek":    true,
	}

	opt := make(map[string]interface{})
	for k, v := range r.Options {
		var newvset interface{}
		if configRaterKeys[k] {
			newv := make(map[int]float64)
			vcast := v.(map[interface{}]interface{})
			for k2, v2 := range vcast {
				k2int := k2.(int)
				v2float, ok := v2.(float64)
				if !ok {
					v2int, ok := v2.(int)
					if !ok {
						log.Fatalf("Rater value '%#v' of key '%#v' for rater '%s' in '%s' is not a float or int", v2, k2, r.Name, k)
					}
					v2float = float64(v2int)
				}
				newv[k2int] = v2float
			}
			newvset = newv
		} else {
			newvset = v
		}
		opt[k] = newvset
	}
	r.Options = opt
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
	contents, err := ioutil.ReadFile(fullPath)
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

// SetupSystemTokens adds tokens like time and facility to samples based on configuration
func (c *Config) SetupSystemTokens() {
	addToken := func(s *Sample, tokenName string, tokenType string, tokenReplacement string) {
		// If there's no _time token, add it to make sure we have a timestamp field in every event
		tokenfound := false
		for _, t := range s.Tokens {
			if t.Name == tokenName {
				tokenfound = true
			}
		}
		for _, l := range s.Lines {
			if _, ok := l[tokenName]; ok {
				tokenfound = true
			}
		}
		if !tokenfound {
			log.Infof("Adding %s field for sample %s", tokenName, s.Name)
			tt := Token{
				Name:   tokenName,
				Type:   tokenType,
				Format: "template",
				Field:  tokenName,
				Token:  fmt.Sprintf("$%s$", tokenName),
				Group:  -1,
			}
			if tokenReplacement != "" {
				tt.Replacement = tokenReplacement
			}
			s.Tokens = append(s.Tokens, tt)
			if s.SinglePass {
				for j := 0; j < len(s.BrokenLines); j++ {
					st := []StringOrToken{
						StringOrToken{T: &tt, S: ""},
					}
					s.BrokenLines[j][tokenName] = st
				}
			}
			for j := 0; j < len(s.Lines); j++ {
				s.Lines[j][tokenName] = fmt.Sprintf("$%s$", tokenName)
			}
		}
	}
	addField := func(s *Sample, name string, value string) {
		for i := 0; i < len(s.Lines); i++ {
			if s.Lines[i][name] == "" {
				s.Lines[i][name] = value
			}
		}
		if s.SinglePass {
			for i := 0; i < len(s.BrokenLines); i++ {
				st := []StringOrToken{
					StringOrToken{T: nil, S: value},
				}
				if _, ok := s.BrokenLines[i][name]; !ok {
					s.BrokenLines[i][name] = st
				}
			}
		}
	}
	syslogOutput := c.Global.Output.OutputTemplate == "rfc3164" || c.Global.Output.OutputTemplate == "rfc5424"
	addTime := c.Global.Output.OutputTemplate == "splunkhec" || c.Global.Output.OutputTemplate == "modinput" || c.Global.AddTime || syslogOutput
	if !c.cc.Export && addTime {
		// Use epochtimestamp for Splunk, or different formats for rfc3164 or rfc5424
		var tokenType string
		var tokenReplacement string
		if addTime && !syslogOutput {
			tokenType = "epochtimestamp"
		} else if c.Global.Output.OutputTemplate == "rfc3164" {
			tokenType = "gotimestamp"
			tokenReplacement = "Jan _2 15:04:05"
		} else if c.Global.Output.OutputTemplate == "rfc5424" {
			tokenType = "gotimestamp"
			tokenReplacement = "2006-01-02T15:04:05.999999Z07:00"
		}
		for i := 0; i < len(c.Samples); i++ {
			s := c.Samples[i]
			addToken(s, "_time", tokenType, tokenReplacement)
			if syslogOutput {
				addField(s, "priority", fmt.Sprintf("%d", defaultSyslogPriority))
				hostname, _ := os.Hostname()
				addField(s, "host", hostname)
				tag := "gogen"
				if s.Lines[0]["sourcetype"] != "" {
					tag = s.Lines[0]["sourcetype"]
				}
				addField(s, "tag", tag)
				addField(s, "pid", fmt.Sprintf("%d", os.Getpid()))
				addField(s, "appName", "gogen")
			}
			// Fixup existing timestamp tokens to all use the same static group, -1
			for j := 0; j < len(s.Tokens); j++ {
				if s.Tokens[j].Type == "timestamp" || s.Tokens[j].Type == "gotimestamp" || s.Tokens[j].Type == "epochtimestamp" {
					s.Tokens[j].Group = -1
				}
			}
		}
	}
}

func (c *Config) walkPath(fullPath string, acceptableExtensions map[string]bool, callback func(string) error) error {
	log.Debugf("walkPath '%s' for extensions: '%v'", fullPath, acceptableExtensions)
	fullPath = os.ExpandEnv(fullPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		fullPath += string(filepath.Separator)
	}
	// filepath.Walk(os.ExpandEnv(fullPath), func(path string, _ os.FileInfo, err error) error {
	// 	log.Debugf("Walking, at %s", path)
	// 	if os.IsNotExist(err) {
	// 		return nil
	// 	} else if err != nil {
	// 		log.Errorf("Error from WalkFunc: %s", err)
	// 		return err
	// 	}
	// 	// Check if extension is acceptable before attempting to parse
	// 	if acceptableExtensions[filepath.Ext(path)] {
	// 		return callback(path)
	// 	}
	// 	return nil
	// })
	files, err := filepath.Glob(fullPath + "*")
	if err != nil {
		return err
	}
	for _, path := range files {
		// log.Debugf("Walking, at %s", path)
		if acceptableExtensions[filepath.Ext(path)] {
			err := callback(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) parseFileConfig(out interface{}, path ...string) error {
	fullPath := filepath.Join(path...)
	log.Debugf("Config Path: %v", fullPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return err
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// log.Debugf("Contents: %s", contents)
	switch filepath.Ext(fullPath) {
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(contents, out); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				log.Errorf("JSON parsing error in file '%s' at offset %d: %v", fullPath, ute.Offset, ute)
			} else {
				log.Errorf("YAML parsing error in file '%s': %v", fullPath, err)
			}
		}
	case ".json":
		if err := json.Unmarshal(contents, out); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				log.Errorf("JSON parsing error in file '%s' at offset %d: %v", fullPath, ute.Offset, ute)
			} else {
				log.Errorf("JSON parsing error in file '%s': %v", fullPath, err)
			}
		}
	}
	// log.Debugf("Out: %#v\n", out)
	return nil
}

func (c *Config) parseWebConfig(out interface{}, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// Try YAML then JSON
	err = yaml.Unmarshal(contents, out)
	if err != nil {
		err = json.Unmarshal(contents, out)
		if err != nil {
			return err
		}
	}
	return nil
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
