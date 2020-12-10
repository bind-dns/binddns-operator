package app

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bind-dns/binddns-operator/pkg/operator"
	"github.com/bind-dns/binddns-operator/pkg/signals"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
	"github.com/bind-dns/binddns-operator/version"
)

var (
	logFile       string
	logMaxSize    int
	logMaxBackups int
	logMaxAge     int
	logCompress   bool
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "run",
		Short: "Start the application.",
		Run: func(cmd *cobra.Command, args []string) {
			// Handle shutdown signals.
			stopCh := signals.SetupSignalHandler()
			// Init log formatter
			zlog.InitLog(logFile, logMaxSize, logMaxBackups, logMaxAge, logCompress)

			o, err := operator.NewDnsOperator()
			if err != nil {
				zlog.Error(err)
				return
			}

			// Start informer
			o.KubeInformerFactory.Start(stopCh)

			if err = o.Run(); err != nil {
				zlog.Panic(err)
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
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", fmt.Sprintf(utils.DefaultLogFile, operator.AppName), "the log output filepath.")
	rootCmd.PersistentFlags().IntVar(&logMaxSize, "log-max-size", utils.DefaultLogMaxSize, "the logfile max-size per file, unit (M).")
	rootCmd.PersistentFlags().IntVar(&logMaxBackups, "log-max-backups", utils.DefaultLogMaxBackups, "max num of the logfiles.")
	rootCmd.PersistentFlags().IntVar(&logMaxAge, "log-max-age", utils.DefaultLogMaxAge, "max age of the logfiles.")
	rootCmd.PersistentFlags().BoolVar(&logCompress, "log-compressed", utils.DefaultLogCompress, "whether the logfile need compress.")
	return rootCmd
}
