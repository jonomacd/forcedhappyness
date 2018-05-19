package domain

import (
	"fmt"
	"log"

	language "cloud.google.com/go/language/apiv1"
	"golang.org/x/net/context"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

var nlpClient *language.Client

func InitNLP() {
	var err error
	nlpClient, err = language.NewClient(context.Background())
	if err != nil {
		panic(fmt.Errorf("Failed to create client: %v", err))
	}

}

func GetNLP(ctx context.Context, text string) (*languagepb.AnnotateTextResponse, error) {

	// Detects the sentiment of the text.
	sentiment, err := nlpClient.AnnotateText(ctx, &languagepb.AnnotateTextRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		Features: &languagepb.AnnotateTextRequest_Features{
			ExtractEntities:          true,
			ExtractDocumentSentiment: true,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})
	if err != nil {
		log.Fatalf("Failed to analyze text: %v", err)
		return nil, err
	}

	return sentiment, nil
}
