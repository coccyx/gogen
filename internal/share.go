package internal

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	log "github.com/coccyx/gogen/logger"
	"github.com/google/go-github/github"
	yaml "gopkg.in/yaml.v2"
)

// Run is an interface to run the generator
type Run interface {
	Once(sample string)
}

// Push pushes the running config to the Gogen API and creates a GitHub gist.  Returns the owner and ID of the Gist.
func Push(name string, run Run) (string, string) {
	c := NewConfig()
	ec := BuildConfig(ConfigConfig{
		FullConfig: c.cc.FullConfig,
		ConfigDir:  c.cc.ConfigDir,
		Home:       c.cc.Home,
		SamplesDir: c.cc.SamplesDir,
		GlobalFile: c.cc.GlobalFile,
		Export:     true,
	})
	if len(c.Samples) > 0 {
		// Push all file based mixes
		for i := range ec.Mix {
			m := ec.Mix[i]
			acceptableExtensions := map[string]bool{".yml": true, ".yaml": true, ".json": true}
			if _, ok := acceptableExtensions[filepath.Ext(m.Sample)]; ok {
				sc := BuildConfig(ConfigConfig{
					FullConfig: m.Sample,
					Export:     true,
				})
				login, _ := push(sc.Samples[0].Name, sc, sc, run)
				ec.Mix[i].Sample = login + "/" + sc.Samples[0].Name
			}
		}
		return push(name, c, ec, run)
	}
	log.Panicf("No samples configured")
	return "", ""
}

func push(name string, genc *Config, pushc *Config, run Run) (string, string) {
	log.Debugf("Pushing config as '%s'", name)
	gh := NewGitHub(true)
	gu, _, err := gh.client.Users.Get("")
	if err != nil {
		log.Fatalf("Error getting user in push: %s", err)
	}
	if len(genc.Samples) > 0 {
		sample := genc.Samples[0]
		gogen := *gu.Login + "/" + name

		run.Once(sample.Name)
		log.Debugf("Buf: %s", genc.Buf.String())

		if sample == nil {
			fmt.Printf("Sample '%s' not found\n", name)
			os.Exit(1)
		}

		oldGogen, err := Get(gogen)
		var version int
		if err != nil {
			version = 0
		} else {
			version = oldGogen.Version + 1
		}
		gi := gh.Push(name, pushc)

		g := GogenInfo{
			Gogen:       gogen,
			Name:        name,
			Description: sample.Description,
			Notes:       sample.Notes,
			Owner:       *gu.Login,
			SampleEvent: genc.Buf.String(),
			GistID:      *gi.ID,
			Version:     version,
		}
		Upsert(g)

		return *gu.Login, *gi.ID
	}
	return "", ""
}

// Pull grabs a config from the Gogen API + GitHub gist and creates it on the filesystem for editing
func Pull(gogen string, dir string, deconstruct bool) {
	gogentokens := strings.Split(gogen, "/")
	var name string
	if len(gogentokens) > 1 {
		name = gogentokens[1]
	} else {
		name = gogen
	}
	g, err := Get(gogen)
	if err != nil {
		log.WithError(err).Fatalf("error retrieving gogen config for gogen '%s'", gogen)
	}
	gist := getGist(g)
	for _, file := range gist.Files {
		filename := filepath.Join(dir, *file.Filename)
		client := &http.Client{}
		resp, err := client.Get(*file.RawURL)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Couldn't open file %s: %s", filename, err)
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Fatalf("Error writing to file %s: %s", filename, err)
		}
		f.Close()
		if deconstruct {
			samplesDir := filepath.Join(dir, "samples")
			templatesDir := filepath.Join(dir, "templates")
			generatorsDir := filepath.Join(dir, "generators")
			err := os.Mkdir(samplesDir, 0755)
			err = os.Mkdir(templatesDir, 0755)
			err = os.Mkdir(generatorsDir, 0755)
			if err != nil && !os.IsExist(err) {
				log.Fatalf("Error creating directories %s or %s", samplesDir, templatesDir)
			}

			cc := ConfigConfig{FullConfig: filename, Export: true}
			c := BuildConfig(cc)
			for x := 0; x < len(c.Samples); x++ {
				s := c.Samples[x]
				for y := 0; y < len(s.Tokens); y++ {
					t := c.Samples[x].Tokens[y]
					if t.SampleString != "" {
						fname := t.SampleString
						if fname[len(fname)-6:] == "sample" {
							f, err := os.OpenFile(filepath.Join(samplesDir, fname), os.O_WRONLY|os.O_CREATE, 0644)
							if err != nil {
								log.Fatalf("Unable to open file %s: %s", filepath.Join(samplesDir, fname), err)
							}
							defer f.Close()
							for _, v := range t.Choice {
								_, err := fmt.Fprintf(f, "%s\n", v)
								if err != nil {
									log.Fatalf("Error writing to file %s: %s", filepath.Join(samplesDir, fname), err)
								}
							}
							c.Samples[x].Tokens[y].Choice = []string{}
						} else if fname[len(fname)-3:] == "csv" {
							if len(s.Lines) > 0 {
								f, err := os.OpenFile(filepath.Join(samplesDir, fname), os.O_WRONLY|os.O_CREATE, 0644)
								if err != nil {
									log.Fatalf("Unable to open file %s: %s", filepath.Join(samplesDir, fname), err)
								}
								defer f.Close()
								w := csv.NewWriter(f)

								keys := make([]string, len(t.FieldChoice[0]))
								i := 0
								for k := range t.FieldChoice[0] {
									keys[i] = k
									i++
								}
								sort.Strings(keys)
								w.Write(keys)

								for _, l := range t.FieldChoice {
									values := make([]string, len(keys))
									for j, k := range keys {
										values[j] = l[k]
									}
									w.Write(values)
								}

								w.Flush()
								c.Samples[x].Tokens[y].FieldChoice = []map[string]string{}
							}
						}

						var outb []byte
						var err error
						if outb, err = yaml.Marshal(s); err != nil {
							log.Fatalf("Cannot Marshal sample '%s', err: %s", s.Name, err)
						}
						outfname := filepath.Join(samplesDir, name+".yml")
						log.Debugf("Writing sample file for sammple '%s' at file: %s", s.Name, outfname)
						err = ioutil.WriteFile(outfname, outb, 0644)
						if err != nil {
							log.Fatalf("Cannot write file %s: %s", outfname, err)
						}
					}
				}
			}

			for _, t := range c.Templates {
				var outb []byte
				var err error
				if outb, err = yaml.Marshal(t); err != nil {
					log.Fatalf("Cannot Marshal template '%s', err: %s", t.Name, err)
				}
				err = ioutil.WriteFile(filepath.Join(templatesDir, t.Name+".yml"), outb, 0644)
				if err != nil {
					log.Fatalf("Error writing file %s", filepath.Join(templatesDir, t.Name+".yml"))
				}
			}

			for i, g := range c.Generators {
				if g.FileName != "" {
					fname := filepath.Base(g.FileName)
					err = ioutil.WriteFile(filepath.Join(generatorsDir, fname), []byte(g.Script), 0644)
					if err != nil {
						log.Fatalf("Error writing file %s", filepath.Join(generatorsDir, fname))
					}
					c.Generators[i].FileName = fname
					c.Generators[i].Script = ""
				}

				var outb []byte
				var err error
				if outb, err = yaml.Marshal(g); err != nil {
					log.Fatalf("Cannot Marshal generator '%s', err: %s", g.Name, err)
				}
				err = ioutil.WriteFile(filepath.Join(generatorsDir, g.Name+".yml"), outb, 0644)
				if err != nil {
					log.Fatalf("Error writing file %s", filepath.Join(generatorsDir, g.Name+".yml"))
				}
			}

			for _, g := range c.Mix {
				Pull(g.Sample, dir, true)
			}

			err = os.Remove(filename)
			if err != nil {
				log.Debugf("Error removing original config file during deconstruction: %s", filename)
			}
		}
		break
	}
}

// PullFile pulls a config from the Gogen API + GitHub gist and writes it to a single file
func PullFile(gogen string, filename string) {
	g, err := Get(gogen)
	if err != nil {
		log.WithError(err).Fatalf("error retrieving gogen config for gogen '%s'", gogen)
	}
	var version int
	cached := false

	var readFrom io.ReadCloser
	cacheFile := filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".configcache_"+url.QueryEscape(gogen))
	versionCacheFile := filepath.Join(os.ExpandEnv("$GOGEN_TMPDIR"), ".versioncache_"+url.QueryEscape(gogen))
	_, err = os.Stat(versionCacheFile)
	if err == nil {
		versionBytes, err := ioutil.ReadFile(versionCacheFile)
		if err != nil {
			log.Fatalf("Error reading version cache file '%s': %s", versionCacheFile, err)
		}
		version, err = strconv.Atoi(string(versionBytes))
		if err != nil {
			log.Fatalf("Error converting value in version cache file '%s' to integer: %s", versionCacheFile, err)
		}
		if version == g.Version {
			log.Debugf("Reading config from cache file '%s'", cacheFile)
			readFrom, err = os.Open(cacheFile)
			if err != nil {
				cached = false
			} else {
				cached = true
			}
		} else {
			log.Debugf("Version mismatch, Gogen version %d cached version %d", g.Version, version)
		}
	}
	if !cached {
		gist := getGist(g)
		for _, file := range gist.Files {
			log.Debugf("Reading config from GitHub")
			client := &http.Client{}
			resp, err := client.Get(*file.RawURL)
			if err != nil {
				log.Fatalf("Could not read from HTTP url '%s' for gist '%s': %s", *file.RawURL, gogen, err)
			}
			readFrom = resp.Body
			break
		}
	}
	// Make a copy of readFrom in case we need to write it to cache
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(readFrom); err != nil {
		log.Fatalf("Couldn't read from readFrom into buffer: %s", err)
	}
	if err = readFrom.Close(); err != nil {
		log.Fatalf("Error closing readFrom: %s", err)
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		log.Fatalf("Couldn't open file %s: %s", filename, err)
	}
	_, err = io.Copy(f, bytes.NewReader(buf.Bytes()))
	if err != nil {
		log.Fatalf("Error writing to file %s: %s", filename, err)
	}

	if !cached {
		os.Remove(versionCacheFile)
		os.Remove(cacheFile)
		versioncachef, err := os.OpenFile(versionCacheFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Couldn't open version cache file '%s': %s", versionCacheFile, err)
		}
		defer versioncachef.Close()
		_, err = versioncachef.WriteString(strconv.Itoa(g.Version))
		if err != nil {
			log.Fatalf("Error writing to version cache file: '%s': %s", versionCacheFile, err)
		}
		cachef, err := os.OpenFile(cacheFile, os.O_WRONLY|os.O_CREATE, 0644)
		defer cachef.Close()
		if err != nil {
			log.Fatalf("Couldn't open cache file '%s': %s", cacheFile, err)
		}
		_, err = io.Copy(cachef, bytes.NewReader(buf.Bytes()))
		if err != nil {
			log.Fatalf("Error writing to cache file '%s': %s", cacheFile, err)
		}
	}
}

func getGist(g GogenInfo) (gist *github.Gist) {
	gh := NewGitHub(false)
	gist, _, err := gh.client.Gists.Get(g.GistID)
	if err != nil {
		log.Fatalf("Couldn't get gist: %s", err)
	}
	return gist
}
