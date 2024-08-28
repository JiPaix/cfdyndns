# cfdyndns

`cfdyndns` is a Go package that provides a dynamic DNS (DDNS) service for domains managed by Cloudflare. It allows you to automatically update DNS records to point to the current public IP address of the machine running the code.

## Features

- Update DNS records (A or AAAA) to the current public IP address
- Support for both one-time updates and scheduled automatic updates
- Compatible with IPv4 and IPv6 addresses
- Option to set records as proxied or unproxied through Cloudflare

## Installation

To install the `cfdyndns` package, use the following go get command:

```
go get github.com/jipaix/cfdyndns
```

## Usage

Here are some examples of how to use the `cfdyndns` package:

### Initializing the client

```go
import "github.com/jipaix/cfdyndns"

// Initialize with your Cloudflare API token
client, err := cfdyndns.New("your-cloudflare-api-token")
if err != nil {
    log.Fatal(err)
}
```

### Updating a DNS record once

```go
err := client.Set("example.com", "subdomain", false)
if err != nil {
    log.Printf("Error updating DNS: %v", err)
}
```

This will update the A or AAAA record for `subdomain.example.com` to point to the current public IP address. The `false` parameter means the record will not be proxied through Cloudflare.

### Setting up automatic updates

```go
stop, err := client.Auto("example.com", "subdomain", true, "*/15 * * * *")
if err != nil {
    log.Printf("Error setting up auto-update: %v", err)
}
defer stop() // Call this when you want to stop the automatic updates
```

This will update the DNS record immediately and then every 15 minutes. The `true` parameter means the record will be proxied through Cloudflare.

## Notes

- Make sure you have a valid Cloudflare API token with the necessary permissions to modify DNS records for your domain.
- The package automatically detects whether to use an A record (for IPv4) or an AAAA record (for IPv6) based on the detected public IP address.
- To update the zone apex (naked domain), use "@" or an empty string as the subdomain.

## Cron Specification

The `Auto` function uses a cron expression to schedule updates. The cron expression follows the standard format as described in the [cron Wikipedia page](https://en.wikipedia.org/wiki/Cron). The format is:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of the month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

The cron library used in this package (github.com/robfig/cron/v3) supports various expressions. Here are some examples:

```go
"30 * * * *"                        // Every hour on the half hour
"30 3-6,20-23 * * *"                // Every day at 30 minutes past the hour, between 3-6am and 8-11pm
"CRON_TZ=Asia/Tokyo 30 04 * * *"    // Every day at 04:30 Tokyo time
"@hourly"                           // Every hour, starting an hour from now
"@every 1h30m"                      // Every hour thirty, starting an hour thirty from now
"@daily"                            // Every day
```

For more detailed information on cron expressions, refer to the [robfig/cron documentation](https://pkg.go.dev/github.com/robfig/cron).

## Credits

This package wouldn't be possible without the following excellent libraries:

- [cloudflare-go](https://github.com/cloudflare/cloudflare-go) by Cloudflare: Official Cloudflare API v4 Go client
- [cron](https://github.com/robfig/cron) by [Rob Figueiredo](https://github.com/robfig): A cron library for Go
- [charmbracelet/log](https://github.com/charmbracelet/log) by [Charm](https://github.com/charmbracelet): A minimal and colorful Go logging library

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

see [LICENSE](LICENSE)