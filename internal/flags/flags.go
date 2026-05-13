package flags

import (
	"fmt"
	"time"

	"github.com/containeroo/tinyflags"
)

// Options holds the application configuration.
type Options struct {
	ConfigPaths      []string       // Paths to YAML config files, merged in order.
	Profile          string         // Profile name to load.
	Parser           string         // Parser override.
	Split            string         // Split delimiter override.
	Fields           []string       // Fields to render.
	HiddenFields     []string       // Fields hidden from display.
	Format           string         // Output format override.
	Search           string         // Initial search query.
	Filters          []string       // Initial include filters.
	Excludes         []string       // Initial exclude filters.
	BufferSize       int            // Maximum raw lines kept in memory.
	BatchSize        int            // Number of lines grouped into one UI update.
	FlushInterval    time.Duration  // Maximum delay before flushing a partial batch.
	Debug            bool           // Enable debug output.
	OverriddenValues map[string]any // Overridden values from environment.
}

// ParseFlags parses flags and environment variables into Options.
func ParseFlags(args []string, version string) (Options, error) {
	opts := Options{}

	tf := tinyflags.NewFlagSet("loggo", tinyflags.ContinueOnError)
	tf.Version(version)
	tf.EnvPrefix("LOGGO_")

	tf.StringSliceVar(&opts.ConfigPaths, "config", []string{}, "Path to YAML config file. Can be repeated; last file wins.").
		Short("c").
		Placeholder("PATH").
		Value()

	tf.StringVar(&opts.Profile, "profile", "", "Profile to load.").
		Short("p").
		Placeholder("NAME").
		Value()

	tf.StringVar(&opts.Parser, "parser", "", "Parser override: auto, json, logfmt, regex, split, raw.").
		Placeholder("TYPE").
		Choices("", "auto", "json", "logfmt", "regex", "split", "raw").
		Value()

	tf.StringVar(&opts.Split, "split", "", "Delimiter for split parser.").
		Placeholder("DELIM").
		Value()

	tf.StringSliceVar(&opts.Fields, "fields", []string{}, "Comma-separated fields to render.").
		Placeholder("LIST").
		Value()

	tf.StringSliceVar(&opts.HiddenFields, "hide-field", []string{}, "Comma-separated fields hidden from display.").
		Placeholder("LIST").
		Value()

	tf.StringVar(&opts.Format, "format", "", "Output format using {field} placeholders.").
		Placeholder("FORMAT").
		Value()

	tf.StringVar(&opts.Search, "search", "", "Initial search query.").
		Short("s").
		Placeholder("TEXT").
		Value()

	tf.StringSliceVar(&opts.Filters, "filter", []string{}, "Comma-separated initial include filters.").
		Short("f").
		Placeholder("LIST").
		Value()

	tf.StringSliceVar(&opts.Excludes, "exclude", []string{}, "Comma-separated initial exclude filters.").
		Short("x").
		Placeholder("LIST").
		Value()

	tf.IntVar(&opts.BufferSize, "buffer-size", 5000, "Maximum raw lines kept in memory.").
		Placeholder("N").
		Validate(func(input int) error {
			if input <= 0 {
				return fmt.Errorf("buffer-size must be a positive integer")
			}

			return nil
		}).
		Value()

	tf.IntVar(&opts.BatchSize, "batch-size", 300, "Number of lines grouped into one UI update.").
		Placeholder("N").
		Validate(func(input int) error {
			if input <= 0 {
				return fmt.Errorf("batch-size must be a positive integer")
			}

			return nil
		}).
		Value()

	tf.DurationVar(&opts.FlushInterval, "flush-interval", 33*time.Millisecond, "Maximum delay before flushing a partial input batch.").
		Placeholder("DURATION").
		Validate(func(input time.Duration) error {
			if input <= 0 {
				return fmt.Errorf("flush-interval must be a positive duration")
			}

			return nil
		}).
		Value()

	tf.BoolVar(&opts.Debug, "debug", false, "Enable debug output.").
		Short("d").
		Value()

	if err := tf.Parse(args); err != nil {
		return Options{}, err
	}

	opts.OverriddenValues = tf.OverriddenValues()

	return opts, nil
}
