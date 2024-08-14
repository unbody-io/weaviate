//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2024 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package additional

import (
	"github.com/sirupsen/logrus"

	"github.com/weaviate/weaviate/entities/modulecapabilities"
	generativegenerateflex "github.com/weaviate/weaviate/usecases/modulecomponents/additional/generateflex"
)

type GraphQLAdditionalGenerativeFlexProvider struct {
	generative AdditionalProperty
}

func NewGenericGenerativeFlexProvider(
	className string,
	additionalGenerativeFlexParameters map[string]modulecapabilities.GenerativeFlexProperty,
	defaultProviderName string,
	logger logrus.FieldLogger,
) *GraphQLAdditionalGenerativeFlexProvider {
	return &GraphQLAdditionalGenerativeFlexProvider{generativegenerateflex.NewGeneric(additionalGenerativeFlexParameters, defaultProviderName, logger)}
}

func (p *GraphQLAdditionalGenerativeFlexProvider) AdditionalProperties() map[string]modulecapabilities.AdditionalProperty {
	additionalProperties := map[string]modulecapabilities.AdditionalProperty{}
	additionalProperties[PropertyGenerate] = p.getGenerate()
	return additionalProperties
}

func (p *GraphQLAdditionalGenerativeFlexProvider) getGenerate() modulecapabilities.AdditionalProperty {
	return modulecapabilities.AdditionalProperty{
		GraphQLNames:           []string{PropertyGenerate},
		GraphQLFieldFunction:   p.generative.AdditionalFieldFn,
		GraphQLExtractFunction: p.generative.ExtractAdditionalFn,
		SearchFunctions: modulecapabilities.AdditionalSearch{
			ExploreGet:  p.generative.AdditionalPropertyFn,
			ExploreList: p.generative.AdditionalPropertyFn,
		},
	}
}
