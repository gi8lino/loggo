package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/containeroo/tinyflags"
	"github.com/gi8lino/loggo/internal/config"
	"github.com/gi8lino/loggo/internal/flags"
	"github.com/gi8lino/loggo/internal/ingest"
	"github.com/gi8lino/loggo/internal/profile"
	"github.com/gi8lino/loggo/internal/terminal"
	"github.com/gi8lino/loggo/internal/tui"
)

// EnvLookup resolves environment variable values.
type EnvLookup func(string) string

// Run is the single entry point for the application.
func Run(
	ctx context.Context,
	version string,
	commit string,
	args []string,
	stdIn *os.File,
	stdOut io.Writer,
	stdErr io.Writer,
	getEnv EnvLookup,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	parsed, err := flags.ParseFlags(args, version)
	if err != nil {
		if tinyflags.IsHelpRequested(err) || tinyflags.IsVersionRequested(err) {
			fmt.Fprint(stdOut, err.Error()) // nolint:errcheck
			return nil
		}

		fmt.Fprintf(stdErr, "loggo: failed to parse flags: %v\n", err) // nolint:errcheck
		return err
	}

	cfg, loadedConfigs, err := config.Load(parsed.ConfigPaths, config.EnvLookup(getEnv))
	if err != nil {
		fmt.Fprintf(stdErr, "loggo: failed to load config: %v\n", err) // nolint:errcheck
		return err
	}

	activeProfile, err := cfg.ResolveProfile(parsed.Profile, config.EnvLookup(getEnv))
	if err != nil {
		fmt.Fprintf(stdErr, "loggo: failed to resolve profile: %v\n", err) // nolint:errcheck
		return err
	}

	overrides := profile.RuntimeOverrides{
		Parser:       parsed.Parser,
		Split:        parsed.Split,
		Fields:       parsed.Fields,
		Format:       parsed.Format,
		HiddenFields: parsed.HiddenFields,
	}

	activeProfile = activeProfile.WithRuntimeOverrides(overrides)

	terminalSession, err := terminal.Open(stdIn, stdErr)
	if err != nil {
		fmt.Fprintf(stdErr, "loggo: failed to open terminal: %v\n", err) // nolint:errcheck
		return err
	}
	defer terminalSession.Close()

	ingestCtx, cancelIngest := context.WithCancel(ctx)
	defer cancelIngest()

	stream := ingest.Start(ingestCtx, stdIn, ingest.Options{
		BatchSize:     parsed.BatchSize,
		FlushInterval: parsed.FlushInterval,
		JoinMultiline: true,
		MaxEventLines: 256,
	})

	model, err := tui.NewModel(cfg.Profiles, activeProfile, stream, tui.InitialState{
		Search:        parsed.Search,
		Include:       parsed.Filters,
		Exclude:       parsed.Excludes,
		ExportProfile: config.ExportLocalProfile,
		HiddenFields:  activeProfile.HiddenFields,
		Overrides:     overrides,
		StopStream:    cancelIngest,
		BufferSize:    parsed.BufferSize,
		Debug:         parsed.Debug,
		LoadedConfigs: loadedConfigs,
		Version:       version,
		Commit:        commit,
	})
	if err != nil {
		fmt.Fprintf(stdErr, "loggo: failed to initialize tui: %v\n", err) // nolint:errcheck
		return err
	}

	if parsed.Debug {
		fmt.Fprintf(stdErr, "loggo: version=%s commit=%s profile=%s\n", version, commit, activeProfile.Name) // nolint:errcheck
		if len(loadedConfigs) > 0 {
			fmt.Fprintf(stdErr, "loggo: loaded configs=%v\n", loadedConfigs) // nolint:errcheck
		}
		if len(parsed.OverriddenValues) > 0 {
			fmt.Fprintf(stdErr, "loggo: env overrides=%v\n", parsed.OverriddenValues) // nolint:errcheck
		}
	}

	program := tea.NewProgram(
		model,
		tea.WithInput(terminalSession.Input),
		tea.WithOutput(terminalSession.Output),
	)

	_, err = program.Run()

	return err
}
