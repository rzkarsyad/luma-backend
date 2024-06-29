package model

type ChatRequest struct {
	Text string `json:"text"`
}

type Part struct {
	Text string `json:"text"`
}

type Message struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type ChatHistory struct {
	Messages []Message `json:"messages"`
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

type APIResponse struct {
	Candidates []Candidate `json:"candidates"`
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
