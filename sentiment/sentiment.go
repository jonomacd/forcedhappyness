package sentiment

import (
	"log"
	"time"

	"github.com/hako/durafmt"
	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"golang.org/x/net/context"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

func CheckPost(ctx context.Context, user domain.User, sentiment *languagepb.AnnotateTextResponse, perspective *PerspectiveResponse) (bool, string, error) {
	now := time.Now()
	score := sentiment.DocumentSentiment.Score

	if perspective != nil {
		if tox, ok := perspective.AttributeScores["TOXICITY"]; ok {
			ps := tox.SummaryScore.Value

			if ps < 0.05 {
				score = score + 0.1
			}
			if ps < 0.1 {
				score = score + 0.1
			}
			if ps < 0.2 {
				score = score + 0.1
			}
			if ps > 0.4 {
				score = score - 0.1
			}
			if ps > 0.5 {
				score = score - 0.1
			}
			if ps > 0.6 {
				score = score - 0.1
			}
			if ps > 0.7 {
				score = score - 0.1
			}
			if ps > 0.8 {
				score = score - 0.1
			}
			if ps > 0.9 {
				score = score - 0.1
			}
			if score < -1 {
				score = -1
			}
		}
	}

	if user.AngryBanThreshold == 0 {
		user.AngryBanThreshold = -0.3
	}
	user.TotalSentimentEMA = user.CalculatSentimentEMA(float64(score))
	overall := user.OverallSentiment()
	log.Printf("User: %s, score: %v, overall: %v angryban: %v expire %s, count %v", user.Username, score, overall, user.AngryBanThreshold, user.AngryBanExpire, user.AngryBanCount)
	if overall < user.AngryBanThreshold && now.After(user.AngryBanExpire) && score < 0 {
		// We need to ban this fuck
		if user.AngryBanCount == 0 {
			// First offense
			user.AngryBanExpire = now.Add(time.Minute * 10)
		} else if user.AngryBanCount == 1 {
			user.AngryBanExpire = now.Add(time.Hour)
		} else if user.AngryBanCount == 2 {
			user.AngryBanExpire = now.Add(time.Hour * 24)
		} else if user.AngryBanCount >= 3 {
			user.AngryBanExpire = now.Add(time.Hour * 24 * 7)
		}

		user.AngryBanCount++
	}

	err := dao.UpdateUserPostAttemptStatistics(ctx, dao.User{User: user}, score)
	if err != nil {
		return false, "", err
	}

	if now.Before(user.AngryBanExpire) && score < 0 {
		ttl := user.AngryBanExpire.Sub(now)
		durafmt.ParseShort(ttl).String()
		return false, "What's with all the negativity? You can't post anything even remotely negative for another " + durafmt.ParseShort(ttl).String() +
			". Remember you can still say something nice.", nil
	}

	if score < -0.9 {
		return false, "Are you okay? That is just too much. Take a deep breath and try again. Be constructive not angry.", nil
	}
	if score < -0.7 {
		return false, "I don't know... That is a bit harsh. Try to reword it. Remember you are talking to a person not a screen.", nil
	}
	if score < -0.5 {
		return false, "Nicer. Be nicer. It's the whole point of why we are here and not some other corner of the internet", nil
	}

	if score < -0.3 {
		return false, "Try being a bit nicer. Remember be constructive not mean.", nil
	}

	if score < -0.2 {
		return false, "This one is on the line and we might be wrong about it. Try being a bit nicer.", nil
	}

	if score < 0 {
		return true, "Hmm... We'll let this slide but if you keep up these sorts of posts we'll start stopping you.", nil
	}

	return true, "", nil
}
