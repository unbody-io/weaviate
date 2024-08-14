package parameters

import (
	"github.com/tailor-inc/graphql/language/ast"
	"github.com/weaviate/weaviate/usecases/modulecomponents/gqlparser"
)

type Params struct {
	Model            string
	Vars             *[]interface{}
	FrequencyPenalty *float64
	MaxTokens        *int
	PresencePenalty  *float64
	Temperature      *float64
	TopP             *float64
}

func extract(field *ast.ObjectField) interface{} {
	out := Params{}
	fields, ok := field.Value.GetValue().([]*ast.ObjectField)
	if ok {
		for _, f := range fields {
			switch f.Name.Value {
			case "model":
				out.Model = gqlparser.GetValueAsStringOrEmpty(f)
			case "frequencyPenalty":
				out.FrequencyPenalty = gqlparser.GetValueAsFloat64(f)
			case "maxTokens":
				out.MaxTokens = gqlparser.GetValueAsInt(f)
			case "presencePenalty":
				out.PresencePenalty = gqlparser.GetValueAsFloat64(f)
			case "temperature":
				out.Temperature = gqlparser.GetValueAsFloat64(f)
			case "topP":
				out.TopP = gqlparser.GetValueAsFloat64(f)

			case "vars":
				values := make([]interface{}, len(f.Value.(*ast.ListValue).Values))
				for i, value := range f.Value.(*ast.ListValue).Values {
					values[i] = map[string]interface{}{}

					for _, sf := range value.(*ast.ObjectValue).Fields {
						switch sf.Name.Value {
						case "name":
							values[i].(map[string]interface{})["name"] = sf.Value.(*ast.StringValue).Value
						case "expression":
							values[i].(map[string]interface{})["expression"] = sf.Value.(*ast.StringValue).Value
						case "formatter":
							values[i].(map[string]interface{})["formatter"] = sf.Value.(*ast.StringValue).Value
						}
					}
				}
				out.Vars = &values

			default:
				// do nothing
			}
		}
	}
	return out
}
