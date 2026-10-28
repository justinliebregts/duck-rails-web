package config

import "os"

func CloudflareZoneID() string {
	return os.Getenv("CLOUDFLARE_ZONE_ID")
}

func CloudflareAPIToken() string {
	return os.Getenv("CLOUDFLARE_API_TOKEN")
}

func CloudflareEnabled() bool {
	return CloudflareZoneID() != "" && CloudflareAPIToken() != ""
}
