package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/jakoblorz/autofone/pkg/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	verbose bool
	mac     string

	storageClient *storage.Client
	storageBucket *storage.BucketHandle
	projectID     = "autofone-355408"

	rootCmd = &cobra.Command{
		Use: "autofone",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config := zap.NewProductionConfig()
			if verbose {
				config = zap.NewDevelopmentConfig()
			}
			log.DefaultLogger, _ = config.Build()

			macAddresses, err := getMacAddr()
			if err != nil {
				log.Print(err)
				return
			}
			if len(macAddresses) == 0 {
				log.Print(errors.New("no mac address found"))
				return
			}
			mac = strings.ReplaceAll(macAddresses[0], ":", "")
			log.Printf("Using MAC Address %s for identification", mac)

			storageClient, err = storage.NewClient(context.Background())
			if err != nil {
				log.Print(err)
				return
			}

			storageBucket = storageClient.Bucket(mac)
			if err := storageBucket.Create(context.Background(), projectID, &storage.BucketAttrs{
				Location: "europe-west3",
			}); err != nil {
				log.Print(err)
				return
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output onto stdout")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}
