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

// InvokeModelWrapper is a wrapper that holds a reference to the Bedrock runtime client
type InvokeModelWrapper struct {
	BedrockRuntimeClient *bedrockruntime.Client
}

// TitanTextRequest represents the request payload for invoking the Titan text generation model
type TitanTextRequest struct {
	InputText            string               `json:"inputText"`
	TextGenerationConfig TextGenerationConfig `json:"textGenerationConfig"`
}

// TextGenerationConfig represents the configuration settings for text generation
type TextGenerationConfig struct {
	Temperature   float64  `json:"temperature"`
	TopP          float64  `json:"topP"`
	MaxTokenCount int      `json:"maxTokenCount"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

// TitanTextResponse represents the response from the Titan text generation model
type TitanTextResponse struct {
	InputTextTokenCount int      `json:"inputTextTokenCount"`
	Results             []Result `json:"results"`
}

// Result represents a single result from the Titan text generation model
type Result struct {
	TokenCount       int    `json:"tokenCount"`
	OutputText       string `json:"outputText"`
	CompletionReason string `json:"completionReason"`
}

// ProcessError handles errors that occur during model invocation
func ProcessError(err error, modelId string) {
	if strings.Contains(err.Error(), "ModelNotFoundException") {
		log.Fatalf(`
                    Model "%v" not found. 
                    Please verify that the requested model exists and is accessible
                    within the specified region.\n
                    `, modelId)
	} else {
		log.Printf("Couldn't invoke model: \"%v\". Here's why: %v\n", modelId, err)
	}
}

// InvokeTitanText invokes the Titan text generation model with the given prompt
func (wrapper InvokeModelWrapper) InvokeTitanText(ctx context.Context, prompt string) (*bedrockruntime.InvokeModelOutput, error) {
	modelId := "amazon.titan-text-express-v1"

	// Create the request payload
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

	// Invoke the model passing the body and model ID
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

// ProcessOutput processes the output from the Titan model
func ProcessOutput(userPrompt string) (string, error) {
	// Load the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	// Create a new Bedrock runtime client
	client := bedrockruntime.NewFromConfig(cfg)
	wrapper := InvokeModelWrapper{
		BedrockRuntimeClient: client,
	}

	ctx := context.Background()
	var response TitanTextResponse

	// Invoke the Titan text generation model
	output, err := wrapper.InvokeTitanText(ctx, userPrompt)
	if err := json.Unmarshal(output.Body, &response); err != nil {
		log.Fatal("failed to unmarshal", err)
	}
	if err != nil {
		log.Fatalf("failed to invoke TitanText model, %v", err)
	}

	return response.Results[0].OutputText, nil
}
