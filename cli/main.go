package main

import (
	"code.google.com/p/getopt"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/scjudd/spiceworks"
	"github.com/scjudd/spiceworks/cli/cookiejar"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"text/tabwriter"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const (
	prettySummaryTrunc = 80
)

// md5hash takes any number of strings and returns a hex MD5 hash string.
func md5hash(strs ...string) string {
	h := md5.New()
	for _, s := range strs {
		io.WriteString(h, s)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// getEnvIfBlank returns s unless s == ""; otherwise, it returns the value of
// the environment variable e.
func getEnvIfBlank(s, e string) string {
	if s == "" {
		return os.Getenv(e)
	}
	return s
}

func main() {
	var server, email, password string
	var pretty, help bool

	// Parse command line flags.
	getopt.StringVarLong(&server, "server", 's', "Spiceworks server URL, i.e., helpdesk.aacc.net")
	getopt.StringVarLong(&email, "email", 'e', "Email address to log in to Spiceworks")
	getopt.StringVarLong(&password, "password", 'p', "Password to log in to Spiceworks")
	getopt.BoolVarLong(&pretty, "pretty", 'P', "Prettify output: less machine-readable, more human-readable.")
	getopt.BoolVarLong(&help, "help", 0, "Display usage information.")
	getopt.Parse()

	// If "--help" passed, display usage information.
	if help {
		getopt.Usage()
		os.Exit(0)
	}

	// If server, email, or password weren't specified on the command line, try
	// to read them from environment variables before throwing a fatal error.
	server = getEnvIfBlank(server, "SPICEWORKS_SERVER")
	email = getEnvIfBlank(email, "SPICEWORKS_EMAIL")
	password = getEnvIfBlank(password, "SPICEWORKS_PASSWORD")
	if server == "" || email == "" || password == "" {
		log.Fatal(errors.New("--server, --email, and --password are required!"))
	}

	// Find the cookiejar path.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	jarPath := path.Join(usr.HomeDir, ".spiceworks_cookiejar.json")

	// Open the cookiejar, compare hashes.
	hash := md5hash(server, email, password)
	jar, hashMatch, err := cookiejar.Open(jarPath, hash)
	if err != nil {
		log.Fatal(err)
	}
	defer jar.Flush()

	// Create the Spiceworks Client.
	baseUrl := "http://" + server + "/"
	client := &spiceworks.Client{&http.Client{Jar: jar}, baseUrl, email, password}

	// If the previously stored hash doesn't match the hash we just generated,
	// have the Client attempt to log in.
	if !hashMatch {
		err = client.Login()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Fetch a slice of all open tickets.
	tickets, err := client.Tickets("open")
	if err != nil {
		log.Fatal(err)
	}

	// Print open tickets
	var w io.Writer = os.Stdout
	if pretty {
		w = new(tabwriter.Writer)
		w.(*tabwriter.Writer).Init(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintln(w, "\x1b[1mID\tSUMMARY\tASSIGNEE\x1b[0m")
	}
	for _, ticket := range tickets {
		if ticket.Assignee.FirstName == "" && ticket.Assignee.LastName == "" {
			ticket.Assignee.FirstName = "Unassigned"
		}
		if pretty && len(ticket.Summary) > prettySummaryTrunc-3 {
			ticket.Summary = ticket.Summary[0:prettySummaryTrunc-3] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s %s\n",
			ticket.Id,
			ticket.Summary,
			ticket.Assignee.FirstName,
			ticket.Assignee.LastName,
		)
	}
	if pretty {
		w.(*tabwriter.Writer).Flush()
	}
}
