package sentiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	language "cloud.google.com/go/language/apiv1"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"golang.org/x/net/context"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

const perspectiveURL = "https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze?key="

var nlpClient *language.Client
var perspectiveKey string

func InitNLP() {
	var err error
	nlpClient, err = language.NewClient(context.Background())
	if err != nil {
		panic(fmt.Errorf("Failed to create client: %v", err))
	}

	perspectiveKey = os.Getenv("PERSPECTIVE_API_CREDENTIALS")
	if perspectiveKey == "" {
		perspectiveKey = dao.GetPerspectiveKey()
	}
}

func GetNLP(ctx context.Context, text string) (*languagepb.AnnotateTextResponse, *PerspectiveResponse, error) {
	// Detects the sentiment of the text.
	now := time.Now()
	perChan := make(chan *PerspectiveResponse)
	go func() {
		perspective, err := GetPerspective(ctx, text)
		if err != nil {
			log.Printf("Perspective call failed %v", err)
			perChan <- nil
			return
		}
		perChan <- perspective
	}()
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
	perspect := <-perChan
	log.Printf("Done NLP for %s", text)
	log.Printf("NLP time %s", time.Since(now))
	if err != nil {
		log.Fatalf("Failed to analyze text: %v", err)
		return nil, nil, err
	}
	if perspect != nil {
		log.Printf("Perspective API: %v", perspect.AttributeScores["TOXICITY"].SummaryScore.Value)
	}
	if sentiment != nil {
		log.Printf("Sentiment API: %v", sentiment.DocumentSentiment.Score)
	}

	return sentiment, perspect, nil
}

/*

curl -H "Content-Type: application/json" --data \
    '{comment: {text: "what kind of idiot name is foo?"},
      languages: ["en"],
      requestedAttributes: {TOXICITY:{}} }' \
    https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze?key=YOUR_KEY_HERE

*/
type PerspectiveRequest struct {
	Comment             PerspectiveComment     `json:"comment,omitempty"`
	Context             PerspectiveContext     `json:"context,omitempty"`
	Languages           []string               `json:"languages,omitempty"`
	RequestedAttributes map[string]interface{} `json:"requestedAttributes,omitempty"`
	DoNotStore          bool                   `json:"doNotStore,omitempty"`
}

type PerspectiveComment struct {
	Text string `json:"text,omitempty"`
}

type PerspectiveContext struct {
	Entries []PerspectiveComment `json:"entries,omitempty"`
}

type PerspectiveResponse struct {
	AttributeScores map[string]PerspectiveModelResponse `json:"attributeScores,omitempty"`
	Languages       []string                            `json:"languages,omitempty"`
	ClientToken     string                              `json:"clientToken,omitempty"`
}

type PerspectiveModelResponse struct {
	SummaryScore PerspectiveScore  `json:"summaryScore,omitempty"`
	SpanScores   []PerspectiveSpan `json:"spanScores,omitempty"`
}

type PerspectiveScore struct {
	Value float64 `json:"value,omitempty"`
	Type  string  `json:"type,omitempty"`
}

type PerspectiveSpan struct {
	Begin int64            `json:"begin,omitempty"`
	End   int64            `json:"end,omitempty"`
	Score PerspectiveScore `json:"score,omitempty"`
}

func GetPerspective(ctx context.Context, text string) (*PerspectiveResponse, error) {
	pr := PerspectiveRequest{
		Comment: PerspectiveComment{
			Text: text,
		},
		Languages: []string{"en"},
		RequestedAttributes: map[string]interface{}{
			"TOXICITY": struct{}{},
		},
	}
	bb, err := json.Marshal(pr)
	if err != nil {
		return nil, err
	}

	rsp, err := http.Post(getPerspectiveURL(), "application/json", bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}

	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	pRsp := &PerspectiveResponse{}
	err = json.Unmarshal(rspBody, &pRsp)
	if err != nil {
		return nil, err
	}

	log.Printf("Got perspective %v", string(rspBody))
	return pRsp, nil
}

func getPerspectiveURL() string {
	return perspectiveURL + perspectiveKey
}
