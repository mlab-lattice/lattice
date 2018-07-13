package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/coreos/go-iptables/iptables"
	"github.com/spf13/cobra"
)

const (
	tableNAT      = "nat"
	chainOutput   = "OUTPUT"
	interfaceName = "eth0"
	envoyUID      = "1000"

	envVarEgressPortHTTP              = "EGRESS_PORT_HTTP"
	envVarEgressPortTCP               = "EGRESS_PORT_TCP"
	envVarRedirectEgressCIDRBlockHTTP = "REDIRECT_EGRESS_CIDR_BLOCK_HTTP"
	envVarRedirectEgressCIDRBlockTCP  = "REDIRECT_EGRESS_CIDR_BLOCK_TCP"
	envVarConfigDir                   = "CONFIG_DIR"
	envVarAdminPort                   = "ADMIN_PORT"
	envVarXDSAPIVersion               = "XDS_API_VERSION"
	envVarXDSAPIHost                  = "XDS_API_HOST"
	envVarXDSAPIPort                  = "XDS_API_PORT"

	// XXX: needed for V2 config (`--service-cluster` and `--service-node` do not set this appropriately)
	envVarServiceCluster = "SERVICE_CLUSTER"
	envVarServiceNode    = "SERVICE_NODE"

	DefaultXDSClusterName             = "xds-api"
	DefaultXDSClusterRefreshDelayMS   = 10000
	DefaultXDSClusterConnectTimeoutMS = 250
	DefaultXDSClusterConnectTimeout   = "0.25s"
)

var envVars = []string{
	envVarEgressPortHTTP,
	envVarEgressPortTCP,
	envVarRedirectEgressCIDRBlockHTTP,
	envVarRedirectEgressCIDRBlockTCP,
	envVarConfigDir,
	envVarAdminPort,
	envVarXDSAPIVersion,
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

		err = addIPTableRedirects(env)
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
	env := map[string]string{}

	getEnvVar := func(key string) error {
		val, ok := os.LookupEnv(key)
		if !ok {
			return fmt.Errorf("%s not set", key)
		}
		env[key] = val
		return nil
	}

	for _, envVar := range envVars {
		if err := getEnvVar(envVar); err != nil {
			return nil, err
		}
	}

	// XXX: see comment above
	if env[envVarXDSAPIVersion] == "2" {
		for _, envVar := range []string{envVarServiceCluster, envVarServiceNode} {
			if err := getEnvVar(envVar); err != nil {
				return nil, err
			}
		}
	}

	return env, nil
}

func localIP() (string, error) {
	interface_, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", err
	}
	addresses, err := interface_.Addrs()
	if err != nil {
		return "", err
	}
	if len(addresses) != 1 {
		return "", fmt.Errorf("expected 1 IP address for interface %v, got %v", interfaceName, addresses)
	}
	// addresses[0].String() give CIDR notation
	address, _, err := net.ParseCIDR(addresses[0].String())
	if err != nil {
		return "", err
	}
	return address.String(), nil
}

func networkContainsIP(cidr, address string) (bool, error) {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}
	ip := net.ParseIP(address)
	if ip == nil {
		return false, fmt.Errorf("couldn't parse IP address: %v", address)
	}
	return cidrNet.Contains(ip), nil
}

func addIPTableRedirects(env map[string]string) error {
	ipt, err := iptables.New()
	if err != nil {
		panic(err)
	}

	// set up exception for outgoing packets destined for this container
	localIP_, err := localIP()
	if err != nil {
		return err
	}
	networkContainsIP_, err := networkContainsIP(env[envVarRedirectEgressCIDRBlockTCP], localIP_)
	if err != nil {
		return err
	} else if !networkContainsIP_ {
		return fmt.Errorf("CIDR %v does not contain local IP address %v",
			envVarRedirectEgressCIDRBlockTCP, localIP_)
	}
	rulespecs := []string{
		"-p", "tcp",
		"-d", localIP_,
		"-j", "ACCEPT",
		"-m", "comment", "--comment",
		fmt.Sprintf("\"lattice local IP redirect exception\""),
	}
	err = ipt.Append(tableNAT, chainOutput, rulespecs...)
	if err != nil {
		return err
	}

	redirects := map[string]map[string]string{
		"HTTP": map[string]string{
			"cidr": envVarRedirectEgressCIDRBlockHTTP,
			"port": envVarEgressPortHTTP,
		},
		"TCP": map[string]string{
			"cidr": envVarRedirectEgressCIDRBlockTCP,
			"port": envVarEgressPortTCP,
		},
	}
	for protocol, parameters := range redirects {
		rulespecs := []string{
			"-p", "tcp",
			"-d", env[parameters["cidr"]],
			"-j", "REDIRECT",
			"!", "-s", "127.0.0.1/32",
			"--to-port", env[parameters["port"]],
			"-m", "owner", "!", "--uid-owner", envoyUID,
			"-m", "comment", "--comment",
			fmt.Sprintf("\"lattice redirect %v traffic to envoy\"", protocol),
		}
		err = ipt.Append(tableNAT, chainOutput, rulespecs...)
		if err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// XDS V1 Config
// -----------------------------------------------------------------------------

type XDSV1BootstrapConfig struct {
	Listeners      []string            `json:"listeners"`
	LDS            XDSV1LDS            `json:"lds"`
	Admin          XDSV1Admin          `json:"admin"`
	ClusterManager XDSV1ClusterManager `json:"cluster_manager"`
}

type XDSV1LDS struct {
	Cluster        string `json:"cluster"`
	RefreshDelayMS int    `json:"refresh_delay_ms"`
}

type XDSV1Admin struct {
	AccessLogPath string `json:"access_log_path"`
	Address       string `json:"address"`
}

type XDSV1ClusterManager struct {
	Clusters []XDSV1Cluster `json:"clusters"`
	CDS      XDSV1CDS       `json:"cds"`
	SDS      XDSV1SDS       `json:"sds"`
}

type XDSV1CDS struct {
	Cluster        XDSV1Cluster `json:"cluster"`
	RefreshDelayMS int          `json:"refresh_delay_ms"`
}

type XDSV1SDS struct {
	Cluster        XDSV1Cluster `json:"cluster"`
	RefreshDelayMS int          `json:"refresh_delay_ms"`
}

type XDSV1Cluster struct {
	Name             string      `json:"name"`
	ConnectTimeoutMS int         `json:"connect_timeout_ms"`
	Type             string      `json:"type"`
	LBType           string      `json:"lb_type"`
	Hosts            []XDSV1Host `json:"hosts"`
}

type XDSV1Host struct {
	URL string `json:"url"`
}

// -----------------------------------------------------------------------------
// XDS V2 Config
// -----------------------------------------------------------------------------

type XDSV2BootstrapConfig struct {
	Node             XDSV2Node             `json:"node"`
	Admin            XDSV2Admin            `json:"admin"`
	StaticResources  XDSV2StaticResources  `json:"static_resources"`
	DynamicResources XDSV2DynamicResources `json:"dynamic_resources"`
}

type XDSV2Node struct {
	Id      string `json:"id"`
	Cluster string `json:"cluster"`
}

type XDSV2Admin struct {
	AccessLogPath string       `json:"access_log_path"`
	Address       XDSV2Address `json:"address"`
}

type XDSV2StaticResources struct {
	Clusters []XDSV2Cluster `json:"clusters"`
}

type XDSV2Cluster struct {
	Name                 string      `json:"name"`
	ConnectTimeout       string      `json:"connect_timeout"`
	Type                 string      `json:"type"`
	LBPolicy             string      `json:"lb_policy"`
	HTTP2ProtocolOptions struct{}    `json:"http2_protocol_options"`
	Hosts                []XDSV2Host `json:"hosts"`
}

// XXX: add extra fields to support non-ads configuration

type XDSV2DynamicResources struct {
	ADSConfig XDSV2ADSConfig `json:"ads_config"`
	LDSConfig XDSV2LDSConfig `json:"lds_config"`
	CDSConfig XDSV2CDSConfig `json:"cds_config"`
}

type XDSV2ADSConfig struct {
	APIType      string             `json:"api_type"`
	GRPCServices []XDSV2GRPCService `json:"grpc_services"`
}

type XDSV2GRPCService struct {
	EnvoyGRPC XDSV2EnvoyGRPC `json:"envoy_grpc"`
}

type XDSV2EnvoyGRPC struct {
	ClusterName string `json:"cluster_name"`
}

type XDSV2LDSConfig struct {
	ADS struct{} `json:"ads"`
}

type XDSV2CDSConfig struct {
	ADS struct{} `json:"ads"`
}

type XDSV2Address struct {
	SocketAddress XDSV2SocketAddress `json:"socket_address"`
}

type XDSV2Host struct {
	SocketAddress XDSV2SocketAddress `json:"socket_address"`
}

type XDSV2SocketAddress struct {
	Address   string `json:"address"`
	PortValue int    `json:"port_value"`
}

func outputEnvoyConfig(env map[string]string) error {
	if err := os.MkdirAll(env[envVarConfigDir], 0644); err != nil {
		return err
	}

	configFilename := filepath.Join(env[envVarConfigDir], "config.json")

	var contents []byte
	var err error

	adminPort, err := strconv.Atoi(env[envVarAdminPort])
	xdsAPIPort, err := strconv.Atoi(env[envVarXDSAPIPort])
	xdsAPIURL := fmt.Sprintf("%v:%v", env[envVarXDSAPIHost], xdsAPIPort)

	switch env[envVarXDSAPIVersion] {
	case "1":
		contents, err = json.MarshalIndent(XDSV1BootstrapConfig{
			Listeners: []string{},
			LDS: XDSV1LDS{
				Cluster:        DefaultXDSClusterName,
				RefreshDelayMS: DefaultXDSClusterRefreshDelayMS,
			},
			Admin: XDSV1Admin{
				AccessLogPath: "/dev/null",
				Address:       fmt.Sprintf("tcp://0.0.0.0:%v", env[envVarAdminPort]),
			},
			ClusterManager: XDSV1ClusterManager{
				Clusters: []XDSV1Cluster{
					{
						Name:             DefaultXDSClusterName,
						ConnectTimeoutMS: DefaultXDSClusterConnectTimeoutMS,
						Type:             "static",
						LBType:           "round_robin",
						Hosts: []XDSV1Host{
							{
								URL: fmt.Sprintf("tcp://%v", xdsAPIURL),
							},
						},
					},
				},
				CDS: XDSV1CDS{
					Cluster: XDSV1Cluster{
						Name:             fmt.Sprintf("%v-cds", DefaultXDSClusterName),
						ConnectTimeoutMS: DefaultXDSClusterConnectTimeoutMS,
						Type:             "static",
						LBType:           "round_robin",
						Hosts: []XDSV1Host{
							{
								URL: fmt.Sprintf("tcp://%v", xdsAPIURL),
							},
						},
					},
					RefreshDelayMS: DefaultXDSClusterRefreshDelayMS,
				},
				SDS: XDSV1SDS{
					Cluster: XDSV1Cluster{
						Name:             fmt.Sprintf("%v-sds", DefaultXDSClusterName),
						ConnectTimeoutMS: DefaultXDSClusterConnectTimeoutMS,
						Type:             "static",
						LBType:           "round_robin",
						Hosts: []XDSV1Host{
							{
								URL: fmt.Sprintf("tcp://%v", xdsAPIURL),
							},
						},
					},
					RefreshDelayMS: DefaultXDSClusterRefreshDelayMS,
				},
			},
		}, "", "  ")
	case "2":
		contents, err = json.MarshalIndent(XDSV2BootstrapConfig{
			Node: XDSV2Node{
				Id:      env[envVarServiceNode],
				Cluster: env[envVarServiceCluster],
			},
			Admin: XDSV2Admin{
				AccessLogPath: "/dev/null",
				Address: XDSV2Address{
					SocketAddress: XDSV2SocketAddress{
						Address:   "0.0.0.0",
						PortValue: adminPort,
					},
				},
			},
			StaticResources: XDSV2StaticResources{
				Clusters: []XDSV2Cluster{
					{
						Name:                 DefaultXDSClusterName,
						ConnectTimeout:       DefaultXDSClusterConnectTimeout,
						Type:                 "STATIC",
						LBPolicy:             "ROUND_ROBIN",
						HTTP2ProtocolOptions: struct{}{},
						Hosts: []XDSV2Host{
							{
								SocketAddress: XDSV2SocketAddress{
									Address:   env[envVarXDSAPIHost],
									PortValue: xdsAPIPort,
								},
							},
						},
					},
				},
			},
			DynamicResources: XDSV2DynamicResources{
				ADSConfig: XDSV2ADSConfig{
					APIType: "GRPC",
					GRPCServices: []XDSV2GRPCService{
						{
							EnvoyGRPC: XDSV2EnvoyGRPC{
								ClusterName: DefaultXDSClusterName,
							},
						},
					},
				},
				LDSConfig: XDSV2LDSConfig{
					ADS: struct{}{},
				},
				CDSConfig: XDSV2CDSConfig{
					ADS: struct{}{},
				},
			},
		}, "", "  ")
	default:
		err = fmt.Errorf("unknown envoy boostrap config version: %v", env[envVarXDSAPIVersion])
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFilename, contents, 0644)
}
