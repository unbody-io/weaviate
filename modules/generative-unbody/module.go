package modgenerativeunbody

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/modulecapabilities"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/modules/generative-unbody/clients"
	"github.com/weaviate/weaviate/modules/generative-unbody/parameters"
)

const Name = "generative-unbody"

func New() *GenerativeUnbodyModule {
	return &GenerativeUnbodyModule{}
}

type GenerativeUnbodyModule struct {
	generative                   generativeClient
	additionalPropertiesProvider map[string]modulecapabilities.GenerativeFlexProperty
}

type generativeClient interface {
	modulecapabilities.GenerativeFlexClient
	MetaInfo() (map[string]interface{}, error)
}

func (m *GenerativeUnbodyModule) Name() string {
	return Name
}

func (m *GenerativeUnbodyModule) Type() modulecapabilities.ModuleType {
	return modulecapabilities.Text2TextGenerative
}

func (m *GenerativeUnbodyModule) Init(ctx context.Context,
	params moduletools.ModuleInitParams,
) error {
	if err := m.initAdditional(ctx, params.GetConfig().ModuleHttpClientTimeout, params.GetLogger()); err != nil {
		return errors.Wrapf(err, "init %s", Name)
	}
	return nil
}

func (m *GenerativeUnbodyModule) initAdditional(ctx context.Context, timeout time.Duration,
	logger logrus.FieldLogger,
) error {
	apiKey := os.Getenv("UNBODY_API_KEY")
	projectId := os.Getenv("UNBODY_PROJECT_ID")
	baseURL := os.Getenv("UNBODY_GENERATIVE_BASE_URL")

	client := clients.New(baseURL, projectId, apiKey, timeout, logger)
	m.generative = client
	m.additionalPropertiesProvider = parameters.AdditionalGenerativeParameters(m.generative)

	return nil
}

func (m *GenerativeUnbodyModule) RootHandler() http.Handler {
	// TODO: remove once this is a capability interface
	return nil
}

func (m *GenerativeUnbodyModule) MetaInfo() (map[string]interface{}, error) {
	return m.generative.MetaInfo()
}

func (m *GenerativeUnbodyModule) AdditionalGenerativeProperties() map[string]modulecapabilities.GenerativeFlexProperty {
	return m.additionalPropertiesProvider
}

// verify we implement the modules.Module interface
var (
	_ = modulecapabilities.Module(New())
	_ = modulecapabilities.MetaProvider(New())
	_ = modulecapabilities.AdditionalGenerativeFlexProperties(New())
)
