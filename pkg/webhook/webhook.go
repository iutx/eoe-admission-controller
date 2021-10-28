package webhook

import (
	"context"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cluster_credential "github.com/iutx/eoe-admission-controller/pkg/webhook/cluster-credential"
)

const (
	listen   = 443
	certPath = "/run/eoe/tls"
)

var (
	localSchemeBuilder = runtime.SchemeBuilder{
		k8sscheme.AddToScheme,
	}
)

type AdmissionFunc func(ctx context.Context, request admission.Request) admission.Response

type Webhook struct {
	CRClient client.Client
	Decoder  *admission.Decoder
}

func New() (*Webhook, error) {
	w := &Webhook{}

	rc, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
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

	w.Decoder, err = admission.NewDecoder(sc)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Webhook) Start() error {
	logrus.Infof("start eoe webhook server at port: %d", listen)

	server := crwebhook.Server{
		Port:    listen,
		CertDir: certPath,
	}

	server.Register("/eoe/cluster-credential", &crwebhook.Admission{
		Handler: &cluster_credential.ValidatingHandler{
			CRClient: w.CRClient,
			Decoder:  w.Decoder,
		},
	})

	// Default inject kubernetes schema
	return server.StartStandalone(context.Background(), nil)
}
