// Package cfdyndns provides a dynamic DNS (DDNS) service for domains managed by Cloudflare.
//
// This package allows users to automatically update DNS records for their Cloudflare-managed domains
// to point to the current public IP address of the machine running the code. It supports both IPv4 (A records)
// and IPv6 (AAAA records) addresses.
//
// Key features:
//   - One-time updates of DNS records to the current public IP (Set method)
//   - Scheduled automatic updates using cron jobs (Auto method)
//   - Support for both proxied and unproxied DNS records
//   - Handles subdomains and zone apex updates
//
// This package is particularly useful for maintaining up-to-date DNS records for machines
// with dynamic IP addresses, effectively turning a Cloudflare-managed domain into a dynamic DNS service.
package cfdyndns

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"time"

	l "github.com/charmbracelet/log"
	"github.com/cloudflare/cloudflare-go"
	"github.com/jipaix/whatsmyip"
	cr "github.com/robfig/cron/v3"
)

var log = setupLogger()

// cfdyndns represents the main structure for interacting with the Cloudflare DNS API.
type cfdyndns struct {
	api  *cloudflare.API
	ip   string
	v4   bool
	cron *cr.Cron
}

// New creates a new instance of cfdyndns with the provided Cloudflare API token.
//
// It initializes the Cloudflare API client, detects the current IP address,
// and sets up a cron scheduler for automatic updates.
//
// Parameters:
//   - token: A string containing the Cloudflare API token.
//
// Returns:
//   - *cfdyndns: A pointer to the newly created cfdyndns instance.
//   - error: An error if any issues occur during initialization.
func New(token string) (*cfdyndns, error) {
	api, err := cloudflare.NewWithAPIToken(token)

	if err != nil {
		log.Error("Error creating Cloudflare API client", "error", err)
		return nil, err
	}

	ip, _, err := whatsmyip.Get()
	if err != nil {
		return nil, err
	}

	ipnet := net.ParseIP(ip)
	if ipnet == nil {
		log.Error("Error parsing IP address", "ip", ip)
		return nil, errors.New("could not parse IP")
	}

	v4 := ipnet.To4() != nil
	if v4 {
		log.Infof("IP (v4): %s", ip)
	} else {
		log.Info("IP (v6): %s", ip)
	}

	cron := cr.New()

	return &cfdyndns{api, ip, v4, cron}, nil
}

// Set updates or creates a DNS record for the specified domain and subdomain.
//
// Parameters:
//   - domain: The main domain name (zone) to update.
//   - subdomain: The subdomain to update or create. Use "@" or an empty string for the zone apex.
//   - proxied: A boolean indicating whether the record should be proxied through Cloudflare.
//
// Returns:
//   - error: An error if any issues occur during the update process.
func (ctx *cfdyndns) Set(domain string, subdomain string, proxied bool) error {
	domain, subdomain = fixZoneAndRecord(domain, subdomain)

	// Get the zone ID for the domain
	id, err := ctx.api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	log.Infof("Found domain %s", domain)

	records, _, err := ctx.api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(id), cloudflare.ListDNSRecordsParams{
		Name: subdomain,
	})
	if err != nil {
		log.Error("Error listing DNS records", "error", err)
		return err
	}

	toAdd := cloudflare.CreateDNSRecordParams{
		Type: (func() string {
			if ctx.v4 {
				return "A"
			}
			return "AAAA"
		})(),
		Name:    subdomain,
		Content: ctx.ip,
		Proxied: &proxied,
	}

	var r cloudflare.DNSRecord
	if len(records) == 0 {
		log.Infof("Adding %s", subdomain)
		r, err = ctx.api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(id), toAdd)
		if err != nil {
			return err
		}
		return nil
	}

	log.Infof("Updating %s", subdomain)
	r, err = ctx.api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(id), cloudflare.UpdateDNSRecordParams{
		ID:      records[0].ID,
		Name:    toAdd.Name,
		Type:    toAdd.Type,
		Content: toAdd.Content,
		Proxied: &proxied,
	})

	if err != nil {
		log.Error("Error updating DNS record", "error", err)
		return err
	}

	log.Info("Done", "record", subdomain, "type", r.Type, "ip", ctx.ip, "proxied", proxied)
	return nil
}

// Auto sets up automatic updating of a DNS record on a specified cron schedule.
//
// It immediately sets the DNS record and then schedules future updates based on the provided cron expression.
//
// Parameters:
//   - domain: The main domain name (zone) to update.
//   - subdomain: The subdomain to update or create. Use "@" or an empty string for the zone apex.
//   - proxied: A boolean indicating whether the record should be proxied through Cloudflare.
//   - cron: A string containing a valid cron expression for scheduling updates.
//
// Returns:
//   - stop: A function that can be called to stop the automatic updates.
//   - error: An error if any issues occur during setup or the initial update.
func (ctx *cfdyndns) Auto(domain string, subdomain string, proxied bool, cron string) (stop func(), err error) {
	domain, subdomain = fixZoneAndRecord(domain, subdomain)

	err = ctx.Set(domain, subdomain, proxied)
	if err != nil {
		return nil, err
	}

	entryID, err := ctx.cron.AddFunc(cron, func() {
		ctx.Set(domain, subdomain, proxied)
	})

	if err != nil {
		log.Error("Error scheduling cron job", "error", err, "record", subdomain)
		return nil, err
	}

	stop = func() {
		ctx.cron.Remove(entryID)
		log.Info("Stopped cron job", "record", subdomain)
	}

	return stop, nil
}

// Stop halts all scheduled cron jobs for automatic updates.
func (ctx *cfdyndns) Stop() {
	ctx.cron.Stop()
}

func fixZoneAndRecord(domain string, subdomain string) (string, string) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	subdomain = strings.ReplaceAll(strings.TrimSpace(strings.ToLower(subdomain)), domain, "")
	subdomain = fmt.Sprintf("%s.%s", subdomain, domain)
	return domain, subdomain
}

// setupLogger initializes and returns a configured logger based on the APP_ENV environment variable.
//
// The function sets the log level according to the following APP_ENV values:
//   - "local", "dev", "development": Debug level
//   - "test", "staging": Info level
//   - "prod", "production": Maximum level (effectively disabling logging)
//   - If APP_ENV is not set: Info level
//   - Any other value: Maximum level
//
// The logger is configured with the following options:
//   - Output to stderr
//   - Timestamp reporting enabled
//   - Caller reporting disabled
//   - Time format set to time.DateTime
//   - Prefix set to "🌐 "
//
// Returns:
//   - *github.com/charmbracelet/log.Logger: A configured logger instance
func setupLogger() *l.Logger {
	env, ok := os.LookupEnv("APP_ENV")
	var lvl l.Level
	if !ok {
		lvl = l.InfoLevel
	} else {
		// Set log level based on APP_ENV
		switch strings.ToLower(env) {
		case "local":
			lvl = l.DebugLevel
		case "dev":
			lvl = l.DebugLevel
		case "development":
			lvl = l.DebugLevel
		case "prod":
			lvl = math.MaxInt32 // Effectively disable logging
		case "production":
			lvl = math.MaxInt32 // Effectively disable logging
		case "test":
			lvl = l.InfoLevel
		case "staging":
			lvl = l.InfoLevel
		default:
			lvl = math.MaxInt32 // Effectively disable logging
		}
	}

	return l.NewWithOptions(os.Stderr, l.Options{
		ReportTimestamp: true,
		ReportCaller:    false,
		TimeFormat:      time.DateTime,
		Level:           lvl,
		Prefix:          "🌩️ ",
	})
}
