package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/go-iptables/iptables"
	"github.com/spf13/cobra"
)

const (
	tableNAT    = "nat"
	chainOutput = "OUTPUT"

	envVarEgressPort              = "EGRESS_PORT"
	envVarRedirectEgressCIDRBlock = "REDIRECT_EGRESS_CIDR_BLOCK"
	envVarConfigDir               = "CONFIG_DIR"
	envVarAdminPort               = "ADMIN_PORT"
	envVarXDSAPIHost              = "XDS_API_HOST"
	envVarXDSAPIPort              = "XDS_API_PORT"
)

var envVars = []string{
	envVarEgressPort,
	envVarRedirectEgressCIDRBlock,
	envVarConfigDir,
	envVarAdminPort,
	envVarXDSAPIHost,
	envVarXDSAPIPort,
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:  "prepare-envoy",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		env, err := parseEnv()
		if err != nil {
			panic(err)
		}

		err = addIPTableRedirect(env)
		if err != nil {
			panic(err)
		}

		err = outputEnvoyConfig(env)
		if err != nil {
			panic(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseEnv() (map[string]string, error) {
	fail := func(key string) error {
		return fmt.Errorf("%s not set", key)
	}

	env := map[string]string{}

	for _, envVar := range envVars {
		val, ok := os.LookupEnv(envVar)
		if !ok {
			return nil, fail(envVar)
		}
		env[envVar] = val
	}

	return env, nil
}

func addIPTableRedirect(env map[string]string) error {
	ipt, err := iptables.New()
	if err != nil {
		panic(err)
	}

	rulespecs := []string{
		"-p", "tcp",
		"-d", env[envVarRedirectEgressCIDRBlock],
		"-j", "REDIRECT",
		"--to-port", env[envVarEgressPort],
		"-m", "comment", "--comment", "\"lattice redirect to envoy\"",
	}
	return ipt.Append(tableNAT, chainOutput, rulespecs...)
}

func outputEnvoyConfig(env map[string]string) error {
	if err := os.MkdirAll(env[envVarConfigDir], 0644); err != nil {
		return err
	}

	configFilename := filepath.Join(env[envVarConfigDir], "config.json")

	xdsAPIURL := fmt.Sprintf("%v:%v", env[envVarXDSAPIHost], env[envVarXDSAPIPort])

	// TODO: factor this out into an envoy config struct
	contents := fmt.Sprintf(`{
  "listeners": [],
  "lds": {
    "cluster": "xds-api",
    "refresh_delay_ms": 10000
  },
  "admin": {
    "access_log_path": "/dev/null",
    "address": "tcp://0.0.0.0:%v"
  },
  "cluster_manager": {
    "clusters": [
      {
        "name": "xds-api",
        "connect_timeout_ms": 250,
        "type": "static",
        "lb_type": "round_robin",
        "hosts": [
          {
            "url": "tcp://%v"
          }
        ]
      }
    ],
    "cds": {
      "cluster": {
        "name": "xds-api-cds",
        "connect_timeout_ms": 250,
        "type": "static",
        "lb_type": "round_robin",
        "hosts": [
          {
            "url": "tcp://%v"
          }
        ]
      },
      "refresh_delay_ms": 10000
    },
    "sds": {
      "cluster": {
        "name": "xds-api-sds",
        "connect_timeout_ms": 250,
        "type": "static",
        "lb_type": "round_robin",
        "hosts": [
          {
            "url": "tcp://%v"
          }
        ]
      },
      "refresh_delay_ms": 10000
    }
  }
}`,
		env[envVarAdminPort],
		xdsAPIURL,
		xdsAPIURL,
		xdsAPIURL,
	)

	return ioutil.WriteFile(configFilename, []byte(contents), 0644)
}
