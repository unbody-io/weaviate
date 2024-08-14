package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/weaviate/weaviate/usecases/modulecomponents"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/entities/modulecapabilities"
	"github.com/weaviate/weaviate/entities/moduletools"
	"github.com/weaviate/weaviate/entities/search"
	"github.com/weaviate/weaviate/modules/generative-unbody/config"
	unbodyparams "github.com/weaviate/weaviate/modules/generative-unbody/parameters"
)

var compile, _ = regexp.Compile(`{([\w\s]*?)}`)

func buildUrlFn(baseURL string) (string, error) {
	path := "/chat/completions"
	return url.JoinPath(baseURL, path)
}

type unbody struct {
	baseURL    string
	projectId  string
	apiKey     string
	buildUrl   func(baseURL string) (string, error)
	httpClient *http.Client
	logger     logrus.FieldLogger
}

func New(baseURL string, projectId string, apiKey string, timeout time.Duration, logger logrus.FieldLogger) *unbody {
	return &unbody{
		baseURL:   baseURL,
		projectId: projectId,
		apiKey:    apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		buildUrl: buildUrlFn,
		logger:   logger,
	}

}

func (v *unbody) generateForPrompt(textProperties map[string]string, prompt string) (string, error) {
	all := compile.FindAll([]byte(prompt), -1)
	for _, match := range all {
		originalProperty := string(match)
		replacedProperty := compile.FindStringSubmatch(originalProperty)[1]
		replacedProperty = strings.TrimSpace(replacedProperty)
		value := textProperties[replacedProperty]
		if value == "" {
			return "", errors.Errorf("Following property has empty value: '%v'. Make sure you spell the property name correctly, verify that the property exists and has a value", replacedProperty)
		}
		prompt = strings.ReplaceAll(prompt, originalProperty, value)
	}
	return prompt, nil
}

func (v *unbody) generatePromptForTask(textProperties []map[string]string, task string) (string, error) {
	marshal, err := json.Marshal(textProperties)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`'%v:
%v`, task, string(marshal)), nil
}

func (v *unbody) GenerateSingleResult(ctx context.Context, textProperties map[string]string, prompt string, options interface{}, debug bool, cfg moduletools.ClassConfig) (*modulecapabilities.GenerateFlexResponse, error) {
	forPrompt, err := v.generateForPrompt(textProperties, prompt)

	if err != nil {
		return nil, err
	}

	return v.Generate(ctx, cfg, forPrompt, options, debug)
}

func (v *unbody) GenerateAllResults(ctx context.Context, textProperties []map[string]string, task string, options interface{}, debug bool, cfg moduletools.ClassConfig) (*modulecapabilities.GenerateFlexResponse, error) {
	prompt, err := v.generatePromptForTask(textProperties, task)

	if err != nil {
		return nil, err
	}

	return v.Generate(ctx, cfg, prompt, options, debug)
}

func (v *unbody) GenerateSingleResultWithMessages(ctx context.Context, result interface{}, messages []interface{}, options interface{}, debug bool, cfg moduletools.ClassConfig) (*modulecapabilities.GenerateFlexResponse, error) {
	return v.GenerateWithMessages(ctx, cfg, messages, result, nil, options, debug)
}

func (v *unbody) GenerateAllResultsWithMessages(ctx context.Context, result []interface{}, messages []interface{}, options interface{}, debug bool, cfg moduletools.ClassConfig) (*modulecapabilities.GenerateFlexResponse, error) {
	return v.GenerateWithMessages(ctx, cfg, messages, result, nil, options, debug)
}

func (v *unbody) Generate(ctx context.Context, cfg moduletools.ClassConfig, prompt string, options interface{}, debug bool) (*modulecapabilities.GenerateFlexResponse, error) {
	settings := config.NewClassSettings(cfg)
	params := v.getParameters(cfg, options)

	unbodyUrl, err := v.buildUnbodyUrl(ctx, settings)
	if err != nil {
		return nil, errors.Wrap(err, "url join path")
	}

	input, err := v.generateInput(prompt, params, settings)
	if err != nil {
		return nil, errors.Wrap(err, "generate input")
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Wrap(err, "marshal body")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", unbodyUrl,
		bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "create POST request")
	}
	apiKey, err := v.getApiKey(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Unbody API Key")
	}

	projectId, err := v.getProjectId(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Unbody Project ID")
	}

	req.Header.Add(v.getApiKeyHeaderAndValue(apiKey))
	req.Header.Add(v.getProjectIdHeaderAndValue(projectId))

	req.Header.Add("Content-Type", "application/json")

	res, err := v.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send POST request")
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	var resBody generateResponse
	if err := json.Unmarshal(bodyBytes, &resBody); err != nil {
		return nil, errors.Wrap(err, "unmarshal response body")
	}

	if res.StatusCode != 200 {
		apiError := &unbodyApiError{
			Message: resBody.Message,
			Errors:  resBody.Errors,
		}

		return nil, v.getError(res.StatusCode, apiError)
	}

	textResponse := resBody.Data.Content
	if len(textResponse) > 0 && textResponse != "" {
		trimmedResponse := strings.Trim(textResponse, "\n")
		return &modulecapabilities.GenerateFlexResponse{
			Result: &trimmedResponse,
			Params: map[string]interface{}{
				"options": map[string]interface{}{
					"Usage": resBody.Data.UsageMetadata,
				},
			},
		}, nil
	}

	return &modulecapabilities.GenerateFlexResponse{
		Result: nil,
	}, nil
}

func (v *unbody) GenerateWithMessages(ctx context.Context, cfg moduletools.ClassConfig, messages []interface{}, data interface{}, vars []interface{}, options interface{}, debug bool) (*modulecapabilities.GenerateFlexResponse, error) {
	settings := config.NewClassSettings(cfg)
	params := v.getParameters(cfg, options)

	unbodyUrl, err := v.buildUnbodyUrl(ctx, settings)
	if err != nil {
		return nil, errors.Wrap(err, "url join path")
	}

	input, err := v.generateInputWithMessages(messages, data, params, settings)
	if err != nil {
		return nil, errors.Wrap(err, "generate input")
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Wrap(err, "marshal body")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", unbodyUrl,
		bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "create POST request")
	}
	apiKey, err := v.getApiKey(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Unbody API Key")
	}

	projectId, err := v.getProjectId(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Unbody Project ID")
	}

	req.Header.Add(v.getApiKeyHeaderAndValue(apiKey))
	req.Header.Add(v.getProjectIdHeaderAndValue(projectId))

	req.Header.Add("Content-Type", "application/json")

	res, err := v.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "send POST request")
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	var resBody generateResponse
	if err := json.Unmarshal(bodyBytes, &resBody); err != nil {
		return nil, errors.Wrap(err, "unmarshal response body")
	}

	if res.StatusCode != 200 {
		apiError := &unbodyApiError{
			Message: resBody.Message,
			Errors:  resBody.Errors,
		}

		return nil, v.getError(res.StatusCode, apiError)
	}

	textResponse := resBody.Data.Content
	if len(textResponse) > 0 && textResponse != "" {
		trimmedResponse := strings.Trim(textResponse, "\n")
		return &modulecapabilities.GenerateFlexResponse{
			Result: &trimmedResponse,
			Params: map[string]interface{}{
				"metadata": map[string]interface{}{
					"finishReason": resBody.Data.FinishReason,
					"usage":        resBody.Data.UsageMetadata,
				},
			},
		}, nil
	}

	return &modulecapabilities.GenerateFlexResponse{
		Result: nil,
	}, nil
}

func (v *unbody) getParameters(cfg moduletools.ClassConfig, options interface{}) unbodyparams.Params {
	var params unbodyparams.Params
	if p, ok := options.(unbodyparams.Params); ok {
		params = p
	}

	return params
}

func (v *unbody) getDebugInformation(debug bool, prompt string) *modulecapabilities.GenerateDebugInformation {
	if debug {
		return &modulecapabilities.GenerateDebugInformation{
			Prompt: prompt,
		}
	}
	return nil
}

func (v *unbody) buildUnbodyUrl(ctx context.Context, settings config.ClassSettings) (string, error) {
	baseURL, err := v.getBaseUrl(ctx)

	if err != nil {
		return "", errors.Wrap(err, "No base URL provided for Unbody API")
	}

	return v.buildUrl(baseURL)
}

func (v *unbody) generateInput(prompt string, params unbodyparams.Params, settings config.ClassSettings) (generateInput, error) {
	messages := make([]interface{}, 0)
	messages = append(messages, map[string]string{
		"type":    "text",
		"role":    "user",
		"content": prompt,
	})

	defaultModel := settings.Model()
	model := params.Model

	if model == "" {
		model = defaultModel
	}

	return generateInput{
		Messages: messages,
		Vars:     make([]interface{}, 0),
		Data:     make([]interface{}, 0),
		Model:    model,
		Params: generateInputParams{
			FrequencyPenalty: params.FrequencyPenalty,
			MaxTokens:        params.MaxTokens,
			PresencePenalty:  params.PresencePenalty,
			Temperature:      params.Temperature,
			TopP:             params.TopP,
		},
	}, nil
}
func (v *unbody) generateInputWithMessages(messages []interface{}, data interface{}, params unbodyparams.Params, settings config.ClassSettings) (generateInput, error) {
	pv := params.Vars
	if pv == nil {
		pv = &[]interface{}{}
	}

	transformed := v.transformData(data)

	defaultModel := settings.Model()
	model := params.Model

	if model == "" {
		model = defaultModel
	}

	return generateInput{
		Messages: messages,
		Vars:     *pv,
		Data:     transformed,
		Model:    model,
		Params: generateInputParams{
			FrequencyPenalty: params.FrequencyPenalty,
			MaxTokens:        params.MaxTokens,
			PresencePenalty:  params.PresencePenalty,
			Temperature:      params.Temperature,
			TopP:             params.TopP,
		},
	}, nil
}

func transformSearchResult(data search.Result) map[string]interface{} {
	obj := data.ObjectWithVector(false)

	transformed := make(map[string]interface{})

	for key, value := range obj.Properties.(map[string]interface{}) {
		transformed[key] = value
	}

	transformed["id"] = obj.ID.String()

	return transformed
}

func (v *unbody) transformData(data interface{}) interface{} {
	switch v := data.(type) {
	case []interface{}:
		var transformedData []interface{}
		for _, res := range v {
			if result, valid := res.(search.Result); valid {
				transformedData = append(transformedData, transformSearchResult(result))
			}
		}
		return transformedData

	case search.Result:
		return transformSearchResult(v)
	}

	return nil
}

func (v *unbody) getError(statusCode int, resBodyError *unbodyApiError) error {
	endpoint := "Unbody"

	if resBodyError != nil {
		err := resBodyError.Message
		if resBodyError.Errors != nil && len(resBodyError.Errors) > 0 {
			err = strings.Join(resBodyError.Errors, ", ")
		}

		return fmt.Errorf("connection to: %s failed with status: %d error: %v", endpoint, statusCode, err)
	}
	return fmt.Errorf("connection to: %s failed with status: %d", endpoint, statusCode)
}

func (v *unbody) getApiKeyHeaderAndValue(apiKey string) (string, string) {
	return "Authorization", apiKey
}

func (v *unbody) getProjectIdHeaderAndValue(projectId string) (string, string) {
	return "X-Project-Id", projectId
}

func (v *unbody) getBaseUrl(ctx context.Context) (string, error) {
	var baseUrl, envVarValue, envVar string

	baseUrl = "X-Unbody-Generative-Base-Url"
	envVar = "UNBODY_GENERATIVE_BASE_URL"

	envVarValue = v.baseURL

	return v.getBaseUrlFromContext(ctx, baseUrl, envVarValue, envVar)
}

func (v *unbody) getBaseUrlFromContext(ctx context.Context, baseUrl, envVarValue, envVar string) (string, error) {
	if baseUrlValue := v.getValueFromContext(ctx, baseUrl); baseUrlValue != "" {
		return baseUrlValue, nil
	}
	if envVarValue != "" {
		return envVarValue, nil
	}
	return "", fmt.Errorf("no project ID found neither in request header: %s nor in environment variable under %s", baseUrl, envVar)
}

func (v *unbody) getApiKey(ctx context.Context) (string, error) {
	var apiKey, envVarValue, envVar string

	apiKey = "X-Unbody-Api-Key"
	envVar = "UNBODY_API_KEY"

	envVarValue = v.apiKey

	return v.getApiKeyFromContext(ctx, apiKey, envVarValue, envVar)
}

func (v *unbody) getApiKeyFromContext(ctx context.Context, apiKey, envVarValue, envVar string) (string, error) {
	if apiKeyValue := v.getValueFromContext(ctx, apiKey); apiKeyValue != "" {
		return apiKeyValue, nil
	}
	if envVarValue != "" {
		return envVarValue, nil
	}
	return "", fmt.Errorf("no api key found neither in request header: %s nor in environment variable under %s", apiKey, envVar)
}

func (v *unbody) getProjectId(ctx context.Context) (string, error) {
	var projectId, envVarValue, envVar string

	projectId = "X-Unbody-Project-Id"
	envVar = "UNBODY_PROJECT_ID"

	envVarValue = v.projectId

	return v.getProjectIdFromContext(ctx, projectId, envVarValue, envVar)
}

func (v *unbody) getProjectIdFromContext(ctx context.Context, projectId, envVarValue, envVar string) (string, error) {
	if projectIdValue := v.getValueFromContext(ctx, projectId); projectIdValue != "" {
		return projectIdValue, nil
	}
	if envVarValue != "" {
		return envVarValue, nil
	}
	return "", fmt.Errorf("no project ID found neither in request header: %s nor in environment variable under %s", projectId, envVar)
}

func (v *unbody) getValueFromContext(ctx context.Context, key string) string {
	if value := ctx.Value(key); value != nil {
		if keyHeader, ok := value.([]string); ok && len(keyHeader) > 0 && len(keyHeader[0]) > 0 {
			return keyHeader[0]
		}
	}
	// try getting header from GRPC if not successful
	if apiKey := modulecomponents.GetValueFromGRPC(ctx, key); len(apiKey) > 0 && len(apiKey[0]) > 0 {
		return apiKey[0]
	}

	return ""
}

type generateInput struct {
	Data     interface{}         `json:"data"`
	Messages []interface{}       `json:"messages,omitempty"`
	Vars     []interface{}       `json:"vars,omitempty"`
	Model    string              `json:"model,omitempty"`
	Params   generateInputParams `json:"params,omitempty"`
}

type generateInputParams struct {
	FrequencyPenalty *float64 `json:"frequencyPenalty,omitempty"`
	MaxTokens        *int     `json:"maxTokens,omitempty"`
	PresencePenalty  *float64 `json:"presencePenalty,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
}

type generateResponse struct {
	Data      *generateResponseData `json:"data,omitempty"`
	Message   string                `json:"message,omitempty"`
	ErrorCode string                `json:"errorCode,omitempty"`
	Errors    []string              `json:"errors,omitempty"`
}

type generateResponseData struct {
	Content       string `json:"content,omitempty"`
	UsageMetadata *usage `json:"usageMetadata,omitempty"`
	FinishReason  string `json:"finishReason,omitempty"`
}

type unbodyApiError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type usage struct {
	InputTokens  *int `json:"inputTokens,omitempty"`
	OutputTokens *int `json:"outputTokens,omitempty"`
	TotalTokens  *int `json:"totalTokens,omitempty"`
}
