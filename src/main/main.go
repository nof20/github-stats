package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Organization is a GH organization, less the reference to Plan.
// https://godoc.org/github.com/google/go-github/github#Organization
type Organization struct {
	Login             bigquery.NullString `bigquery:"login"`
	ID                bigquery.NullInt64  `bigquery:"id"`
	AvatarURL         bigquery.NullString `bigquery:"avatar_url"`
	HTMLURL           bigquery.NullString `bigquery:"html_url"`
	Name              bigquery.NullString `bigquery:"name"`
	Company           bigquery.NullString `bigquery:"company"`
	Blog              bigquery.NullString `bigquery:"blog"`
	Location          bigquery.NullString `bigquery:"location"`
	Email             bigquery.NullString `bigquery:"email"`
	Description       bigquery.NullString `bigquery:"description"`
	PublicRepos       bigquery.NullInt64  `bigquery:"public_repos"`
	PublicGists       bigquery.NullInt64  `bigquery:"public_gists"`
	Followers         bigquery.NullInt64  `bigquery:"followers"`
	Following         bigquery.NullInt64  `bigquery:"following"`
	CreatedAt         bigquery.NullTime   `bigquery:"created_at"`
	UpdatedAt         bigquery.NullTime   `bigquery:"updated_at"`
	TotalPrivateRepos bigquery.NullInt64  `bigquery:"total_private_repos"`
	OwnedPrivateRepos bigquery.NullInt64  `bigquery:"owned_private_repos"`
	PrivateGists      bigquery.NullInt64  `bigquery:"private_gists"`
	DiskUsage         bigquery.NullInt64  `bigquery:"disk_usage"`
	Collaborators     bigquery.NullInt64  `bigquery:"collaborators"`
	BillingEmail      bigquery.NullString `bigquery:"billing_email"`
	Type              bigquery.NullString `bigquery:"type"`
	NodeID            bigquery.NullString `bigquery:"node_id"`

	// API URLs
	URL              bigquery.NullString `bigquery:"url"`
	EventsURL        bigquery.NullString `bigquery:"events_url"`
	HooksURL         bigquery.NullString `bigquery:"hooks_url"`
	IssuesURL        bigquery.NullString `bigquery:"issues_url"`
	MembersURL       bigquery.NullString `bigquery:"members_url"`
	PublicMembersURL bigquery.NullString `bigquery:"public_members_url"`
	ReposURL         bigquery.NullString `bigquery:"repos_url"`
}

func main() {
	ctx := context.Background()

	// Authenticate to GitHub using saved personal access token
	dat, err := ioutil.ReadFile("github-access-token.txt")
	if err != nil {
		log.Fatalf("Cannot load GitHub access token: %q", err)
	}
	s := string(dat)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s},
	)
	tc := oauth2.NewClient(ctx, ts)
	gc := github.NewClient(tc)

	// Make GitHub API calls
	orgs, _, err := gc.Organizations.List(ctx, "nof20", nil)
	if err != nil {
		log.Fatalf("Error: %q", err)
		os.Exit(1)
	}

	// Authenticate to BigQuery and create table
	projectID := "cron-trigger-test"
	bc, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Cannot create BigQuery client: %v", err)
	}
	ds := bc.Dataset("GitHub")
	if err := ds.Create(ctx, nil); err != nil {
		log.Printf("Cannot create BigQuery dataset: %v", err)
	}
	t := ds.Table("Organizations")
	schema, err := bigquery.InferSchema(Organization{})
	if err != nil {
		log.Printf("Cannot infer BigQuery schema: %q", err)
	}
	if err = t.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
		log.Printf("Cannot create table: %q", err)
	}
	u := t.Uploader()
	if err := u.Put(ctx, orgs); err != nil {
		log.Fatalf("Cannot save data to BigQuery: %q", err)
	}

	// Save to BigQuery

}
