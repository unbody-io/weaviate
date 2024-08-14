package generateflex

import (
	"log"

	"github.com/tailor-inc/graphql/language/ast"
)

func (p *GenerateFlexProvider) parseGenerateArguments(args []*ast.Argument) *Params {
	out := &Params{Options: make(map[string]interface{})}

	out.SingleResult = false
	out.GroupedResult = false

	for _, arg := range args {
		switch arg.Name.Value {
		case "singleResult":
			out.SingleResult = true
			obj := arg.Value.(*ast.ObjectValue).Fields
			for _, field := range obj {
				switch field.Name.Value {
				case "prompt":
					out.Prompt = &field.Value.(*ast.StringValue).Value
				case "messages":
					values := make([]interface{}, len(field.Value.(*ast.ListValue).Values))
					for i, value := range field.Value.(*ast.ListValue).Values {
						values[i] = map[string]interface{}{}

						for _, field := range value.(*ast.ObjectValue).Fields {
							switch field.Name.Value {
							case "role":
								values[i].(map[string]interface{})["role"] = field.Value.(*ast.StringValue).Value
							case "content":
								values[i].(map[string]interface{})["content"] = field.Value.(*ast.StringValue).Value
							case "name":
								values[i].(map[string]interface{})["name"] = field.Value.(*ast.StringValue).Value
							case "type":
								values[i].(map[string]interface{})["type"] = field.Value.(*ast.StringValue).Value
							}
						}
					}
					out.Messages = &values
				case "debug":
					out.Debug = field.Value.(*ast.BooleanValue).Value
				default:
					if value := p.extractGenerativeParameter(field); value != nil {
						out.Options[field.Name.Value] = value
					}
				}
			}
		case "groupedResult":
			obj := arg.Value.(*ast.ObjectValue).Fields
			out.GroupedResult = true

			for _, field := range obj {
				switch field.Name.Value {
				case "task":
					out.Task = &field.Value.(*ast.StringValue).Value
				case "properties":
					inp := field.Value.GetValue().([]ast.Value)
					out.Properties = make([]string, len(inp))

					for i, value := range inp {
						out.Properties[i] = value.(*ast.StringValue).Value
					}
				case "messages":
					values := make([]interface{}, len(field.Value.(*ast.ListValue).Values))
					for i, value := range field.Value.(*ast.ListValue).Values {
						values[i] = map[string]interface{}{}

						for _, field := range value.(*ast.ObjectValue).Fields {
							switch field.Name.Value {
							case "role":
								values[i].(map[string]interface{})["role"] = field.Value.(*ast.StringValue).Value
							case "content":
								values[i].(map[string]interface{})["content"] = field.Value.(*ast.StringValue).Value
							case "name":
								values[i].(map[string]interface{})["name"] = field.Value.(*ast.StringValue).Value
							case "type":
								values[i].(map[string]interface{})["type"] = field.Value.(*ast.StringValue).Value
							}
						}
					}
					out.Messages = &values
				case "debug":
					out.Debug = field.Value.(*ast.BooleanValue).Value
				default:
					if value := p.extractGenerativeParameter(field); value != nil {
						out.Options[field.Name.Value] = value
					}
				}
			}

		default:
			// ignore what we don't recognize
			log.Printf("Igonore not recognized value: %v", arg.Name.Value)
		}
	}

	return out
}

func (p *GenerateFlexProvider) extractGenerativeParameter(field *ast.ObjectField) interface{} {
	if len(p.additionalGenerativeParameters) > 0 {
		if generative, ok := p.additionalGenerativeParameters[field.Name.Value]; ok {
			if extractFn := generative.ExtractRequestParamsFunction; extractFn != nil {
				return extractFn(field)
			}
		}
	}
	return nil
}
