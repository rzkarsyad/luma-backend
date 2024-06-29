package repository

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"luma-backend/model"
)

type AIModelConnector struct {
	Client *http.Client
}

func CsvToSlice(data string) (map[string][]string, error) {
	r := csv.NewReader(strings.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	table := make(map[string][]string)
	headers := records[0]

	for _, header := range headers {
		table[header] = make([]string, 0)
	}

	for _, record := range records[1:] {
		for i, value := range record {
			table[headers[i]] = append(table[headers[i]], value)
		}
	}

	return table, nil
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (model.Response, error) {
	url := "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq"
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return model.Response{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return model.Response{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return model.Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Response{}, err
	}

	var response model.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return model.Response{}, err
	}

	return response, nil
}

func (c *AIModelConnector) GeminiRecommendation(query string, table map[string][]string, token string) (model.APIResponse, error) {
	prompt := query + "\n"
	for column, values := range table {
		prompt += column + ": " + strings.Join(values, ", ") + "\n"
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + token

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return model.APIResponse{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return model.APIResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return model.APIResponse{}, err
	}
	defer resp.Body.Close()

	var geminiResponse model.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&geminiResponse)
	if err != nil {
		return model.APIResponse{}, err
	}

	return geminiResponse, nil
}
