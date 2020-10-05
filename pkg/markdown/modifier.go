package markdown

import (
	"fmt"
	"regexp"
	"strings"
)

type Flag string
type Parameter string

type Modifiers struct {
	fmt.Stringer
	Values map[Parameter]string
	Flags  map[Flag]bool
}

func NewModifiers() *Modifiers {
	return &Modifiers{
		Values: map[Parameter]string{},
		Flags:  map[Flag]bool{},
	}
}

func (m *Modifiers) Ingest(from *Modifiers) {
	for k, v := range from.Flags {
		m.Flags[k] = v
	}

	for k, v := range from.Values {
		m.Values[k] = v
	}
}

func (m *Modifiers) String() string {
	keys := make([]Flag, len(m.Flags))

	i := 0
	for k := range m.Flags {
		keys[i] = k
		i++
	}

	return fmt.Sprintf("Flags: %s, Values: %v", keys, m.Values)
}

func (m *Modifiers) HasAny(names ...string) bool {
	for _, name := range names {
		if _, ok := m.Flags[Flag(name)]; ok {
			return true
		}

		if _, ok := m.Values[Parameter(name)]; ok {
			return true
		}
	}

	return false
}

func (m *Modifiers) HasAll(names ...string) bool {
	for _, name := range names {
		if _, ok := m.Flags[Flag(name)]; ok {
			continue
		}

		if _, ok := m.Values[Parameter(name)]; ok {
			continue
		}

		return false
	}

	return true
}

func (m *Modifiers) GetValue(name Parameter) *string {
	if val, ok := m.Values[name]; ok {
		return &val
	}
	return nil
}

func (m *Modifiers) GetFlag(name Flag) *bool {
	if val, ok := m.Flags[name]; ok {
		return &val
	}
	return nil
}

var nameMatch = `[a-z_\-0-9]+`
var kvMatch = `[\=\:]`
var modifierDetect = regexp.MustCompile(`(` + nameMatch + kvMatch + `\".*?\")\s*|(` + nameMatch + kvMatch + `'.*?')\s*|(` + nameMatch + kvMatch + `[^\s]+)\s*|(` + nameMatch + `)\s*`)

const (
	quotedKV       = 1
	quotedSingleKV = 2
	unquotedKV     = 3
	flagV          = 4
)

func ParseModifiers(line string, kvSep string) *Modifiers {

	result := NewModifiers()

	for _, match := range modifierDetect.FindAllStringSubmatch(line, -1) {
		if subject := match[quotedKV]; subject != "" {
			if kv := strings.SplitN(subject, kvSep, 2); len(kv) == 2 {
				result.Values[Parameter(kv[0])] = strings.Trim(kv[1], "\"")
			}
		} else if subject := match[quotedSingleKV]; subject != "" {
			if kv := strings.SplitN(subject, kvSep, 2); len(kv) == 2 {
				result.Values[Parameter(kv[0])] = strings.Trim(kv[1], "'")
			}
		} else if subject := match[unquotedKV]; subject != "" {
			if kv := strings.SplitN(subject, kvSep, 2); len(kv) == 2 {
				result.Values[Parameter(kv[0])] = kv[1]
			}
		} else if subject := match[flagV]; subject != "" {
			result.Flags[Flag(subject)] = true
		}
	}

	return result
}

func MergeModifiers(left, right *Modifiers) *Modifiers {
	result := NewModifiers()
	result.Ingest(left)
	result.Ingest(right)
	return result
}
