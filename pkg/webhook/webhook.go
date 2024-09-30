package webhook

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	zaplogfmt "github.com/sykesm/zap-logfmt"
	uzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cluster_credential "github.com/iutx/eoe-admission-controller/pkg/webhook/patchers"
)

const (
	listen         = 443
	certPath       = "/run/eoe/tls"
	certCustomPath = "CERT_PATH"
	mode           = "MODE"
	devMode        = "DEV"
	devInsecure    = "DEV_INSECURE"
	apiServerAddr  = "APISERVER_ADDR"
	kubeconfigPath = "KUBECONFIG_PATH"
)

var (
	localSchemeBuilder = runtime.SchemeBuilder{
		k8sscheme.AddToScheme,
	}
)

type AdmissionFunc func(ctx context.Context, request admission.Request) admission.Response

type Webhook struct {
	CRClient client.Client
	Decoder  admission.Decoder
}

func New() (*Webhook, error) {
	var (
		w   = &Webhook{}
		rc  *rest.Config
		err error
	)

	leveler := uzap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.DPanicLevel
	})
	stackTraceLeveler := uzap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return false
	})
	logfmtEncoder := zaplogfmt.NewEncoder(uzap.NewProductionEncoderConfig())
	logger := zap.New(
		zap.Level(leveler),
		zap.StacktraceLevel(stackTraceLeveler),
		zap.UseDevMode(false),
		zap.WriteTo(os.Stdout),
		zap.Encoder(logfmtEncoder))
	log.SetLogger(logger)

	if os.Getenv(mode) == devMode {
		rc, err = clientcmd.BuildConfigFromFlags(os.Getenv(apiServerAddr), os.Getenv(kubeconfigPath))
		if err != nil {
			return nil, err
		}

		if os.Getenv(devInsecure) == "true" {
			rc.TLSClientConfig.Insecure = true
		}
	} else {
		rc, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	sc := runtime.NewScheme()
	schemeBuilder := &runtime.SchemeBuilder{}

	for _, s := range localSchemeBuilder {
		schemeBuilder.Register(s)
	}

	if err = schemeBuilder.AddToScheme(sc); err != nil {
		return nil, err
	}

	if crc, err := client.New(rc, client.Options{Scheme: sc}); err != nil {
		return nil, err
	} else {
		w.CRClient = crc
	}

	w.Decoder = admission.NewDecoder(sc)
	return w, nil
}

func (w *Webhook) Start() error {
	logrus.Infof("start eoe webhook server at port: %d", listen)

	options := crwebhook.Options{
		Port:    listen,
		CertDir: certPath,
	}

	if os.Getenv(certCustomPath) != "" {
		options.CertDir = os.Getenv(certCustomPath)
	}

	server := crwebhook.NewServer(options)
	server.Register("/eoe/patches", &crwebhook.Admission{
		Handler: &cluster_credential.MutatingWebhookHandler{
			CRClient: w.CRClient,
			Decoder:  w.Decoder,
		},
	})

	// Default inject kubernetes schema
	return server.Start(context.Background())
}
