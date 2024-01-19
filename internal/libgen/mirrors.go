package libgen

import "net/url"

// searchMirors contain all valid mirrors to be used
// for querying against library genesis
var searchMirors = []url.URL{
	{
		Scheme: "https",
		Host:   "libgen.is",
		Path:   "search.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.rs",
		Path:   "search.php",
	},
}
