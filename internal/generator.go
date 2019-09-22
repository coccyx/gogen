package internal

import (
	"math/rand"
	"strconv"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// GeneratorConfig holds our configuration for custom generators
type GeneratorConfig struct {
	Name           string                 `json:"name" yaml:"name"`
	Init           map[string]string      `json:"init,omitempty" yaml:"init,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
	Script         string                 `json:"script" yaml:"script"`
	FileName       string                 `json:"fileName,omitempty" yaml:"fileName,omitempty"`
	SingleThreaded bool                   `json:"singleThreaded,omitempty" yaml:"singleThreaded,omitempty"`
}

// GenQueueItem represents one generation job
type GenQueueItem struct {
	S        *Sample
	Count    int
	Event    int
	Earliest time.Time
	Latest   time.Time
	Now      time.Time
	OQ       chan *OutQueueItem
	Rand     *rand.Rand
	Cache    *CacheItem
}

// Generator will generate count events from earliest to latest time and put them
// in the output queue
type Generator interface {
	Gen(item *GenQueueItem) error
}

// GeneratorState maintains what a custom generator needs to store
type GeneratorState struct {
	LuaState *lua.LTable
	LuaLines *lua.LTable
}

// NewGeneratorState generates a GeneratorState object
func NewGeneratorState(s *Sample) *GeneratorState {
	gs := new(GeneratorState)

	gs.LuaState = new(lua.LTable)
	for k, v := range s.CustomGenerator.Init {
		vAsNum, err := strconv.ParseFloat(v, 64)
		if err == nil {
			gs.LuaState.RawSet(lua.LString(k), lua.LNumber(vAsNum))
		} else {
			gs.LuaState.RawSet(lua.LString(k), lua.LString(v))
		}
	}
	gs.LuaLines = new(lua.LTable)
	for _, line := range s.Lines {
		lualine := new(lua.LTable)
		for k, v := range line {
			lualine.RawSetString(k, lua.LString(v))
		}
		gs.LuaLines.Append(lualine)
	}
	return gs
}
