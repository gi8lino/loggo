package tui

import (
	"maps"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/gi8lino/loggo/internal/filter"
	"github.com/gi8lino/loggo/internal/ingest"
	"github.com/gi8lino/loggo/internal/logentry"
	logparser "github.com/gi8lino/loggo/internal/parser"
	"github.com/gi8lino/loggo/internal/profile"
)

const (
	frameInterval    = 33 * time.Millisecond
	maxLinesPerFrame = 5000
)

// InitialState contains startup state for the TUI.
type InitialState struct {
	Search        string
	Include       []string
	Exclude       []string
	HiddenFields  []string
	Overrides     profile.RuntimeOverrides
	StopStream    func()
	BufferSize    int
	Debug         bool
	LoadedConfigs []string
	Version       string
	Commit        string
}

// RawLine stores a raw input line before profile parsing.
type RawLine struct {
	Index    int
	Text     string
	Received time.Time
}

// Mode describes the active input mode.
type Mode int

const (
	modeNormal Mode = iota
	modeSearch
	modeFilterField
	modeFilterOperator
	modeFilterValue
	modeExcludeField
	modeExcludeOperator
	modeExcludeValue
	modeColumns
	modeProfile
	modeInspect
	modeHelp
)

// filterOperator describes one guided filter operator.
type filterOperator struct {
	Label      string
	Token      string
	NeedsValue bool
}

// Model stores the complete TUI state.
type Model struct {
	width                int
	height               int
	stream               <-chan ingest.Batch
	profiles             map[string]profile.Profile
	profileNames         []string
	activeProfile        profile.Profile
	activeParser         logparser.Parser
	profileCursor        int
	raw                  []RawLine
	parsed               []logentry.Entry
	visible              []int
	pending              []string
	filterSet            *filter.Set
	search               string
	searchMatcher        filter.Matcher
	include              []string
	exclude              []string
	filterContext        int
	overrides            profile.RuntimeOverrides
	stopStream           func()
	hiddenFields         map[string]struct{}
	showHeaders          bool
	columnFieldOptions   []string
	columnHiddenDraft    map[string]struct{}
	columnFieldCursor    int
	input                string
	mode                 Mode
	selected             int
	vimGotoPending       bool
	follow               bool
	paused               bool
	eof                  bool
	err                  error
	bufferSize           int
	nextIndex            int
	filterFieldCursor    int
	filterOperatorCursor int
	filterField          string
	filterOperator       filterOperator
	filterFieldOptions   []string
	debug                bool
	loadedConfigs        []string
	version              string
	commit               string
}

// NewModel creates a TUI model from profiles and initial state.
func NewModel(
	profiles map[string]profile.Profile,
	active profile.Profile,
	stream <-chan ingest.Batch,
	initial InitialState,
) (Model, error) {
	if initial.BufferSize <= 0 {
		initial.BufferSize = 5000
	}

	activeParser, err := logparser.New(active)
	if err != nil {
		return Model{}, err
	}

	names := slices.Sorted(maps.Keys(profiles))
	profileCursor := 0

	for index, name := range names {
		if name == active.Name {
			profileCursor = index
			break
		}
	}

	model := Model{
		stream:        stream,
		profiles:      profiles,
		profileNames:  names,
		activeProfile: active,
		activeParser:  activeParser,
		profileCursor: profileCursor,
		search:        initial.Search,
		include:       append([]string{}, initial.Include...),
		exclude:       append([]string{}, initial.Exclude...),
		overrides:     initial.Overrides,
		stopStream:    initial.StopStream,
		hiddenFields:  fieldSet(initial.HiddenFields),
		showHeaders:   true,
		follow:        true,
		bufferSize:    initial.BufferSize,
		debug:         initial.Debug,
		loadedConfigs: initial.LoadedConfigs,
		version:       initial.Version,
		commit:        initial.Commit,
	}

	if err := model.rebuildFilter(); err != nil {
		return Model{}, err
	}

	model.setSearchQuery(initial.Search)
	model.rebuildVisible()

	return model, nil
}

// Init starts stream waiting and frame ticking.
func (m Model) Init() tea.Cmd {
	return tea.Batch(waitForBatch(m.stream), tickFrame())
}

// Update updates the TUI state from input, stream, and frame messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case ingest.Batch:
		if len(typed.Lines) > 0 {
			m.pending = append(m.pending, typed.Lines...)
		}
		if typed.Err != nil {
			m.err = typed.Err
		}
		if typed.EOF {
			m.eof = true
			return m, nil
		}

		return m, waitForBatch(m.stream)

	case frameMsg:
		m.processPending(maxLinesPerFrame)

		if !m.eof || len(m.pending) > 0 {
			return m, tickFrame()
		}

		return m, nil

	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height

		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(typed)
	}

	return m, nil
}
