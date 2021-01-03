package utils

var (
	// DefaultLogFile define the default log output file.
	DefaultLogFile = "/var/log/%s.log"
)

const (
	// DefaultLogMaxSize define the default log size per file, unit (M).
	DefaultLogMaxSize = 500
	// DefaultLogMaxBackups define the default log max-backup num.
	DefaultLogMaxBackups = 15
	// DefaultLogMaxAge define the max age of the log files.
	DefaultLogMaxAge = 30
	// DefaultLogCompress define whether the log need compress.
	DefaultLogCompress = true

	// DefaultWorkThread used to define the num of update dns-rules threads num.
	DefaultWorkThreads = 4

	DefaultRootDomain = "binddns.com"

	DefaultEnableHttpApi = true
)
