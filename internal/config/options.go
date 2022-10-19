package config

// Options
type Options struct {
	TimeZone         string
	MaxAllowedOffset int

	DryRun      bool
	Verbose     bool
	Recursive   bool
	IgnoreErrs  bool
	Interactive bool
}
