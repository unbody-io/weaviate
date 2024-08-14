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

package parameters

import (
	"fmt"

	"github.com/tailor-inc/graphql"
)

func VariableType(prefix string) *graphql.InputObject {
	return graphql.NewInputObject(graphql.InputObjectConfig{
		Name: fmt.Sprintf("%sGenerativeVariableType", prefix),
		Fields: graphql.InputObjectConfigFieldMap{
			"name": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"formatter": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"expression": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	})
}

func input(prefix string) *graphql.InputObjectFieldConfig {
	return &graphql.InputObjectFieldConfig{
		Description: fmt.Sprintf("%s settings", name),
		Type: graphql.NewInputObject(graphql.InputObjectConfig{
			Name: fmt.Sprintf("%s%sInputObject", prefix, name),
			Fields: graphql.InputObjectConfigFieldMap{
				"model": &graphql.InputObjectFieldConfig{
					Description: "model",
					Type:        graphql.String,
				},
				"frequencyPenalty": &graphql.InputObjectFieldConfig{
					Description: "frequencyPenalty",
					Type:        graphql.Float,
				},
				"maxTokens": &graphql.InputObjectFieldConfig{
					Description: "maxTokens",
					Type:        graphql.Int,
				},
				"presencePenalty": &graphql.InputObjectFieldConfig{
					Description: "presencePenalty",
					Type:        graphql.Float,
				},
				"temperature": &graphql.InputObjectFieldConfig{
					Description: "temperature",
					Type:        graphql.Float,
				},
				"topP": &graphql.InputObjectFieldConfig{
					Description: "topP",
					Type:        graphql.Float,
				},
				"vars": {
					Type:         graphql.NewList(VariableType(prefix)),
					DefaultValue: make([]interface{}, 0),
				},
			},
		}),
		DefaultValue: nil,
	}
}

func output(prefix string) *graphql.Field {
	return &graphql.Field{Type: graphql.NewObject(graphql.ObjectConfig{
		Name: fmt.Sprintf("%s%sFields", prefix, name),
		Fields: graphql.Fields{
			"finishReason": &graphql.Field{Type: graphql.String},
			"usage": &graphql.Field{Type: graphql.NewObject(graphql.ObjectConfig{
				Name: fmt.Sprintf("%s%sUsageMetadataFields", prefix, name),
				Fields: graphql.Fields{
					"inputTokens":  &graphql.Field{Type: graphql.Int},
					"outputTokens": &graphql.Field{Type: graphql.Int},
					"totalTokens":  &graphql.Field{Type: graphql.Int},
				},
			})},
		},
	})}
}
