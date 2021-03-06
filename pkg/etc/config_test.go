package etc

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

type Envs map[string]string

func TestGetLogLevel(t *testing.T) {
	testCases := []struct {
		name             string
		envs             Envs
		expectedLogLevel logrus.Level
	}{
		{
			name:             "Should return default log level when env is not set",
			expectedLogLevel: logrus.InfoLevel,
		},
		{
			name: "Should return default log level when env has invalid value",
			envs: Envs{
				"SCANNER_LOG_LEVEL": "unknown_level",
			},
			expectedLogLevel: logrus.InfoLevel,
		},
		{
			name: "Should return log level set as env",
			envs: Envs{
				"SCANNER_LOG_LEVEL": "trace",
			},
			expectedLogLevel: logrus.TraceLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setenvs(t, tc.envs)
			assert.Equal(t, tc.expectedLogLevel, GetLogLevel())
		})
	}
}

func TestGetConfig(t *testing.T) {
	testCases := []struct {
		name           string
		envs           Envs
		expectedConfig Config
	}{
		{
			name: "Should return default config",
			expectedConfig: Config{
				API: APIConfig{
					Addr:         ":8080",
					ReadTimeout:  parseDuration(t, "15s"),
					WriteTimeout: parseDuration(t, "15s"),
					IdleTimeout:  parseDuration(t, "60s"),
				},
				TLS: TLSConfig{
					InsecureSkipVerify: false,
				},
				Clair: ClairConfig{
					URL: "http://harbor-harbor-clair:6060",
				},
				Store: Store{
					RedisURL:      "redis://harbor-harbor-redis:6379",
					Namespace:     "harbor.scanner.clair:store",
					PoolMaxActive: 5,
					PoolMaxIdle:   5,
					ScanJobTTL:    parseDuration(t, "1h"),
				},
			},
		},
		{
			name: "Should overwrite default config with envs",
			envs: Envs{
				"SCANNER_API_SERVER_ADDR":            ":7654",
				"SCANNER_API_SERVER_TLS_CERTIFICATE": "/certs/tls.crt",
				"SCANNER_API_SERVER_TLS_KEY":         "/certs/tls.key",
				"SCANNER_API_SERVER_READ_TIMEOUT":    "1h17m",
				"SCANNER_API_SERVER_WRITE_TIMEOUT":   "2h5m",
				"SCANNER_API_SERVER_IDLE_TIMEOUT":    "3m15s",

				"SCANNER_TLS_INSECURE_SKIP_VERIFY": "true",
				"SCANNER_TLS_CLIENTCAS":            "test/data/ca.crt",

				"SCANNER_CLAIR_URL": "https://demo.clair:7080",
			},
			expectedConfig: Config{
				API: APIConfig{
					Addr:           ":7654",
					TLSCertificate: "/certs/tls.crt",
					TLSKey:         "/certs/tls.key",
					ReadTimeout:    parseDuration(t, "1h17m"),
					WriteTimeout:   parseDuration(t, "2h5m"),
					IdleTimeout:    parseDuration(t, "3m15s"),
				},
				TLS: TLSConfig{
					InsecureSkipVerify: false,
				},
				Clair: ClairConfig{
					URL: "https://demo.clair:7080",
				},
				Store: Store{
					RedisURL:      "redis://harbor-harbor-redis:6379",
					Namespace:     "harbor.scanner.clair:store",
					PoolMaxActive: 5,
					PoolMaxIdle:   5,
					ScanJobTTL:    parseDuration(t, "1h"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setenvs(t, tc.envs)

			cfg, err := GetConfig()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedConfig.API, cfg.API)
			assert.Equal(t, tc.expectedConfig.Clair, cfg.Clair)
			assert.Equal(t, tc.expectedConfig.Store, cfg.Store)
		})
	}

}

func TestAPIConfig_IsTLSEnabled(t *testing.T) {
	testCases := []struct {
		name     string
		envs     Envs
		expected bool
	}{
		{
			name: "Should return true when cert and key are set",
			envs: Envs{
				"SCANNER_API_SERVER_TLS_CERTIFICATE": "/certs/tls.crt",
				"SCANNER_API_SERVER_TLS_KEY":         "/certs/tls.key",
			},
			expected: true,
		},
		{
			name: "Should return false when only cert is set",
			envs: Envs{
				"SCANNER_API_SERVER_TLS_CERTIFICATE": "/certs/tls.crt",
			},
			expected: false,
		},
		{
			name: "Should return false when only key is set",
			envs: Envs{
				"SCANNER_API_SERVER_TLS_KEY": "/certs/tls.key",
			},
			expected: false,
		},
		{
			name:     "Should return false when neither cert nor key is set",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setenvs(t, tc.envs)
			config, _ := GetConfig()
			assert.Equal(t, tc.expected, config.API.IsTLSEnabled())
		})
	}
}

func setenvs(t *testing.T, envs Envs) {
	t.Helper()
	os.Clearenv()
	for key, value := range envs {
		err := os.Setenv(key, value)
		require.NoError(t, err)
	}
}

func parseDuration(t *testing.T, s string) time.Duration {
	t.Helper()
	duration, err := time.ParseDuration(s)
	require.NoError(t, err)
	return duration
}
