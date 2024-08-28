package cfdyndns

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func ExampleNew() {
	token := os.Getenv("CF_API_TOKEN")
	if token == "" {
		log.Fatal("CF_API_TOKEN environment variable is not set")
	}
	cfdyndns, err := New(token)
	if err != nil {
		log.Fatal("Error creating cfdyndns instance", "error", err)
	}

	// use cfdyndns

	cfdyndns.Set("example.com", "www", false)
}

func ExampleSet() {
	token := os.Getenv("CF_API_TOKEN")
	if token == "" {
		log.Fatal("CF_API_TOKEN environment variable is not set")
	}
	cfdyndns, err := New(token)
	if err != nil {
		log.Fatal("Error creating cfdyndns instance", "error", err)
	}

	err = cfdyndns.Set("example.com", "www", false)
	if err != nil {
		log.Fatal("Couldn't set record", "error", err)
	}
}

func ExampleAuto() {
	token := os.Getenv("CF_API_TOKEN")
	if token == "" {
		log.Fatal("CF_API_TOKEN environment variable is not set")
	}
	cfdyndns, err := New(token)
	if err != nil {
		log.Fatal("Error creating cfdyndns instance", "error", err)
	}

	stop1, err := cfdyndns.Auto("example.com", "www", false, "@daily")
	if err != nil {
		log.Fatal("Couldn't auto set records", "error", err)
	}

	stop2, err := cfdyndns.Auto("example.com", "www", false, "@daily")
	if err != nil {
		log.Fatal("Couldn't auto set records", "error", err)
	}

	// later
	stop1()
	// or
	stop2()
}

func TestNew(t *testing.T) {
	token := os.Getenv("CF_API_TOKEN")
	if token == "" {
		t.Fatal("CF_API_TOKEN environment variable is not set")
	}
	_, err := New(token)
	if err != nil {
		t.Fatal("Error creating cfdyndns instance", "error", err)
	}
}

func TestNewNoToken(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("Expected error creating cfdyndns instance without token")
	}
}

func TestSet(t *testing.T) {
	token := os.Getenv("CF_API_TOKEN")
	domain := os.Getenv("CF_DOMAIN")
	subdomain1 := os.Getenv("CF_SUBDOMAIN1")
	subdomain2 := os.Getenv("CF_SUBDOMAIN2")

	if token == "" {
		t.Fatal("CF_API_TOKEN environment variable is not set")
	}
	if domain == "" {
		t.Fatal("CF_DOMAIN environment variable is not set")
	}
	if subdomain1 == "" {
		t.Fatal("CF_SUBDOMAIN1 environment variable is not set")
	}
	if subdomain2 == "" {
		t.Fatal("CF_SUBDOMAIN2 environment variable is not set")
	}

	cfdyndns, err := New(token)
	if err != nil {
		t.Fatal("Error creating cfdyndns instance", "error", err)
	}

	err = cfdyndns.Set(domain, subdomain1, true)
	if err != nil {
		t.Fatal("Couldn't set proxy record", "error", err)
	}

	err = cfdyndns.Set(domain, subdomain1, false)
	if err != nil {
		t.Fatal("Couldn't unproxify record", "error", err)
	}

	err = cfdyndns.Set(domain, subdomain2, false)
	if err != nil {
		t.Fatal("Couldn't set non-proxy record", "error", err)
	}

	err = cfdyndns.Set(domain, subdomain2, true)
	if err != nil {
		t.Fatal("Couldn't proxify record", "error", err)
	}
}

func TestAuto(t *testing.T) {
	token := os.Getenv("CF_API_TOKEN")
	domain := os.Getenv("CF_DOMAIN")
	subdomain1 := os.Getenv("CF_SUBDOMAIN1")
	subdomain2 := os.Getenv("CF_SUBDOMAIN2")

	if token == "" {
		t.Fatal("CF_API_TOKEN environment variable is not set")
	}
	if domain == "" {
		t.Fatal("CF_DOMAIN environment variable is not set")
	}
	if subdomain1 == "" {
		t.Fatal("CF_SUBDOMAIN environment variable is not set")
	}
	if subdomain2 == "" {
		t.Fatal("CF_SUBDOMAIN environment variable is not set")
	}

	cfdyndns, err := New(token)
	if err != nil {
		t.Fatal("Error creating cfdyndns instance", "error", err)
	}

	stop, err := cfdyndns.Auto(domain, subdomain1, true, "* * * * *")
	if err != nil {
		t.Fatal("Couldn't auto set proxy record", "error", err)
	}

	stop2, err := cfdyndns.Auto(domain, subdomain1, true, "* * * * *")
	if err != nil {
		t.Fatal("Couldn't auto set proxy record", "error", err)
	}

	log.Info("Immediately stopping cron job1")
	stop()

	log.Info("Sleeping for 2 minutes to make sure cron job2 is done.")
	time.Sleep(time.Minute * 2)

	stop2()
}
