package internal

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	strftime "github.com/cactus/gostrftime"
	log "github.com/coccyx/gogen/logger"
	"github.com/pbnjay/strptime"
	uuid "github.com/satori/go.uuid"
	lua "github.com/yuin/gopher-lua"
)

const randStringLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randHexLetters = "ABCDEF0123456789"

// Sample is the main configuration data structure which is passed around through Gogen
// Publicly exported options are brought in through YAML or JSON configs, and some state is maintained in private unexposed variables.
type Sample struct {
	Name            string              `json:"name" yaml:"name"`
	Description     string              `json:"description,omitempty" yaml:"description,omitempty"`
	Notes           string              `json:"notes,omitempty" yaml:"notes,omitempty"`
	Disabled        bool                `json:"disabled" yaml:"disabled"`
	Generator       string              `json:"generator,omitempty" yaml:"generator,omitempty"`
	RaterString     string              `json:"rater,omitempty" yaml:"rater,omitempty"`
	Interval        int                 `json:"interval,omitempty" yaml:"interval,omitempty"`
	Delay           int                 `json:"delay,omitempty" yaml:"delay,omitempty"`
	Count           int                 `json:"count,omitempty" yaml:"count,omitempty"`
	Earliest        string              `json:"earliest,omitempty" yaml:"earliest,omitempty"`
	Latest          string              `json:"latest,omitempty" yaml:"latest,omitempty"`
	Begin           string              `json:"begin,omitempty" yaml:"begin,omitempty"`
	End             string              `json:"end,omitempty" yaml:"end,omitempty"`
	EndIntervals    int                 `json:"endIntervals,omitempty" yaml:"endIntervals,omitempty"`
	RandomizeCount  float32             `json:"randomizeCount,omitempty" yaml:"randomizeCount,omitempty"`
	RandomizeEvents bool                `json:"randomizeEvents,omitempty" yaml:"randomizeEvents,omitempty"`
	Tokens          []Token             `json:"tokens,omitempty" yaml:"tokens,omitempty"`
	Lines           []map[string]string `json:"lines,omitempty" yaml:"lines,omitempty"`
	Field           string              `json:"field,omitempty" yaml:"field,omitempty"`
	FromSample      string              `json:"fromSample,omitempty" yaml:"fromSample,omitempty"`
	SinglePass      bool                `json:"singlepass,omitempty" yaml:"singlepass,omitempty"`

	// Internal use variables
	Rater           Rater                        `json:"-" yaml:"-"`
	Output          *Output                      `json:"-" yaml:"-"`
	EarliestParsed  time.Duration                `json:"-" yaml:"-"`
	LatestParsed    time.Duration                `json:"-" yaml:"-"`
	BeginParsed     time.Time                    `json:"-" yaml:"-"`
	EndParsed       time.Time                    `json:"-" yaml:"-"`
	Current         time.Time                    `json:"-" yaml:"-"` // If we are backfilling or generating for a specified time window, what time is it?
	Realtime        bool                         `json:"-" yaml:"-"` // Are we done doing batch backfill or specified time window?
	BrokenLines     []map[string][]StringOrToken `json:"-" yaml:"-"`
	ReplayOffsets   []time.Duration              `json:"-" yaml:"-"`
	CustomGenerator *GeneratorConfig             `json:"-" yaml:"-"`
	GeneratorState  *GeneratorState              `json:"-" yaml:"-"`
	LuaMutex        *sync.Mutex                  `json:"-" yaml:"-"`
	Buf             *bytes.Buffer                `json:"-" yaml:"-"`
	realSample      bool                         // Used to represent samples which aren't just used to store lines from CSV or raw
}

// Clock allows for implementers to keep track of their own view
// of current time.  In Gogen, this is used for being able to generate
// events between certain time windows, or backfill from a certain time
// while continuing to run in real time.
type Clock interface {
	Now() time.Time
}

// Now returns the current time for the sample, and handles
// whether we are currently generating a backfill or
// specified time window or whether we should be generating
// events in realtime
func (s *Sample) Now() time.Time {
	if !s.Realtime {
		return s.Current
	}
	return time.Now()
}

// Token describes a replacement task to run against a sample
type Token struct {
	Name           string              `json:"name" yaml:"name"`
	Format         string              `json:"format" yaml:"format"`
	Token          string              `json:"token" yaml:"token"`
	Type           string              `json:"type" yaml:"type"`
	Replacement    string              `json:"replacement,omitempty" yaml:"replacement,omitempty"`
	Group          int                 `json:"group,omitempty" yaml:"group,omitempty"`
	Sample         *Sample             `json:"-" yaml:"-"`
	Parent         *Sample             `json:"-" yaml:"-"`
	SampleString   string              `json:"sample,omitempty" yaml:"sample,omitempty"`
	Field          string              `json:"field,omitempty" yaml:"field,omitempty"`
	SrcField       string              `json:"srcField,omitempty" yaml:"srcField,omitempty"`
	Precision      int                 `json:"precision,omitempty" yaml:"precision,omitempty"`
	Lower          int                 `json:"lower,omitempty" yaml:"lower,omitempty"`
	Upper          int                 `json:"upper,omitempty" yaml:"upper,omitempty"`
	Length         int                 `json:"length,omitempty" yaml:"length,omitempty"`
	WeightedChoice []WeightedChoice    `json:"weightedChoice,omitempty" yaml:"weightedChoice,omitempty"`
	FieldChoice    []map[string]string `json:"fieldChoice,omitempty" yaml:"fieldChoice,omitempty"`
	Choice         []string            `json:"choice,omitempty" yaml:"choice,omitempty"`
	Script         string              `json:"script,omitempty" yaml:"script,omitempty"`
	Init           map[string]string   `json:"init,omitempty" yaml:"init,omitempty"`
	RaterString    string              `json:"rater,omitempty" yaml:"rater,omitempty"`
	Disabled       bool                `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Rater          Rater               `json:"-" yaml:"-"`

	L                          *lua.LState `json:"-" yaml:"-"`
	luaState                   *lua.LTable
	mutex                      *sync.Mutex
	weightedChoiceTotals       []int
	weightedChoiceRunningTotal int
}

// WeightedChoice is a simple data structure for allowing a list of items with a Choice to pick and a Weight for that choice
type WeightedChoice struct {
	Weight int    `json:"weight" yaml:"weight"`
	Choice string `json:"choice" yaml:"choice"`
}

type tokenpos struct {
	Pos1  int
	Pos2  int
	Token int
}

type tokenspos []tokenpos

func (tp tokenspos) Len() int           { return len(tp) }
func (tp tokenspos) Less(i, j int) bool { return tp[i].Pos1 < tp[j].Pos2 }
func (tp tokenspos) Swap(i, j int)      { tp[i], tp[j] = tp[j], tp[i] }

// StringOrToken is used for SinglePass and stores either a string or a token
type StringOrToken struct {
	S string
	T *Token
}

// Replace replaces any instances of this token in the string pointed to by event.  Since time is native is Gogen, we can pass in
// earliest and latest time ranges to generate the event between.  Lastly, some times we want to span a selected choice over multiple
// tokens.  Passing in a pointer to choice allows the replacement to choose a preselected row in FieldChoice or Choice.
func (t Token) Replace(event *string, choice int, et time.Time, lt time.Time, now time.Time, randgen *rand.Rand) (int, error) {
	// s := t.Sample
	e := *event

	if offsets, err := t.GetReplacementOffsets(*event); err != nil {
		return choice, nil
	} else {
		retchoice := choice
		lastoffset := 0
		*event = ""
		for _, match := range offsets {
			replacement, newchoice, err := t.GenReplacement(retchoice, et, lt, now, randgen)
			if err != nil {
				return -1, err
			}
			*event = *event + e[lastoffset:match[0]] + replacement
			retchoice = newchoice
			lastoffset = match[1]
		}
		*event += e[lastoffset:]
		return retchoice, nil
	}
}

// GetReplacementOffsets returns the beginning and end of a token inside an event string
func (t Token) GetReplacementOffsets(event string) ([][]int, error) {
	ret := make([][]int, 0)
	switch t.Format {
	case "template":
		offset := 0
		for {
			pos := strings.Index(event[offset:], t.Token)
			if pos < 0 {
				break
			}
			ret = append(ret, []int{offset + pos, offset + pos + len(t.Token)})
			offset += pos + len(t.Token)
		}
	case "regex":
		re, err := regexp.Compile(t.Token)
		if err != nil {
			return ret, err
		}
		matches := re.FindAllStringSubmatchIndex(event, -1)
		if matches != nil {
			for _, match := range matches {
				if len(match) >= 4 {
					ret = append(ret, []int{match[2], match[3]})
				}
			}
		}
	}
	if len(ret) > 0 {
		return ret, nil
	}
	return ret, fmt.Errorf("Token '%s' not found in field '%s': '%s'", t.Token, t.Field, event)
}

// GenReplacement generates a replacement value for the token.  choice allows the user to specify
// a specific value to choose in the array.  This is useful for saving picks amongst tokens.
func (t Token) GenReplacement(choice int, et time.Time, lt time.Time, now time.Time, randgen *rand.Rand) (string, int, error) {
	switch t.Type {
	case "timestamp", "gotimestamp", "epochtimestamp":
		td := lt.Sub(et)

		var tdr int
		if int(td) > 0 {
			tdr = randgen.Intn(int(td))
		}
		rd := time.Duration(tdr)
		replacementTime := lt.Add(rd * -1)
		switch t.Type {
		case "timestamp":
			return strftime.Format(t.Replacement, replacementTime), -1, nil
		case "gotimestamp":
			return replacementTime.Format(t.Replacement), -1, nil
		case "epochtimestamp":
			return strconv.FormatInt(replacementTime.Unix(), 10), -1, nil
		}
	case "static":
		return t.Replacement, -1, nil
	case "rated":
		switch t.Replacement {
		case "int":
			var ret int
			if (t.Upper - t.Lower) > 0 {
				ret = randgen.Intn(t.Upper-t.Lower) + t.Lower
			} else if (t.Upper - t.Lower) <= 0 {
				ret = t.Upper
			}
			rate := t.Rater.TokenRate(t, now)
			rated := float64(ret) * rate
			if rated < 0 {
				ret = int(rated - 0.5)
			} else {
				ret = int(rated + 0.5)
			}
			return strconv.Itoa(ret), -1, nil
		case "float":
			lower := t.Lower * int(math.Pow10(t.Precision))
			upper := t.Upper * int(math.Pow10(t.Precision))
			var f float64
			if (upper - lower) > 0 {
				f = float64(randgen.Intn(upper-lower)+lower) / math.Pow10(t.Precision)
			} else {
				f = float64(upper) / math.Pow10(t.Precision)
			}
			rate := t.Rater.TokenRate(t, now)
			f = f * rate
			return strconv.FormatFloat(f, 'f', t.Precision, 64), -1, nil
		}
	case "random":
		switch t.Replacement {
		case "int":
			ri := randgen.Intn(t.Upper-t.Lower) + t.Lower
			return strconv.Itoa(ri), -1, nil
		case "float":
			lower := t.Lower * int(math.Pow10(t.Precision))
			upper := t.Upper * int(math.Pow10(t.Precision))
			f := float64(randgen.Intn(upper-lower)+lower) / math.Pow10(t.Precision)
			return strconv.FormatFloat(f, 'f', t.Precision, 64), -1, nil
		case "string", "hex":
			var ret string
			for i := 0; i < t.Length; i++ {
				if t.Replacement == "string" {
					ri := randgen.Intn(len(randStringLetters))
					ret += randStringLetters[ri : ri+1]
				} else {
					ri := randgen.Intn(len(randHexLetters))
					ret += randHexLetters[ri : ri+1]
				}
			}
			return ret, -1, nil
		case "guid":
			u := uuid.NewV4()
			return u.String(), -1, nil
		case "ipv4":
			var ret string
			for i := 0; i < 4; i++ {
				ri := randgen.Intn(255)
				ret += strconv.Itoa(ri)
				if i < 3 {
					ret += "."
				}
			}
			return ret, -1, nil
		case "ipv6":
			var ret string
			for i := 0; i < 8; i++ {
				ri := randgen.Intn(65535)
				ret += fmt.Sprintf("%x", ri)
				if i < 7 {
					ret += ":"
				}
			}
			return ret, -1, nil
		}
	case "choice":
		if choice == -1 {
			choice = randgen.Intn(len(t.Choice))
		} else if choice < 0 || choice > len(t.Choice) {
			return "", -1, fmt.Errorf("Choice out of range")
		}
		return t.Choice[choice], choice, nil
	case "weightedChoice":
		if choice == -1 {
			// From http://eli.thegreenplace.net/2010/01/22/weighted-random-generation-in-python/
			if t.weightedChoiceTotals == nil {
				t.weightedChoiceTotals = make([]int, len(t.WeightedChoice))
				for i, w := range t.WeightedChoice {
					t.weightedChoiceRunningTotal += w.Weight
					t.weightedChoiceTotals[i] = t.weightedChoiceRunningTotal
				}
			}

			r := randgen.Float64() * float64(t.weightedChoiceRunningTotal)
			for j, total := range t.weightedChoiceTotals {
				if r < float64(total) {
					choice = j
					break
				}
			}
		} else if choice < 0 || choice > len(t.WeightedChoice) {
			return "", -1, fmt.Errorf("Choice out of range")
		}
		return t.WeightedChoice[choice].Choice, choice, nil
	case "fieldChoice":
		if choice == -1 {
			choice = randgen.Intn(len(t.FieldChoice))
		} else if choice < 0 || choice > len(t.FieldChoice) {
			return "", -1, fmt.Errorf("Choice out of range")
		}
		return t.FieldChoice[choice][t.SrcField], choice, nil
	case "script":
		t.mutex.Lock()
		defer t.mutex.Unlock()
		L := lua.NewState()
		defer L.Close()
		L.SetGlobal("state", t.luaState)
		if err := L.DoString(t.Script); err != nil {
			log.Errorf("Error executing script for token '%s' in sample '%s': %s", t.Name, t.Parent.Name, err)
		}
		return lua.LVAsString(L.Get(-1)), -1, nil
	}
	return "", -1, fmt.Errorf("GenReplacement called with invalid type for token '%s' with type '%s'", t.Name, t.Type)
}

// ParseTimestamp will return a time.Time based on the configured token's setup
func (t Token) ParseTimestamp(eventts string) (time.Time, error) {
	switch t.Type {
	case "timestamp":
		ts, err := strptime.Parse(eventts, t.Replacement)
		if err != nil {
			return time.Time{}, err
		}
		return ts, nil
	case "gotimestamp":
		ts, err := time.Parse(t.Replacement, eventts)
		if err != nil {
			return time.Time{}, err
		}
		return ts, nil
	case "epochtimestamp":
		tsi, err := strconv.ParseInt(eventts, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		ts := time.Unix(tsi, 0)
		return ts, nil
	default:
		return time.Time{}, fmt.Errorf("Token not a timestamp token")
	}
}
