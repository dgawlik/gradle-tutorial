package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

type OpenAIConfig struct {
	APIKey              string
	APIEndpoint         string
	DefaultModel        string
	DefaultMaxTokens    int
	DefaultInstructions string
}

type OpenAITranslation struct {
	TranslatedText string                `json:"translation"`
	Words          []map[string][]string `json:"words"`
}

type OpenAIRequest struct {
	Model        string `json:"model"`
	Prompt       string `json:"input"`
	MaxTokens    int    `json:"max_output_tokens"`
	Instructions string `json:"instructions"`
}

type App struct {
	Ctx             context.Context // Exported
	Config          OpenAIConfig    // Exported
	FirestoreClient *firestore.Client
}

// TranslationEntry represents a saved translation
type TranslationEntry struct {
	ID           int64     `json:"id"`
	OriginalText string    `json:"originalText"`
	Translation  string    `json:"translation"`
	Language     string    `json:"language"`
	CreatedAt    time.Time `json:"createdAt"`
}

// WordDefinition represents a translated word with its meanings
type WordDefinition struct {
	TranslationID int64    `json:"translationId"`
	OriginalWord  string   `json:"originalWord"`
	Meanings      []string `json:"meanings"`
}

// Database handles all data operations
type Database struct {
	Translations []TranslationEntry `json:"translations"`
	Definitions  []WordDefinition   `json:"definitions"`
	NextTransID  int                `json:"nextTransId"`
	NextWordID   int                `json:"nextWordId"`
	DBPath       string
}

type ChatMessage struct {
	Role    string        `json:"role"`
	Content []ChatContent `json:"content"`
}

type ChatContent struct {
	Text string `json:"text"`
}

type ChatGPTResponse struct {
	Output []ChatMessage `json:"output"`
}

func NewApp(firestore *firestore.Client, fCtx context.Context) *App {
	app := App{}
	app.Config = OpenAIConfig{
		APIKey:              "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",
		APIEndpoint:         "https://api.openai.com/v1/responses",
		DefaultModel:        "gpt-4.1",
		DefaultMaxTokens:    3000,
		DefaultInstructions: "\"You will output the following json structure: {\\\"translation\\\": \\\"<translated text>\\\", \\\"words\\\": [{\\\"<word>\\\": <word translation>}]. <translated text> is the translation of the from German to English. Try to follow idioms instead of dumb translating. \\\"words\\\" contains a pair: german word - 3 most common translations in english. Every word from original text should be in the list. Please do not include any more additional information in the output.\"",
	}

	app.FirestoreClient = firestore
	app.Ctx = fCtx

	return &app
}

func (a *App) Startup(ctx context.Context) {
	a.Ctx = ctx
}

func (a *App) DeleteTranslation(id int) error {

	docRef := a.FirestoreClient.Collection("translations").Doc(fmt.Sprintf("%d", id))
	_, err := docRef.Delete(a.Ctx)
	if err != nil {
		return fmt.Errorf("failed to delete translation from Firestore: %w", err)
	}
	return nil
}

func (a *App) GetTranslationNew(text string) (*TranslationEntry, error) {
	translation, err := a.translateWithChatGPT(text)
	if err != nil {
		return nil, err
	}

	docRef := a.FirestoreClient.Collection("counters").Doc("translationsCounter")
	snap, err := docRef.Get(a.Ctx)
	data := snap.Data()

	idx, ok := data["idx"].(int64) // Type assertion to string

	if !ok {
		return nil, fmt.Errorf("failed to get translation counter from Firestore: %w", err)
	}

	docRef.Set(a.Ctx, map[string]interface{}{
		"idx": idx + 1,
	})

	entry := TranslationEntry{
		ID:           idx + 1,
		OriginalText: text,
		Translation:  translation.TranslatedText,
		Language:     "de",
		CreatedAt:    time.Now(),
	}

	docRef2 := a.FirestoreClient.Collection("translations").Doc(fmt.Sprintf("%d", entry.ID))
	_, err = docRef2.Set(a.Ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("failed to save translation to Firestore: %w", err)
	}

	bulkWriter := a.FirestoreClient.BulkWriter(a.Ctx)
	wordsToInsert := make(map[string]WordDefinition)

	for _, wordEntry := range translation.Words {
		for translatedW, meanings := range wordEntry {
			wordsToInsert[translatedW] = WordDefinition{
				TranslationID: entry.ID,
				OriginalWord:  translatedW,
				Meanings:      meanings,
			}
		}
	}

	for word, wordDef := range wordsToInsert {
		docRef := a.FirestoreClient.Collection("words").Doc(word)
		_, err := bulkWriter.Create(docRef, wordDef)
		if err != nil {
			return nil, fmt.Errorf("failed to bulk write word definitions: %w", err)
		}
	}

	bulkWriter.End()

	return &entry, nil
}

func (a *App) GetTranslation(id int) (*TranslationEntry, error) {
	docRef, err := a.FirestoreClient.Collection("translations").Doc(fmt.Sprintf("%d", id)).Get(a.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get translation from Firestore: %w", err)
	}
	if !docRef.Exists() {
		return nil, fmt.Errorf("translation with ID %d not found", id)
	}
	var entry TranslationEntry
	if err := docRef.DataTo(&entry); err != nil {
		return nil, fmt.Errorf("failed to parse translation data: %w", err)
	}
	return &entry, nil
}

func (a *App) translateWithChatGPT(text string) (*OpenAITranslation, error) {
	if a.Config.APIKey == "" {
		return nil, fmt.Errorf("ChatGPT API key not configured")
	}

	reqBody := OpenAIRequest{
		Model:        a.Config.DefaultModel,
		Prompt:       text,
		Instructions: a.Config.DefaultInstructions,
		MaxTokens:    a.Config.DefaultMaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", a.Config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.Config.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatGPTResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if len(chatResp.Output) == 0 {
		return nil, fmt.Errorf("no translation provided in response")
	}

	translation := strings.TrimSpace(chatResp.Output[0].Content[0].Text)
	translation = strings.Trim(translation, "\"")
	fmt.Println("Translation:", translation)

	payload := OpenAITranslation{}
	if err := json.Unmarshal([]byte(translation), &payload); err != nil {
		return nil, fmt.Errorf("error unmarshaling translation payload: %w", err)
	}

	return &payload, nil
}

func (a *App) GetWordDefinitions(word string) ([]string, error) {
	docRef, err := a.FirestoreClient.Collection("words").Doc(word).Get(a.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get word definitions from Firestore: %w", err)
	}

	if docRef.Exists() {
		var definitions []string
		var wordDef WordDefinition
		if err := docRef.DataTo(&wordDef); err != nil {
			return nil, fmt.Errorf("failed to parse word definition data: %w", err)
		}
		definitions = append(definitions, wordDef.Meanings...)
		return definitions, nil
	} else {
		return nil, fmt.Errorf("word %s not found in Firestore", word)
	}
}

// GetAllTranslations returns all saved translations
func (a *App) GetAllTranslations() ([]TranslationEntry, error) {
	query := a.FirestoreClient.Collection("translations").Documents(a.Ctx)
	var translations []TranslationEntry
	for {
		doc, err := query.Next()
		if err != nil {
			break
		}
		if doc.Exists() {
			var entry TranslationEntry
			if err := doc.DataTo(&entry); err != nil {
				return nil, fmt.Errorf("failed to parse translation data: %w", err)
			}
			translations = append(translations, entry)
		}
	}
	return translations, nil
}
