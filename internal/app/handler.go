package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"metafarm/internal/storage"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

func pingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	}
}

type analysisResult struct {
	Name             string `json:"name"`
	DaysBetweenWater int    `json:"days_between_water"`
	DaysToMaturity   int    `json:"days_to_maturity"`
}

func analysisHandler(storage storage.Storage, openaiKey string) gin.HandlerFunc {
	type request struct {
		EncodedImage string `json:"encoded_image"`
	}

	type response struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
		ID int `json:"id"`
	}

	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		encodedImage := req.EncodedImage
		if strings.HasPrefix(encodedImage, "data:image") {
			parts := strings.Split(encodedImage, ",")
			if len(parts) != 2 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format"})
				return
			}
			encodedImage = parts[1]
		}

		if _, err := base64.StdEncoding.DecodeString(encodedImage); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 image"})
			return
		}

		id, err := storage.CreateAnalysis()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create analysis record"})
			return
		}

		go func() {
			client := openai.NewClient(openaiKey)

			prompt := `Analyze this plant image and provide the following information in JSON format:
			{
				"name": "Korean name of the plant (prioritizing edible crops)",
				"days_between_water": number of days between watering,
				"days_to_maturity": number of days until harvest
			}
			Only respond with the JSON object, no additional text and no markdown.`

			completionResp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT4oMini,
					Messages: []openai.ChatCompletionMessage{
						{
							Role: openai.ChatMessageRoleUser,
							MultiContent: []openai.ChatMessagePart{
								{
									Type: openai.ChatMessagePartTypeText,
									Text: prompt,
								},
								{
									Type: openai.ChatMessagePartTypeImageURL,
									ImageURL: &openai.ChatMessageImageURL{
										URL: fmt.Sprintf("data:image/jpeg;base64,%s", encodedImage),
									},
								},
							},
						},
					},
				},
			)

			if err != nil {
				storage.UpdateAnalysisStatus(id, "failed", fmt.Sprintf("OpenAI API error: %v", err))
				return
			}

			var result analysisResult
			if err := json.Unmarshal([]byte(completionResp.Choices[0].Message.Content), &result); err != nil {
				log.Printf("Failed to parse analysis result: %s", completionResp.Choices[0].Message.Content)
				storage.UpdateAnalysisStatus(id, "failed", fmt.Sprintf("Failed to parse analysis result: %v", err))
				return
			}

			if err := storage.UpdateAnalysis(id, result.Name, result.DaysBetweenWater, result.DaysToMaturity); err != nil {
				storage.UpdateAnalysisStatus(id, "failed", fmt.Sprintf("Failed to update analysis: %v", err))
				return
			}
		}()

		response := response{
			Status: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusAccepted,
				Message: "Analysis started",
			},
			ID: id,
		}

		c.JSON(http.StatusAccepted, response)
	}
}

func analysisResultHandler(storage storage.Storage) gin.HandlerFunc {
	type response struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`

		Name             string `json:"name,omitempty"`
		DaysBetweenWater int    `json:"days_between_water,omitempty"`
		DaysToMaturity   int    `json:"days_to_maturity,omitempty"`
		Error            string `json:"error,omitempty"`
	}

	return func(c *gin.Context) {
		idStr := c.Param("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing analysis ID"})
			return
		}

		var id int
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid analysis ID"})
			return
		}

		analysis, err := storage.GetAnalysis(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analysis"})
			return
		}

		if analysis == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
			return
		}

		resp := response{
			Status: struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    http.StatusOK,
				Message: "Success",
			},
		}

		switch analysis.Status {
		case "completed":
			resp.Name = analysis.Name
			resp.DaysBetweenWater = analysis.DaysBetweenWater
			resp.DaysToMaturity = analysis.DaysToMaturity
		case "failed":
			resp.Status.Code = http.StatusInternalServerError
			resp.Status.Message = "Analysis failed"
			resp.Error = analysis.Error
		case "pending":
			resp.Status.Code = http.StatusAccepted
			resp.Status.Message = "Analysis in progress"
		}

		c.JSON(resp.Status.Code, resp)
	}
}
