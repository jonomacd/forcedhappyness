package sentiment

import (
	"io/ioutil"
	"log"
	"net/http"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"golang.org/x/net/context"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"mvdan.cc/xurls"
)

var (
	urlMatcher = xurls.Relaxed()
)

func CheckLinks(post domain.Post, user domain.User) ([]domain.LinkDetails, bool, error) {
	foundURLs := urlMatcher.FindAllString(string(post.Text), -1)

	// They uploaded their own image
	if len(post.LinkDetails) > 0 {
		foundURLs = append(foundURLs, post.LinkDetails[0].Url)
	}
	linkDetails := make([]domain.LinkDetails, len(foundURLs))
	for ii, url := range foundURLs {
		res, err := http.Get(url)
		if err != nil {
			return nil, false, err
		}

		unknown, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, false, err
		}

		fileType := http.DetectContentType(unknown)
		log.Printf("Found URL %s with content type %s", url, fileType)
		linkDetails[ii] = domain.LinkDetails{
			MIME: fileType,
			Url:  url,
		}

		if linkDetails[ii].IsImage() {
			ai, err := checkImage(url)
			if err != nil {
				return nil, false, err
			}
			block := shouldBlock(ai.SafeSearchAnnotation)
			if block {
				return nil, true, nil
			}

			imageText := ""
			confidence := float32(0.0)
			iter := 0
			if ai.FullTextAnnotation != nil {
				imageText = ai.FullTextAnnotation.Text
				for _, page := range ai.FullTextAnnotation.Pages {
					for _, block := range page.Blocks {
						confidence += block.Confidence
						iter++
						log.Printf("confidence %v ", confidence)
					}
				}
				confidence = confidence / float32(iter)
			}

			log.Printf("confidence %v ---- Image text %v", confidence, imageText)
			if imageText != "" && confidence > 0.85 {
				atr, pr, err := GetNLP(context.Background(), imageText)
				if err != nil {
					return nil, false, err
				}

				block, _, err = CheckPost(context.Background(), user, atr, pr)
				if err != nil {
					return nil, false, err
				}

				if block {
					return nil, true, nil
				}
			}
		}
	}

	return linkDetails, false, nil
}

func checkImage(url string) (*visionpb.AnnotateImageResponse, error) {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	img := vision.NewImageFromURI(url)
	ai, err := client.AnnotateImage(ctx, &visionpb.AnnotateImageRequest{
		Image: img,
		Features: []*visionpb.Feature{
			{Type: visionpb.Feature_DOCUMENT_TEXT_DETECTION},
			{Type: visionpb.Feature_SAFE_SEARCH_DETECTION},
			{Type: visionpb.Feature_LABEL_DETECTION, MaxResults: 5},
		},
	})
	if err != nil {
		return nil, err
	}

	return ai, nil
}

func shouldBlock(ss *visionpb.SafeSearchAnnotation) bool {
	if ss == nil {
		return false
	}

	if blockLikelyhood(ss.Adult) {
		return true
	}

	return false
}

func blockLikelyhood(lh visionpb.Likelihood) bool {
	switch lh {
	case visionpb.Likelihood_LIKELY, visionpb.Likelihood_VERY_LIKELY:
		return true
	}

	return false
}
