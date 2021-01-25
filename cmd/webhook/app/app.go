package app

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
	"github.com/bind-dns/binddns-operator/pkg/webhook"
	"github.com/bind-dns/binddns-operator/version"
)

var (
	logFile         string
	listenPort      string
	gracefulTimeout int

	tlsCertFile string
	tlsKeyFile  string
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "binddns-webhook",
		Short: "binddns-webhook used to intercept and verify DnsDomain/DnsRule crd resources.",
		Run: func(cmd *cobra.Command, args []string) {
			// Init log zlog.
			zlog.DefaultLog(logFile)

			// Init kubeclient
			if err := kube.InitKubernetesClient(); err != nil {
				zlog.Error(err)
				return
			}

			server := webhook.NewAdmissionWebhookServer()
			server.ListenPort = listenPort
			server.ShutdownTimeout = int64(gracefulTimeout)
			server.TlsCertFile = tlsCertFile
			server.TlsKeyFile = tlsKeyFile
			if err := server.Run(); err != nil {
				zlog.Error(err)
			}
		},
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version detail info.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.String())
		},
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", fmt.Sprintf(utils.DefaultLogFile, webhook.AppName), "the log output filepath.")
	rootCmd.PersistentFlags().StringVar(&listenPort, "listen-port", ":8443", "the log output filepath.")
	rootCmd.PersistentFlags().IntVar(&gracefulTimeout, "graceful-shutdown-timeout", 20, "the max timeout of server graceful shutdown.")
	rootCmd.PersistentFlags().StringVar(&tlsCertFile, "tls-certfile", "/etc/webhook/certs/cert.pem", "the x509 Certificate of http server for HTTPS")
	rootCmd.PersistentFlags().StringVar(&tlsKeyFile, "tls-keyfile", "/etc/webhook/certs/key.pem", "the x509 private key of http server for HTTPS")
	return rootCmd
}
