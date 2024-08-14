package parameters

import "github.com/weaviate/weaviate/entities/modulecapabilities"

const name = "unbody"

func AdditionalGenerativeParameters(client modulecapabilities.GenerativeFlexClient) map[string]modulecapabilities.GenerativeFlexProperty {
	return map[string]modulecapabilities.GenerativeFlexProperty{
		name:       {Client: client, RequestParamsFunction: input, ResponseParamsFunction: output, ExtractRequestParamsFunction: extract},
		"options":  {Client: client, RequestParamsFunction: input, ResponseParamsFunction: output, ExtractRequestParamsFunction: extract},
		"metadata": {Client: client, RequestParamsFunction: input, ResponseParamsFunction: output, ExtractRequestParamsFunction: extract},
	}
}
