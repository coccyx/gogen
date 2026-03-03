package internal

import (
	"sort"
	"strconv"
	"sync"
	"time"

	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/timeparser"
	lua "github.com/yuin/gopher-lua"
)

// validate takes a sample and checks against any rules which may cause the configuration to be invalid.
// This hopefully centralizes logic for valid configs, disabling any samples which are not valid and
// preventing this logic from sprawling all over the code base.
// Also finds any references from tokens to other samples and
// updates the token to point to the sample data
// Also fixes up any additional things which are needed, like weighted choice string
// string map to the randutil Choice struct
func (c *Config) validate(s *Sample) {
	if !s.realSample {
		return
	}

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

	// Parse earliest and latest as relative times
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

	c.resolveTokenSamples(s)

	if s.EarliestParsed > s.LatestParsed {
		log.Errorf("Earliest time cannot be greater than latest for sample '%s', disabling Sample", s.Name)
		s.Disabled = true
		return
	}
	if s.Interval == 0 && s.Generator != "replay" {
		log.Infof("No interval set for sample '%s', setting endIntervals to 1", s.Name)
		s.EndIntervals = 1
	}

	c.validateTokens(s)
	c.computeSinglePass(s)
	c.setupGenerator(s)
}

// resolveTokenSamples resolves references from tokens to other samples,
// setting up Choice, WeightedChoice, or FieldChoice data on each token.
func (c *Config) resolveTokenSamples(s *Sample) {
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
		s.Tokens[i].Parent = s
		s.Tokens[i].luaState = new(lua.LTable)
		for j := 0; j < len(c.Samples); j++ {
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
}

// validateTokens checks token configurations for validity, disabling the sample if any token is invalid.
func (c *Config) validateTokens(s *Sample) {
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
}

// computeSinglePass checks if SinglePass optimization is feasible for the sample
// by verifying all tokens can be located in each line without overlapping.
func (c *Config) computeSinglePass(s *Sample) {
	if s.Disabled {
		return
	}
	s.SinglePass = true

	var tlines []map[string]tokenspos

outer:
	for _, l := range s.Lines {
		tp := make(map[string]tokenspos)
		for j, t := range s.Tokens {
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
				if lastpos > pos.Pos1 {
					log.Infof("Token '%s' extends beyond beginning of token '%s', disabling SinglePass", lasttoken, s.Tokens[pos.Token].Name)
					s.SinglePass = false
					break outer
				}
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
		// Break up each line and field according to the positions of the tokens
		for i, line := range s.Lines {
			if len(tlines) >= i && len(tlines) > 0 {
				bline := make(map[string][]StringOrToken)
				for field := range line {
					var bfield []StringOrToken
					if _, ok := tlines[i][field]; !ok {
						bf := StringOrToken{T: nil, S: line[field]}
						bfield = append(bfield, bf)
					} else {
						lastpos := 0
						for _, tp := range tlines[i][field] {
							if tp.Pos1 == 0 {
								bf := StringOrToken{T: &s.Tokens[tp.Token], S: ""}
								bfield = append(bfield, bf)
								lastpos = tp.Pos2
							} else {
								bf := StringOrToken{T: nil, S: s.Lines[i][field][lastpos:tp.Pos1]}
								bfield = append(bfield, bf)
								bf = StringOrToken{T: &s.Tokens[tp.Token], S: ""}
								bfield = append(bfield, bf)
								lastpos = tp.Pos2
							}
						}
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

// setupGenerator configures the sample's generator: replay offsets for replay generators,
// or custom Lua generator linkage for non-sample generators.
func (c *Config) setupGenerator(s *Sample) {
	if s.Generator == "replay" {
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
					if i > 0 {
						s.ReplayOffsets[i-1] = lastts.Sub(ts) * -1
						avgOffset = (avgOffset + s.ReplayOffsets[i-1]) / 2
					}
					lastts = ts
					break inner2
				}
			}
			s.ReplayOffsets[len(s.ReplayOffsets)-1] = avgOffset
		}
		log.WithFields(log.Fields{
			"sample":        s.Name,
			"ReplayOffsets": s.ReplayOffsets,
		}).Debugf("ReplayOffsets values")
	} else if s.Generator != "sample" {
		for _, g := range c.Generators {
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

// validateRater returns a copy of the rater with the Options properly cast
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
