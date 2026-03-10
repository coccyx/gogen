package internal

import (
	"fmt"
	"os"

	log "github.com/coccyx/gogen/logger"
)

// SetupSystemTokens adds tokens like time and facility to samples based on configuration
func (c *Config) SetupSystemTokens() {
	addToken := func(s *Sample, tokenName string, tokenType string, tokenReplacement string) {
		// If there's no _time token, add it to make sure we have a timestamp field in every event
		tokenfound := false
		for _, t := range s.Tokens {
			if t.Name == tokenName {
				tokenfound = true
				break
			}
		}
		if !tokenfound {
			for _, l := range s.Lines {
				if _, ok := l[tokenName]; ok {
					tokenfound = true
					break
				}
			}
		}
		if !tokenfound {
			log.Infof("Adding %s token for sample %s", tokenName, s.Name)
			tt := Token{
				Name:   tokenName,
				Type:   tokenType,
				Format: "template",
				Field:  tokenName,
				Token:  fmt.Sprintf("$%s$", tokenName),
				Group:  -1,
				Parent: s,
			}
			if tokenReplacement != "" {
				tt.Replacement = tokenReplacement
			}
			s.Tokens = append(s.Tokens, tt)
			if s.SinglePass {
				for j := 0; j < len(s.BrokenLines); j++ {
					st := []StringOrToken{
						{T: &tt, S: ""},
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
		log.Infof("Adding %s field for sample %s", name, s.Name)
		for i := 0; i < len(s.Lines); i++ {
			if s.Lines[i][name] == "" {
				s.Lines[i][name] = value
			}
		}
		if s.SinglePass {
			for i := 0; i < len(s.BrokenLines); i++ {
				st := []StringOrToken{
					{T: nil, S: value},
				}
				if _, ok := s.BrokenLines[i][name]; !ok {
					s.BrokenLines[i][name] = st
				}
			}
		}
	}
	syslogOutput := c.Global.Output.OutputTemplate == "rfc3164" || c.Global.Output.OutputTemplate == "rfc5424"
	addTime := c.Global.Output.OutputTemplate == "splunkhec" ||
		c.Global.Output.OutputTemplate == "elasticsearch" ||
		c.Global.AddTime ||
		syslogOutput
	if !c.cc.Export && addTime {
		// Use epochtimestamp for Splunk, or different formats for rfc3164 or rfc5424
		var tokenType string
		var tokenReplacement string
		tokenName := "_time"
		if c.Global.Output.OutputTemplate == "elasticsearch" {
			tokenName = "@timestamp"
			tokenType = "gotimestamp"
			tokenReplacement = "2006-01-02T15:04:05.999Z07:00"
		} else if !syslogOutput {
			tokenType = "epochtimestamp"
		} else if c.Global.Output.OutputTemplate == "rfc3164" {
			tokenType = "gotimestamp"
			tokenReplacement = "Jan _2 15:04:05"
		} else if c.Global.Output.OutputTemplate == "rfc5424" {
			tokenType = "gotimestamp"
			tokenReplacement = "2006-01-02T15:04:05.999999Z07:00"
		}
		hostname, _ := os.Hostname()
		for i := 0; i < len(c.Samples); i++ {
			s := c.Samples[i]
			addToken(s, tokenName, tokenType, tokenReplacement) // Timestamp
			// Add fields for syslog output
			if syslogOutput {
				addField(s, "priority", fmt.Sprintf("%d", defaultSyslogPriority))
				addField(s, "host", hostname)
				tag := "gogen"
				if len(s.Lines) > 0 && s.Lines[0]["sourcetype"] != "" {
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
