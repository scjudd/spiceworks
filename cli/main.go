package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/scjudd/spiceworks"
	"github.com/scjudd/spiceworks/cli/cookies"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"text/tabwriter"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const (
	summaryTrunc = 80
)

func main() {
	var server, email, password string
	flag.StringVar(&server, "s", "", "Spiceworks server URL, i.e., helpdesk.aacc.net")
	flag.StringVar(&email, "e", "", "Email address to log in to Spiceworks")
	flag.StringVar(&password, "p", "", "Password to log in to Spiceworks")
	flag.Parse()

	if server == "" {
		server = os.Getenv("SPICEWORKS_SERVER")
	}

	if email == "" {
		email = os.Getenv("SPICEWORKS_EMAIL")
	}

	if password == "" {
		password = os.Getenv("SPICEWORKS_PASSWORD")
	}

	if server == "" || email == "" || password == "" {
		log.Fatal(errors.New("-s, -e, and -p are required!"))
	}

	// BUG(scjudd): If this file exists, the provided server, email, and password
	// are ignored. We should hash these values along with the cookiejar data and
	// verify that they match up before short-circuiting the login process.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	cookiepath := path.Join(usr.HomeDir, ".spiceworks_cookies.json")

	jar, err := cookies.Open(cookiepath)
	if err != nil {
		log.Fatal(err)
	}
	defer jar.Flush()

	baseUrl := "http://" + server + "/"
	client := &spiceworks.Client{&http.Client{Jar: jar}, baseUrl, email, password}

	// BUG(scjudd): The cookies package is extremely basic and doesn't, at present
	// do anything to keep track of cookie expiration. This means an expired
	// cookie will short-circuit the login process and prevent us from
	// successfully making authenticated requests.
	u, err := url.Parse("http://helpdesk.aacc.net/login")
	if err != nil {
		log.Fatal(err)
	}
	if len(jar.Cookies(u)) == 0 {
		err = client.Login()
		if err != nil {
			log.Fatal(err)
		}
	}

	tickets, err := client.Tickets("open")
	if err != nil {
		log.Fatal(err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "\x1b[1mID\tSUMMARY\tASSIGNEE\x1b[0m")
	for _, ticket := range tickets {
		if len(ticket.Summary) > summaryTrunc-3 {
			ticket.Summary = ticket.Summary[0:summaryTrunc-3] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s %s\n",
			ticket.Id,
			ticket.Summary,
			ticket.Assignee.FirstName,
			ticket.Assignee.LastName,
		)
	}
	w.Flush()
}
