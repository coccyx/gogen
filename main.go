package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/run"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/profile"
	"gopkg.in/urfave/cli.v1"
	yaml "gopkg.in/yaml.v2"
)

var c *config.Config
var envVarMap map[string]string

// Version is the version from ./VERSION set by govvv
var Version string

// BuildDate is the build date, set by govvv
var BuildDate string

// GitSummary is the git commit set by govvv
var GitSummary string

func init() {
	envVarMap = map[string]string{
		"info":           "GOGEN_INFO",
		"debug":          "GOGEN_DEBUG",
		"generators":     "GOGEN_GENERATORS",
		"outputters":     "GOGEN_OUTPUTTERS",
		"outputTemplate": "GOGEN_OUTPUTTEMPLATE",
		"outputter":      "GOGEN_OUT",
		"filename":       "GOGEN_FILENAME",
		"topic":          "GOGEN_TOPIC",
		"url":            "GOGEN_URL",
		"splunkHECToken": "GOGEN_HEC_TOKEN",
		"samplesDir":     "GOGEN_SAMPLES_DIR",
		"config":         "GOGEN_CONFIG",
		"addTime":        "GOGEN_ADDTIME",
		"bufferBytes":    "GOGEN_BUFFERBYTES",
		"cacheIntervals": "GOGEN_CACHEINTERVALS",
	}
}

// Setup the running environment
func Setup(clic *cli.Context) {
	if clic.Bool("debug") {
		log.SetDebug(true)
	} else if clic.Bool("info") {
		log.SetInfo()
	}
	if len(clic.String("logFile")) > 0 {
		log.SetOutput(os.ExpandEnv(clic.String("logFile")))
	}
	if clic.Bool("logJson") {
		log.EnableJSONOutput()
	}

	if len(clic.String("configDir")) > 0 {
		os.Setenv("GOGEN_CONFIG_DIR", clic.String("configDir"))
	}
	if len(clic.String("samplesDir")) > 0 {
		os.Setenv("GOGEN_SAMPLES_DIR", clic.String("samplesDir"))
	}
	if len(clic.String("tempDir")) > 0 {
		os.Setenv("GOGEN_TMPDIR", clic.String("tempDir"))
	} else {
		var tmpDir string
		usr, err := user.Current()
		if err != nil {
			tmpDir = os.TempDir()
		} else {
			tmpDir = usr.HomeDir
		}
		os.Setenv("GOGEN_TMPDIR", tmpDir)
	}
	if len(clic.String("config")) > 0 {
		cstr := clic.String("config")
		if cstr[0:4] == "http" || cstr[len(cstr)-3:] == "yml" || cstr[len(cstr)-4:] == "yaml" || cstr[len(cstr)-4:] == "json" {
			os.Setenv("GOGEN_FULLCONFIG", cstr)
		} else {
			config.PullFile(cstr, filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".config.yml"))
			config.ResetConfig()
			os.Setenv("GOGEN_FULLCONFIG", filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".config.yml"))
		}
	}

	c = config.NewConfig()

	if clic.Bool("utc") {
		c.Global.UTC = true
	}

	if clic.Int("generators") > 0 {
		log.Infof("Setting generators to %d", clic.Int("generators"))
		c.Global.GeneratorWorkers = clic.Int("generators")
	}
	if clic.Int("outputters") > 0 {
		log.Infof("Setting generators to %d", clic.Int("outputters"))
		c.Global.OutputWorkers = clic.Int("outputters")
	}
	if clic.Bool("addTime") {
		log.Infof("Adding _time to all Samples")
		c.Global.AddTime = true
	}
	if clic.Int("cacheIntervals") > 0 {
		log.Infof("Setting cacheIntervals to %d", clic.Int("cacheIntervals"))
		c.Global.CacheIntervals = clic.Int("cacheIntervals")
	}
	if clic.Bool("fullRetard") {
		log.Infof("(⊙_ ☉) Going Full Retard (⊙_ ☉)")
		c.Global.CacheIntervals = 2147483647
	}

	for i := 0; i < len(c.Samples); i++ {
		if len(clic.String("outputter")) > 0 {
			log.Infof("Setting outputter to '%s'", clic.String("outputter"))
			if clic.String("outputter") == "tcp" {
				c.Samples[i].Output.Outputter = "network"
				c.Samples[i].Output.Protocol = "tcp"
			} else {
				c.Samples[i].Output.Outputter = clic.String("outputter")
			}
		}
		if len(clic.String("filename")) > 0 {
			log.Infof("Setting filename to '%s'", clic.String("filename"))
			c.Samples[i].Output.FileName = clic.String("filename")
		}
		if len(clic.String("topic")) > 0 {
			log.Infof("Setting topic to '%s'", clic.String("topic"))
			c.Samples[i].Output.Topic = clic.String("topic")
		}
		if len(clic.String("url")) > 0 {
			log.Infof("Setting all endpoint urls to '%s'", clic.String("url"))
			c.Samples[i].Output.Endpoints = []string{clic.String("url")}
		}
		if len(clic.String("splunkHECToken")) > 0 {
			log.Infof("Setting HTTP Header to 'Authorization: Splunk %s'", clic.String("splunkHECToken"))
			if c.Samples[i].Output.Headers == nil {
				c.Samples[i].Output.Headers = make(map[string]string)
			}
			c.Samples[i].Output.Headers["Authorization"] = "Splunk " + clic.String("splunkHECToken")
		}
		if len(clic.String("outputTemplate")) > 0 {
			log.Infof("Setting outputTemplate to '%s'", clic.String("outputTemplate"))
			c.Samples[i].Output.OutputTemplate = clic.String("outputTemplate")
		}
		if clic.Int("bufferBytes") > 0 {
			log.Infof("Setting bufferBytes to '%d'", clic.Int("bufferBytes"))
			c.Samples[i].Output.BufferBytes = clic.Int("bufferBytes")
		}
	}

	// Must call from runtime in case we are overriding AddTime or Facility from command line
	c.SetupSystemTokens()

	// log.Debugf("Global: %#v", c.Global)
	// log.Debugf("Default Tokens: %#v", c.DefaultTokens)
	// log.Debugf("Default Sample: %#v", c.DefaultSample)
	// log.Debugf("Samples: %#v", c.Samples)
	// log.Debugf("Pretty Values %# v\n", pretty.Formatter(c))
	// j, _ := json.MarshalIndent(c, "", "  ")
	// log.Debugf("JSON Config: %s\n", j)
}

func table(l []config.GogenList) {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetColWidth(132)
	t.SetAutoWrapText(false)
	t.SetHeader([]string{"Gogen", "Description"})
	for _, li := range l {
		t.Append([]string{li.Gogen, li.Description})
	}
	t.Render()
}

func main() {
	defer func() {
		os.Remove(filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".config.yml"))
	}()
	if config.ProfileOn {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
		// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	}
	rand.Seed(time.Now().UnixNano())

	app := cli.NewApp()
	app.Name = "gogen"
	app.Usage = "Generate data for demos and testing"
	app.Version = Version
	cli.VersionFlag = cli.BoolFlag{Name: "version"}
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Clint Sharp",
			Email: "clint@typhoon.org",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "gen",
			Usage: "Generate Events",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "sample, s",
					Usage: "Only run sample `name`, can specify multiple",
				},
				cli.IntFlag{
					Name:  "count, c",
					Usage: "Output `number` events",
				},
				cli.IntFlag{
					Name:  "interval, i",
					Usage: "Output every `seconds` seconds",
				},
				cli.IntFlag{
					Name:  "endIntervals, ei",
					Usage: "Only run from `number` intervals",
				},
				cli.StringFlag{
					Name:  "begin, b",
					Usage: "Set begin time, in relative time syntax (e.g. -60m for minus 60 mins, now for now, etc)",
				},
				cli.StringFlag{
					Name:  "end, e",
					Usage: "Set end time, in relative time syntax (e.g. -60m for minus 60 mins, now for now, etc)",
				},
				cli.BoolFlag{
					Name:  "realtime, r",
					Usage: "Set to real time, don't stop until killed",
				},
				cli.BoolFlag{
					Name:  "wait, w",
					Usage: "Wait between intervals when backfilling",
				},
			},
			Action: func(clic *cli.Context) error {
				if len(c.Samples) == 0 {
					fmt.Printf("No samples configured, exiting\n")
					os.Exit(1)
				}
				for i := 0; i < len(c.Samples); i++ {
					if clic.Int("interval") > 0 {
						log.Infof("Setting interval to %d for sample '%s'", clic.Int("interval"), c.Samples[i].Name)
						c.Samples[i].Interval = clic.Int("interval")
					}
					if clic.Int("endIntervals") > 0 {
						log.Infof("Setting endIntervals to %d", clic.Int("endIntervals"))
						c.Samples[i].EndIntervals = clic.Int("endIntervals")
						config.ParseBeginEnd(c.Samples[i])
					}
					if clic.Int("count") > 0 {
						log.Infof("Setting count to %d for sample '%s'", clic.Int("count"), c.Samples[i].Name)
						c.Samples[i].Count = clic.Int("count")
					}
					if len(clic.String("begin")) > 0 {
						log.Infof("Setting begin to %s for sample '%s'", clic.String("begin"), c.Samples[i].Name)
						c.Samples[i].Begin = clic.String("begin")
					}
					if len(clic.String("end")) > 0 {
						log.Infof("Setting end to %s for sample '%s'", clic.String("end"), c.Samples[i].Name)
						c.Samples[i].End = clic.String("end")
					} else {
						if clic.Bool("realtime") {
							c.Samples[i].End = ""
						}
					}
					if len(clic.String("begin")) > 0 || len(clic.String("end")) > 0 {
						if clic.Int("endIntervals") == 0 {
							c.Samples[i].EndIntervals = 0
						}
						config.ParseBeginEnd(c.Samples[i])
					}
					if clic.Bool("realtime") {
						if clic.Int("endIntervals") == 0 {
							c.Samples[i].EndIntervals = 0
						}
						if len(clic.String("begin")) == 0 {
							c.Samples[i].Realtime = true
						}
					}
					if clic.Bool("wait") {
						c.Samples[i].Wait = true
					}
				}
				samplesSlice := clic.StringSlice("sample")
				samplesStr := strings.Join(samplesSlice, " ")
				samplesMap := make(map[string]bool, len(samplesSlice))
				for _, sampleName := range samplesSlice {
					samplesMap[sampleName] = true
				}
				if len(samplesSlice) > 0 {
					log.Infof("Generating only for samples '%s'", samplesStr)
					matched := false
					for i := 0; i < len(c.Samples); i++ {
						if samplesMap[c.Samples[i].Name] {
							matched = true
						} else {
							c.Samples[i].Disabled = true
						}
					}
					if !matched {
						log.Errorf("No sample matched for '%s'", samplesStr)
						os.Exit(1)
					}
					c.Clean()
				}
				run.Run(c)
				return nil
			},
		},
		{
			Name:  "config",
			Usage: "Print config to stdout",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format, f"},
				cli.BoolFlag{
					Name:  "noexport, ne",
					Usage: "Don't set to export",
				},
			},
			Action: func(clic *cli.Context) error {
				if !clic.Bool("noexport") {
					os.Setenv("GOGEN_ALWAYS_REFRESH", "1")
					os.Setenv("GOGEN_EXPORT", "1")
				}
				c = config.NewConfig()
				var outb []byte
				var err error
				if clic.String("format") == "json" {
					if outb, err = json.MarshalIndent(c, "", "  "); err != nil {
						log.Panicf("JSON output error: %v", err)
					}
				} else {
					if outb, err = yaml.Marshal(c); err != nil {
						log.Panicf("YAML output error: %v", err)
					}
				}
				out := string(outb)
				fmt.Print(out)
				return nil
			},
		},
		{
			Name:  "login",
			Usage: "Login to GitHub",
			Action: func(clic *cli.Context) error {
				_ = config.NewGitHub(true)
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "List all published Gogens",
			Action: func(clic *cli.Context) error {
				fmt.Printf("Showing all Gogens:\n\n")
				l := config.List()
				table(l)
				return nil
			},
		},
		{
			Name:  "search",
			Usage: "Search for Gogens",
			Action: func(clic *cli.Context) error {
				var q string
				for _, a := range clic.Args() {
					q += a + " "
				}
				q = strings.TrimRight(q, " ")
				fmt.Printf("Returning results for search: \"%s\"\n\n", q)
				l := config.Search(q)
				if len(l) > 0 {
					table(l)
				} else {
					fmt.Println("   No results found.")
				}
				return nil
			},
		},
		{
			Name:      "info",
			Usage:     "Get info on a specific Gogen",
			ArgsUsage: "[owner/name]",
			Action: func(clic *cli.Context) error {
				if len(clic.Args()) == 0 {
					fmt.Println("Error: Must specify a Gogen in owner/name format")
					os.Exit(1)
				}
				g, err := config.Get(clic.Args()[0])
				if err != nil {
					log.WithError(err).Fatalf("Error retrieving gogen")
				}
				fmt.Printf("Details for Gogen %s\n", g.Gogen)
				fmt.Printf("------------------------------------------------------\n")
				fmt.Printf("%15s : %s\n", "Gogen", g.Gogen)
				fmt.Printf("%15s : %s\n", "Owner", g.Owner)
				fmt.Printf("%15s : %s\n", "Name", g.Name)
				fmt.Printf("%15s : %s\n", "Description", g.Description)
				fmt.Printf("%15s : %s\n", "Gist Link", fmt.Sprintf("https://gist.github.com/%s/%s", g.Owner, g.GistID))
				if len(g.Notes) > 0 {
					fmt.Printf("Notes:\n")
					fmt.Printf("------------------------------------------------------\n")
					fmt.Printf("%s\n", g.Notes)
				}
				var event map[string]interface{}
				var eventbytes []byte
				err = json.Unmarshal([]byte(g.SampleEvent), &event)
				if err != nil {
					eventbytes = []byte(g.SampleEvent)
				} else {
					eventbytes, _ = json.MarshalIndent(event, "", "  ")
				}
				fmt.Printf("Sample Event:\n")
				fmt.Printf("------------------------------------------------------\n")
				fmt.Printf("%s\n", string(eventbytes))
				return nil
			},
		},
		{
			Name:  "push",
			Usage: "Push running config to Gogen sharing service",
			ArgsUsage: "[name]\n\n" + "This will push your running config to the Gogen sharing API.  This will publish the running config\n" +
				"in a Git Gist and make an entry in the Gogen API database pointing to the gist with a bit of metadata.\n\n" +
				"The [name] argument will be the name of the config published.  The entry in the database\n" +
				"will get its Description and Notes from the first sample.  If a mix is specified, it will\n" +
				"attempt to push all referenced configs in the sample.",
			Action: func(clic *cli.Context) error {
				// config.ResetConfig()
				// _ = config.NewConfig()
				if len(clic.Args()) == 0 {
					fmt.Println("Error: Must specify a name to publish this config")
					os.Exit(1)
				}
				var r run.Runner
				owner, id := config.Push(clic.Args().First(), r)
				fmt.Printf("Push successful.  Gist: https://gist.github.com/%s/%s\n", owner, id)
				return nil
			},
		},
		{
			Name:      "pull",
			Usage:     "Pull a config down for editing",
			ArgsUsage: "[owner/name] [directory]",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "deconstruct, d"},
			},
			Action: func(clic *cli.Context) error {
				if len(clic.Args()) == 0 {
					fmt.Println("Error: Must specify a Gogen in owner/name format")
					os.Exit(1)
				} else if len(clic.Args()) < 2 {
					fmt.Println("Error: Must specify a directory to place config files")
					os.Exit(1)
				}
				config.Pull(clic.Args()[0], clic.Args()[1], clic.Bool("deconstruct"))
				return nil
			},
		},
		{
			Name:  "env",
			Usage: "Outputs environment variables based on command line options to pass to eval $(gogen <foo> env)",
			Action: func(clic *cli.Context) error {
				var out string
				for _, flag := range clic.GlobalFlagNames() {
					if clic.GlobalIsSet(flag) {
						out = fmt.Sprintf("%sexport %s=%s\n", out, envVarMap[flag], clic.GlobalString(flag))
					}
				}
				fmt.Printf(out)
				return nil
			},
		},
		{
			Name:  "unsetenv",
			Usage: "Outputs unset commands for environment variabels to clear config",
			Action: func(clic *cli.Context) error {
				var out string
				for _, v := range envVarMap {
					if len(os.Getenv(v)) > 0 {
						out = fmt.Sprintf("%sunset %s\n", out, v)
					}
				}
				fmt.Print(out)
				return nil
			},
		},
		{
			Name:  "version",
			Usage: "Outputs version info",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "versiononly, v"},
			},
			Action: func(clic *cli.Context) error {
				if clic.Bool("versiononly") {
					fmt.Printf("%s", Version)
					return nil
				}
				fmt.Printf("Version: %s\n", Version)
				fmt.Printf("Build Date: %s\n", BuildDate)
				fmt.Printf("Git Summary: %s\n", GitSummary)
				return nil
			},
		},
	}
	app.Before = func(clic *cli.Context) error {
		Setup(clic)
		return nil
	}
	app.Action = func(clic *cli.Context) error {
		clic.App.Command("gen").Run(clic)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "utc, u",
			Usage:  "Outputs time in UTC",
			EnvVar: "GOGEN_UTC",
		},
		cli.BoolFlag{
			Name:   "info, v",
			Usage:  "Sets info level logging",
			EnvVar: "GOGEN_INFO",
		},
		cli.BoolFlag{
			Name:   "debug, vv",
			Usage:  "Sets debug level logging",
			EnvVar: "GOGEN_DEBUG",
		},
		cli.IntFlag{
			Name:   "generators, g",
			Usage:  "Sets number of generator `threads`",
			EnvVar: "GOGEN_GENERATORS",
		},
		cli.IntFlag{
			Name:   "outputters, os",
			Usage:  "Sets number of outputter `threads`",
			EnvVar: "GOGEN_OUTPUTTERS",
		},
		cli.StringFlag{
			Name:   "outputTemplate, ot",
			Usage:  "Use output template (raw|csv|json|splunkhec|splunktcp|splunktcpuf|rfc3134|rfc5424|elasticsearch) for formatting output",
			EnvVar: "GOGEN_OUTPUTTEMPLATE",
		},
		cli.StringFlag{
			Name:   "outputter, o",
			Usage:  "Use outputter (stdout|devnull|file|http|tcp|splunktcp|splunktcpuf) for output",
			EnvVar: "GOGEN_OUT",
		},
		cli.StringFlag{
			Name:   "filename, f",
			Usage:  "Set `filename`, only usable with file output",
			EnvVar: "GOGEN_FILENAME",
		},
		cli.StringFlag{
			Name:   "topic, t",
			Usage:  "Set `topic`, only usable with Kafka output",
			EnvVar: "GOGEN_TOPIC",
		},
		cli.StringFlag{
			Name:   "url",
			Usage:  "Override all endpoint URLs to just `url` url",
			EnvVar: "GOGEN_URL",
		},
		cli.StringFlag{
			Name:   "splunkHECToken",
			Usage:  "Set Authorization: Splunk <token> HTTP header for Splunk's HTTP Event Collector",
			EnvVar: "GOGEN_HEC_TOKEN",
		},
		cli.StringFlag{
			Name:   "configDir, cd",
			Usage:  "Sets `directory` to search for config files, default '$GOGEN_HOME/config'",
			EnvVar: "GOGEN_CONFIG_DIR",
		},
		cli.StringFlag{
			Name:   "samplesDir, sd",
			Usage:  "Sets `directory` to search for sample files, default 'config/samples'",
			EnvVar: "GOGEN_SAMPLES_DIR",
		},
		cli.StringFlag{
			Name:   "tempDir, td",
			Usage:  "Sets `directory` to store temporary files, default $HOME",
			EnvVar: "GOGEN_TMPDIR",
		},
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "`Path` or URL to a full config",
			EnvVar: "GOGEN_CONFIG",
		},
		cli.BoolFlag{
			Name:   "addTime, at",
			Usage:  "Always add _time field, no matter of outputTemplate",
			EnvVar: "GOGEN_ADDTIME",
		},
		cli.IntFlag{
			Name:   "bufferBytes, bb",
			Usage:  "Sets size of output buffers",
			EnvVar: "GOGEN_BUFFERBYTES",
		},
		cli.StringFlag{
			Name:   "logFile, lf",
			Usage:  "Output internal logs to a file instead of stderr",
			EnvVar: "GOGEN_LOGFILE",
		},
		cli.BoolFlag{
			Name:   "logJson, lj",
			Usage:  "Output internal logs as JSON instead of human readable",
			EnvVar: "GOGEN_LOGJSON",
		},
		cli.IntFlag{
			Name:   "cacheIntervals, ci",
			Usage:  "Number of intervals to cache generation",
			EnvVar: "GOGEN_CACHEINTERVALS",
		},
		cli.BoolFlag{
			Name:   "fullRetard, fr",
			Usage:  "Go Full Retard",
			Hidden: true,
		},
	}
	app.Run(os.Args)
}
