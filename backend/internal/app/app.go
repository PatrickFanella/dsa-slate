package app

import "os"

const serviceName = "dsa-slate"

func ServiceName() string {
	return serviceName
}

func AdminAPIToken() string {
	return os.Getenv("ADMIN_API_TOKEN")
}

func GoogleCivicAPIKey() string {
	return os.Getenv("GOOGLE_CIVIC_API_KEY")
}
