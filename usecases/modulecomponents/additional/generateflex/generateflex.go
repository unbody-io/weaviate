package generateflex

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/tailor-inc/graphql"
	"github.com/tailor-inc/graphql/language/ast"
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/modulecapabilities"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/entities/search"
)

const maximumNumberOfGoroutines = 10

type GenerateFlexProvider struct {
	additionalGenerativeParameters map[string]modulecapabilities.GenerativeFlexProperty
	defaultProviderName            string
	maximumNumberOfGoroutines      int
	logger                         logrus.FieldLogger
}

func NewGeneric(
	additionalGenerativeParameters map[string]modulecapabilities.GenerativeFlexProperty,
	defaultProviderName string,
	logger logrus.FieldLogger,
) *GenerateFlexProvider {
	return &GenerateFlexProvider{
		additionalGenerativeParameters: additionalGenerativeParameters,
		defaultProviderName:            defaultProviderName,
		maximumNumberOfGoroutines:      maximumNumberOfGoroutines,
		logger:                         logger,
	}
}

func (p *GenerateFlexProvider) AdditionalPropertyDefaultValue() interface{} {
	return &Params{}
}

func (p *GenerateFlexProvider) ExtractAdditionalFn(param []*ast.Argument, class *models.Class) interface{} {
	return p.parseGenerateArguments(param)
}

func (p *GenerateFlexProvider) AdditionalFieldFn(classname string) *graphql.Field {
	return p.additionalGenerateField(classname)
}

func (p *GenerateFlexProvider) AdditionalPropertyFn(ctx context.Context,
	in []search.Result, params interface{}, limit *int,
	argumentModuleParams map[string]interface{}, cfg moduletools.ClassConfig,
) ([]search.Result, error) {
	if parameters, ok := params.(*Params); ok {
		if len(parameters.Options) > 1 {
			var providerNames []string
			for name := range parameters.Options {
				providerNames = append(providerNames, name)
			}
			return nil, fmt.Errorf("multiple providers selected: %v, please choose only one", providerNames)
		}
		return p.generateResult(ctx, in, parameters, limit, argumentModuleParams, cfg)
	}
	return nil, errors.New("wrong parameters")
}
