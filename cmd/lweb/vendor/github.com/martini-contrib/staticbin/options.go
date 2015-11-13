package staticbin

// Options represents configuration options for the staticbin.Static middleware.
type Options struct {
	// SkipLogging will disable [Static] log messages when a static file is served.
	SkipLogging bool
	// IndexFile defines which file to serve as index if it exists.
	IndexFile string
}

// retrieveOptions retrieves an options from the array of options.
func retrieveOptions(options []Options) Options {
	var opt Options

	if len(options) > 0 {
		opt = options[0]
	}

	// Set the default value to opt.IndexFile.
	if opt.IndexFile == "" {
		opt.IndexFile = defaultIndexFile
	}

	return opt
}
