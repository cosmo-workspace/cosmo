package kosmo

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Client struct {
	client.Client
}

func NewClient(c client.Client) Client {
	return Client{Client: c}
}

func NewClientByRestConfig(cfg *rest.Config, scheme *runtime.Scheme) (Client, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return Client{}, err
	}

	clientOptions := client.Options{Scheme: scheme, Mapper: mapper}
	client, err := client.New(cfg, clientOptions)
	if err != nil {
		return Client{}, err
	}

	return NewClient(client), nil
}
