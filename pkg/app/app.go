package app

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/pluralsh/helmify/pkg/file"
	"github.com/pluralsh/helmify/pkg/processor/job"
	"github.com/pluralsh/helmify/pkg/processor/statefulset"

	"github.com/sirupsen/logrus"

	"github.com/pluralsh/helmify/pkg/config"
	"github.com/pluralsh/helmify/pkg/decoder"
	"github.com/pluralsh/helmify/pkg/helm"
	"github.com/pluralsh/helmify/pkg/processor"
	"github.com/pluralsh/helmify/pkg/processor/configmap"
	"github.com/pluralsh/helmify/pkg/processor/crd"
	"github.com/pluralsh/helmify/pkg/processor/daemonset"
	"github.com/pluralsh/helmify/pkg/processor/deployment"
	"github.com/pluralsh/helmify/pkg/processor/rbac"
	"github.com/pluralsh/helmify/pkg/processor/secret"
	"github.com/pluralsh/helmify/pkg/processor/service"
	"github.com/pluralsh/helmify/pkg/processor/storage"
	"github.com/pluralsh/helmify/pkg/processor/webhook"
)

// Start - application entrypoint for processing input to a Helm chart.
func Start(stdin io.Reader, config config.Config) error {
	err := config.Validate()
	if err != nil {
		return err
	}
	setLogLevel(config)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		logrus.Debug("Received termination, signaling shutdown")
		cancelFunc()
	}()
	appCtx := New(config, helm.NewOutput())
	appCtx = appCtx.WithProcessors(
		configmap.New(),
		crd.New(),
		daemonset.New(),
		deployment.New(),
		statefulset.New(),
		storage.New(),
		service.New(),
		service.NewIngress(),
		rbac.ClusterRoleBinding(),
		rbac.Role(),
		rbac.RoleBinding(),
		rbac.ServiceAccount(),
		secret.New(),
		webhook.Issuer(),
		webhook.Certificate(),
		webhook.ValidatingWebhook(),
		webhook.MutatingWebhook(),
		job.NewCron(),
		job.NewJob(),
	).WithDefaultProcessor(processor.Default()).WithPreProcessor(processor.New())
	if len(config.Files) != 0 {
		file.Walk(config.Files, config.FilesRecursively, func(filename string, fileReader io.Reader) {
			objects := decoder.Decode(ctx.Done(), fileReader)
			for obj := range objects {
				appCtx.Add(obj, filename)
			}
		})
	} else {
		objects := decoder.Decode(ctx.Done(), stdin)
		for obj := range objects {
			appCtx.Add(obj, "")
		}
	}

	return appCtx.CreateHelm(ctx.Done())
}

func setLogLevel(config config.Config) {
	logrus.SetLevel(logrus.ErrorLevel)
	if config.Verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if config.VeryVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
