package generateflex

import (
	"fmt"

	"github.com/tailor-inc/graphql"
)

var GenerativeMessageType = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "GenerativeMessageType",
	Fields: graphql.InputObjectConfigFieldMap{
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"role": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"name": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"content": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
})

func (p *GenerateFlexProvider) additionalGenerateField(className string) *graphql.Field {
	generate := &graphql.Field{
		Args: graphql.FieldConfigArgument{
			"singleResult": &graphql.ArgumentConfig{
				Description: "Results per object",
				Type: graphql.NewInputObject(graphql.InputObjectConfig{
					Name:   fmt.Sprintf("%sIndividualResultsArg", className),
					Fields: p.singleResultArguments(className),
				}),
				DefaultValue: nil,
			},
			"groupedResult": &graphql.ArgumentConfig{
				Description: "Grouped results of all objects",
				Type: graphql.NewInputObject(graphql.InputObjectConfig{
					Name:   fmt.Sprintf("%sAllResultsArg", className),
					Fields: p.groupedResultArguments(className),
				}),
				DefaultValue: nil,
			},
		},
		Type: graphql.NewObject(graphql.ObjectConfig{
			Name:   fmt.Sprintf("%sAdditionalGenerate", className),
			Fields: p.fields(className),
		}),
	}
	return generate
}

func (p *GenerateFlexProvider) singleResultArguments(className string) graphql.InputObjectConfigFieldMap {
	argumentFields := graphql.InputObjectConfigFieldMap{
		"prompt": &graphql.InputObjectFieldConfig{
			Description: "prompt",
			Type:        graphql.String,
		},
		"messages": {
			Type: graphql.NewList(GenerativeMessageType),
		},
		"debug": &graphql.InputObjectFieldConfig{
			Description: "debug",
			Type:        graphql.Boolean,
		},
	}
	p.inputArguments(argumentFields, fmt.Sprintf("%sSingleResult", className))
	return argumentFields
}

func (p *GenerateFlexProvider) groupedResultArguments(className string) graphql.InputObjectConfigFieldMap {
	argumentFields := graphql.InputObjectConfigFieldMap{
		"task": &graphql.InputObjectFieldConfig{
			Description: "task",
			Type:        graphql.String,
		},
		"properties": &graphql.InputObjectFieldConfig{
			Description:  "Properties used for the generation",
			Type:         graphql.NewList(graphql.String),
			DefaultValue: nil,
		},
		"messages": {
			Type: graphql.NewList(GenerativeMessageType),
		},
		"debug": &graphql.InputObjectFieldConfig{
			Description: "debug",
			Type:        graphql.Boolean,
		},
	}
	p.inputArguments(argumentFields, fmt.Sprintf("%sGroupedResult", className))
	return argumentFields
}

func (p *GenerateFlexProvider) inputArguments(argumentFields graphql.InputObjectConfigFieldMap, prefix string) {
	for _, generativeParameters := range p.additionalGenerativeParameters {
		if generativeParameters.RequestParamsFunction != nil {
			argumentFields["options"] = generativeParameters.RequestParamsFunction(prefix)
		}
	}
}

func (p *GenerateFlexProvider) fields(className string) graphql.Fields {
	fields := graphql.Fields{
		"singleResult":  &graphql.Field{Type: graphql.String},
		"groupedResult": &graphql.Field{Type: graphql.String},
		"error":         &graphql.Field{Type: graphql.String},
		"debug": &graphql.Field{Type: graphql.NewObject(graphql.ObjectConfig{
			Name: fmt.Sprintf("%sDebugFields", className),
			Fields: graphql.Fields{
				"prompt": &graphql.Field{Type: graphql.String},
			},
		})},
	}
	for _, generativeParameters := range p.additionalGenerativeParameters {
		if generativeParameters.ResponseParamsFunction != nil {
			fields["metadata"] = generativeParameters.ResponseParamsFunction(className)
		}
	}
	return fields
}
