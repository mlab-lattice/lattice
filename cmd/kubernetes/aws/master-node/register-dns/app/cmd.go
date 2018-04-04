package app

import (
	"fmt"
	"os"

	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/terraform/aws"
	tf "github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstf "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"

	"github.com/spf13/cobra"
)

var (
	workDirectory string

	terraformS3Bucket         string
	terraformS3KeyPrefix      string
	terraformModuleSourcePath string

	region string

	name                 string
	route53PrivateZoneID string
	instancePrivateIP    string
)

// Cmd represents the base command when called without any subcommands
var Cmd = &cobra.Command{
	Use:   "register-dns",
	Short: "Registers dns for a master node",
	Run: func(cmd *cobra.Command, args []string) {
		if err := apply(); err != nil {
			panic(err)
		}
	},
}

func apply() error {
	config := &tf.Config{
		Provider: awstf.Provider{
			Region: region,
		},
		Backend: tf.S3BackendConfig{
			Region: region,
			Bucket: terraformS3Bucket,
			Key: fmt.Sprintf(
				"%v/dns/terraform.tfstate",
				terraformS3KeyPrefix,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"dns": kubetf.NewMasterNodeDNS(
				terraformModuleSourcePath,
				region,
				name,
				route53PrivateZoneID,
				instancePrivateIP,
			),
		},
	}

	logfile, err := tf.Apply(workDirectory, config)
	if err != nil {
		fmt.Printf("error applying, logfile: %v\n", logfile)
	}

	return err
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	Cmd.Flags().StringVar(&workDirectory, "work-directory", "", "")
	Cmd.Flags().StringVar(&terraformS3Bucket, "terraform-state-s3-bucket", "", "")
	Cmd.Flags().StringVar(&terraformS3KeyPrefix, "terraform-state-s3-key-prefix", "", "")
	Cmd.Flags().StringVar(&terraformModuleSourcePath, "terraform-module-source-path", "", "")
	Cmd.Flags().StringVar(&region, "region", "", "")
	Cmd.Flags().StringVar(&name, "name", "", "")
	Cmd.Flags().StringVar(&route53PrivateZoneID, "route53-private-zone-id", "", "")
	Cmd.Flags().StringVar(&instancePrivateIP, "instance-private-ip", "", "")
}
