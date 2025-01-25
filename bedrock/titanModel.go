package titan

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type InvokeModelWrapper struct {
	BedrockRuntimeClient *bedrockruntime.Client
}

// The next two structs are just nested JSON
type TitanTextRequest struct {
	InputText            string               `json:"inputText"`
	TextGenerationConfig TextGenerationConfig `json:"textGenerationConfig"`
}

type TextGenerationConfig struct {
	Temperature   float64  `json:"temperature"`
	TopP          float64  `json:"topP"`
	MaxTokenCount int      `json:"maxTokenCount"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

// The next two structs are nested JSON
type TitanTextResponse struct {
	InputTextTokenCount int      `json:"inputTextTokenCount"`
	Results             []Result `json:"results"`
}

type Result struct {
	TokenCount       int    `json:"tokenCount"`
	OutputText       string `json:"outputText"`
	CompletionReason string `json:"completionReason"`
}

func ProcessError(err error, modelId string) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "no such host") {
		log.Printf(`The Bedrock service is not available in the selected region.
                    Please double-check the service availability for your region at
                    https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/.\n`)
	} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
		log.Printf(`Could not resolve the foundation model from model identifier: \"%v\".
                    Please verify that the requested model exists and is accessible
                    within the specified region.\n
                    `, modelId)
	} else {
		log.Printf("Couldn't invoke model: \"%v\". Here's why: %v\n", modelId, err)
	}
}

func (wrapper InvokeModelWrapper) InvokeTitanText(ctx context.Context, prompt string) (*bedrockruntime.InvokeModelOutput, error) {
	modelId := "amazon.titan-text-express-v1"

	body, err := json.Marshal(TitanTextRequest{
		InputText: prompt,
		TextGenerationConfig: TextGenerationConfig{
			Temperature:   0,
			TopP:          1,
			MaxTokenCount: 3000,
		},
	})

	if err != nil {
		log.Fatal("failed to marshal", err)
	}

	// Invoke model passing the body and model ID
	output, err := wrapper.BedrockRuntimeClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Body:        body,
	})

	if err != nil {
		ProcessError(err, modelId)
	}

	return output, nil
}

func ProcessOutput(userPrompt string) (string, error) {
	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)
	wrapper := InvokeModelWrapper{
		BedrockRuntimeClient: client,
	}

	ctx := context.Background()
	var response TitanTextResponse

	output, err := wrapper.InvokeTitanText(ctx, userPrompt)
	if err := json.Unmarshal(output.Body, &response); err != nil {
		log.Fatal("failed to unmarshal", err)
	}
	if err != nil {
		log.Fatalf("failed to invoke TitanText model, %v", err)
	}

	return response.Results[0].OutputText, nil
}
