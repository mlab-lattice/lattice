package app

import (
	"fmt"
	"os"
	"time"

	kubetf "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	awstf "github.com/mlab-lattice/system/pkg/terraform/provider/aws"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/spf13/cobra"
)

var (
	workDirectory string

	terraformS3Bucket         string
	terraformS3KeyPrefix      string
	terraformModuleSourcePath string

	region     string
	clusterID  string
	name       string
	instanceID string
	deviceName string
)

// Cmd represents the base command when called without any subcommands
var Cmd = &cobra.Command{
	Use:   "attach-etcd-volume",
	Short: "Attaches an etcd volume to a master node",
	Run: func(cmd *cobra.Command, args []string) {
		err := wait.ExponentialBackoff(
			wait.Backoff{
				Duration: 5 * time.Second,
				Factor:   2,
				Jitter:   0.5,
				Steps:    5,
			},
			apply,
		)

		if err != nil {
			panic(err)
		}
	},
}

func apply() (bool, error) {
	config := &tf.Config{
		Provider: awstf.Provider{
			Region: region,
		},
		Backend: tf.S3BackendConfig{
			Region: region,
			Bucket: terraformS3Bucket,
			Key: fmt.Sprintf(
				"%v/etcd-volume/terraform.tfstate",
				terraformS3KeyPrefix,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"etcd-volume-attachment": kubetf.NewMasterNodeEtcdVolumeAttachment(
				terraformModuleSourcePath,
				region,
				clusterID,
				name,
				instanceID,
				deviceName,
			),
		},
	}

	logfile, err := tf.Apply(workDirectory, config)
	if err != nil {
		fmt.Printf("error applying, logfile: %v\n", logfile)
		return false, nil
	}

	return true, nil
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
	Cmd.Flags().StringVar(&terraformS3Bucket, "s3-bucket", "", "")
	Cmd.Flags().StringVar(&terraformS3KeyPrefix, "s3-key-prefix", "", "")
	Cmd.Flags().StringVar(&terraformModuleSourcePath, "module-source-path", "", "")
	Cmd.Flags().StringVar(&region, "region", "", "")
	Cmd.Flags().StringVar(&clusterID, "cluster-id", "", "")
	Cmd.Flags().StringVar(&name, "name", "", "")
	Cmd.Flags().StringVar(&instanceID, "instance-id", "", "")
	Cmd.Flags().StringVar(&deviceName, "device-name", "", "")
}
