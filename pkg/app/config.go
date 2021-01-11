package app

// Config stores input from command line arguments as well as configs set
// globally. It exposes lancp's core functionality, like running in send and
// receive mode.
type Config struct {
	// FilePath points to a file on disk that will be sent to the receiver.
	FilePath string

	// Port that lancp runs on locally, or listens for messages on locally.
	// Stored in the format ":0000".
	Port string

	// TLSPort is the port that lancp communicates via TLS on. Stored in the
	// format ":0000".
	TLSPort string
}
