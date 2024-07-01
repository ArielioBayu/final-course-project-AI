package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type AIModelConnector struct {
	Client *http.Client
}

type Inputs struct {
	Table map[string][]string `json:"table"`
	Query string              `json:"query"`
}

type Response struct {
	Answer      string   `json:"answer"`
	Coordinates [][]int  `json:"coordinates"`
	Cells       []string `json:"cells"`
	Aggregator  string   `json:"aggregator"`
}

func CsvToSlice(data string) (map[string][]string, error) {

	reader := csv.NewReader(strings.NewReader(data))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	headers := records[0]

	for _, header := range headers {
		result[header] = []string{}
	}

	for _, datas := range records[1:] {
		for i, value := range datas {
			result[headers[i]] = append(result[headers[i]], value)
		}
	}

	return result, nil
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Authorization", "Bearer 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")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return Response{}, errors.New(string(bodyBytes))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Response{}, err
	}

	return response, nil
}

func main() {
	// Load CSV data
	fileCsv := "data-series.csv"
	fileContent, err := os.ReadFile(fileCsv)
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// Convert CSV to map
	dataMap, err := CsvToSlice(string(fileContent))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	connector := AIModelConnector{
		Client: &http.Client{},
	}

	token := os.Getenv("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")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Sistem Manajemen Energi Rumah Pintar")
	fmt.Println("Masukkan pertanyaan Anda tentang konsumsi energi:")

	for scanner.Scan() {
		query := scanner.Text()
		if query == "exit" {
			break
		}

		payload := Inputs{
			Table: dataMap,
			Query: query,
		}

		response, err := connector.ConnectAIModel(payload, token)
		if err != nil {
			fmt.Println("Error connecting to AI model:", err)
			continue
		}

		fmt.Println("response:", response.Answer)
		fmt.Println("Masukkan pertanyaan lainnya atau ketik 'exit' untuk keluar:")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from input:", err)
	}
}
