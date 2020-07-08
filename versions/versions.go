package versions

const (
	// PROGRAM is the name of this software
	PROGRAM = "secure-dns"
	// VERSION should be replaced at compile
	VERSION = "UNKNOWN"
	// BUILDHASH should be replaced at compile
	BUILDHASH = ""

	// USERAGENT for DNS over HTTPS
	USERAGENT = PROGRAM + "/" + VERSION + " https://github.com/jinliming2/secure-dns"
)
