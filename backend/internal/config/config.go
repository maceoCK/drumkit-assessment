package config

import (
	"encoding/json"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration.
type Config struct {
	AppEnv         string `envconfig:"APP_ENV" default:"local"`
	TurvoBaseURL   string `envconfig:"TURVO_BASE_URL" default:"https://app.turvo.com"`
	TurvoAPIPrefix string `envconfig:"TURVO_API_PREFIX" default:"/v1"`
	// OAuth (preferred)
	TurvoClientID                     string   `envconfig:"TURVO_CLIENT_ID"`
	TurvoClientSecret                 string   `envconfig:"TURVO_CLIENT_SECRET"`
	TurvoAPIKey                       string   `envconfig:"TURVO_API_KEY"`
	TurvoOAuthUsername                string   `envconfig:"TURVO_USERNAME"`
	TurvoOAuthPassword                string   `envconfig:"TURVO_PASSWORD"`
	TurvoOAuthScope                   string   `envconfig:"TURVO_SCOPE" default:"read+trust+write"`
	TurvoOAuthUserType                string   `envconfig:"TURVO_USER_TYPE" default:"business"`
	TurvoTenant                       string   `envconfig:"TURVO_TENANT"`
	TurvoUseAWSSigV4                  bool     `envconfig:"TURVO_USE_AWS_SIGV4" default:"false"`
	WebhookSecret                     string   `envconfig:"WEBHOOK_SECRET"`
	AllowedOrigins                    []string `envconfig:"ALLOWED_ORIGINS" default:"*"`
	LogLevel                          string   `envconfig:"LOG_LEVEL" default:"info"`
	TurvoDefaultCustomerID            int      `envconfig:"TURVO_DEFAULT_CUSTOMER_ID" default:"0"`
	TurvoDefaultOriginLocationID      int      `envconfig:"TURVO_DEFAULT_ORIGIN_LOCATION_ID" default:"0"`
	TurvoDefaultDestinationLocationID int      `envconfig:"TURVO_DEFAULT_DESTINATION_LOCATION_ID" default:"0"`
	AWSRegion                         string   `envconfig:"AWS_REGION" default:"us-east-1"`
	SecretsManagerTurvoSecretName     string   `envconfig:"SECRETS_MANAGER_TURVO_SECRET_NAME"`
}

// Load loads the configuration depending on APP_ENV.
func Load() (*Config, error) {
	var cfg Config

	// In local, load .env first
	if err := godotenv.Load(); err == nil {
		log.Printf("Loaded .env file")
	}

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	// If deployed (any env other than local), fetch secrets from Secrets Manager (optional)
	if cfg.AppEnv != "local" && cfg.SecretsManagerTurvoSecretName != "" {
		secretJSON, err := FetchSecret(cfg.AWSRegion, cfg.SecretsManagerTurvoSecretName)
		if err != nil {
			log.Printf("warning: failed to fetch secrets: %v", err)
		} else {
			// Expected JSON keys include all envs above
			var m map[string]string
			if err := json.Unmarshal([]byte(secretJSON), &m); err == nil {
				if v := m["TURVO_CLIENT_ID"]; v != "" {
					cfg.TurvoClientID = v
				}
				if v := m["TURVO_CLIENT_SECRET"]; v != "" {
					cfg.TurvoClientSecret = v
				}
				if v := m["TURVO_API_KEY"]; v != "" {
					cfg.TurvoAPIKey = v
				}
				if v := m["TURVO_USERNAME"]; v != "" {
					cfg.TurvoOAuthUsername = v
				}
				if v := m["TURVO_PASSWORD"]; v != "" {
					cfg.TurvoOAuthPassword = v
				}
				if v := m["TURVO_SCOPE"]; v != "" {
					cfg.TurvoOAuthScope = v
				}
				if v := m["TURVO_USER_TYPE"]; v != "" {
					cfg.TurvoOAuthUserType = v
				}
				if v := m["TURVO_BASE_URL"]; v != "" {
					cfg.TurvoBaseURL = v
				}
				if v := m["TURVO_TENANT"]; v != "" {
					cfg.TurvoTenant = v
				}
				if v := m["TURVO_API_PREFIX"]; v != "" {
					cfg.TurvoAPIPrefix = v
				}
			}
		}
	}

	log.Printf("Configuration loaded: %+v", cfg)
	return &cfg, nil
}
