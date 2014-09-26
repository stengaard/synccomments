package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	podio "github.com/andreas/podio-go"
)

func die(e string) {
	fmt.Fprintln(os.Stderr, e)
	os.Exit(1)
}

func dieErr(err error) {
	if err != nil {
		die(err.Error())
	}
}

func usage() {

	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintln(os.Stderr, "Client settings can be found in Account Settings -> API Key")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "App information can be found in the individual App settings")
	fmt.Fprintln(os.Stderr, "dropdown menu -> Developer")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "You must supply all flags")

}

func pdie(s string) {
	flag.Usage()
	fmt.Fprintln(os.Stderr, "")
	die(s)
}

func main() {
	var (
		fromId  = flag.Int("from", 0, "the app id to move comments from")
		toId    = flag.Int("to", 0, "the app id to move comments to")
		toToken = flag.String("totoken", "", "app token from the receiving app")

		clientId     = flag.String("clientid", "", "client id")
		clientSecret = flag.String("clientsecret", "", "client secret")
		force        = flag.Bool("f", false, "force comment inclusion even if preflight checks fail")
	)

	flag.Usage = usage
	flag.Parse()

	if *fromId == 0 || *toId == 0 {
		pdie("Please supply ids of both originating and receiving apps")
	}

	if *toToken == "" {
		pdie("Please supply token of the receiving app")
	}

	if *clientId == "" || *clientSecret == "" {
		pdie("Please supply client id and client secret. Account Settings -> API Keys")
	}

	from := uint(*fromId)
	to := uint(*toId)

	auth, err := podio.AuthWithAppCredentials(*clientId, *clientSecret, to, *toToken)
	dieErr(err)

	api := podio.NewClient(auth)

	fmt.Println("preflight checks:")

	fromItems, err := api.GetItems(from)
	dieErr(err)
	fmt.Printf(" Found %d items in originating app\n", len(fromItems.Items))
	toItems, err := api.GetItems(to)
	dieErr(err)
	fmt.Printf(" Found %d items in target app\n", len(toItems.Items))

	type itemLink struct {
		from, to *podio.Item
	}

	// title is the only thing that we can use to track this
	idx := map[string]*itemLink{}

	for _, item := range fromItems.Items {
		idx[item.Title] = &itemLink{
			from: item,
		}
	}

	missOrg := []string{}
	for _, item := range toItems.Items {
		link, ok := idx[item.Title]
		if !ok {
			missOrg = append(missOrg, item.Title)
			continue
		}
		link.to = item
	}
	if len(missOrg) > 0 {
		fmt.Println("Missing in the originating app:\n", strings.Join(missOrg, "\n - "))
	}

	missTarget := []string{}
	for title, link := range idx {
		if link.to == nil {
			missTarget = append(missTarget, title)
			continue
		}
		if link.to == nil || link.from == nil {
			panic("invariant broken - need both from and to")
		}
	}
	if len(missTarget) > 0 {
		fmt.Println("Missing in the target app:\n", strings.Join(missTarget, "\n -"))

	}

	if len(missOrg) > 0 || len(missTarget) > 0 {
		fmt.Println("incoherent world view")
		if !*force {
			os.Exit(1)
		}
	} else {
		fmt.Println(":) everything lines up - moving comments")
	}

	const template = "Imported comment by %s @ %s:\n%s"
	for title, link := range idx {
		comments, err := api.GetComments("item", link.from.Id)
		if err != nil {
			fmt.Println("err fetching comments on", title)
			dieErr(err)
		}
		fmt.Printf("moving %d comments with title: %s \n", len(comments), title)

		targetComments, err := api.GetComments("item", link.to.Id)
		dieErr(err)

		seenComment := map[string]bool{}
		for _, comment := range targetComments {
			seenComment[comment.ExternalId] = true
		}

		skipped := 0
		for _, comment := range comments {
			extId := fmt.Sprintf("%d/%d", link.from.Id, comment.Id)
			if seenComment[extId] {
				skipped++
				continue
			}
			d := comment.CreatedOn.Format("2006-01-02")
			msg := fmt.Sprintf(template, comment.CreatedBy.Name, d, comment.Value)
			params := map[string]interface{}{"external_id": extId}
			api.CommentAttr("item", link.to.Id, params, msg)
		}
		if skipped > 0 {
			fmt.Println("Skipped", skipped, "comments since they were there already")
		}
	}

}
