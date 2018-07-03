package dao

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"strings"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/datastore"
	"github.com/jonomacd/forcedhappyness/site/domain"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

const (
	KindPost = "post"
)

type Post struct {
	domain.Post `datastore:",flatten"`

	SentimentMagnitude float32
	SentimentScore     float32
	EntitiesString     []string
	AnalysisBytes      []byte `datastore:",noindex"`

	K *datastore.Key `datastore:"__key__"`
}

func CreatePost(ctx context.Context, p Post) (string, error) {
	if p.ID == "" {
		p.ID = newToken("p_", 15)
	}

	if p.IsReply {
		if err := updateReplyCount(ctx, p.Parent, 1); err != nil {
			return "", err
		}
	} else {
		p.TopParent = p.ID
	}
	entities := make([]string, 0, len(p.Analysis.Entities))
	for _, ent := range p.Analysis.Entities {
		if !strings.HasPrefix(ent.Name, "#") && !strings.HasPrefix(ent.Name, "@") {
			entities = append(entities, strings.ToLower(ent.Name))
		}
	}
	p.SentimentMagnitude = p.Analysis.DocumentSentiment.Magnitude
	p.SentimentScore = p.Analysis.DocumentSentiment.Score
	p.EntitiesString = entities
	bb, err := json.Marshal(p.Analysis)
	if err != nil {
		return "", err
	}
	p.AnalysisBytes = bb

	// Escape post text html. We can't let this get in the DB without escaping!
	p.Post.Text = template.HTML(template.HTMLEscapeString(string(p.Post.Text)))
	for ii, mu := range p.Post.MentionsUsername {
		p.Post.MentionsUsername[ii] = template.HTMLEscapeString(mu)
	}
	for ii, ht := range p.Post.Hashtags {
		p.Post.Hashtags[ii] = template.HTMLEscapeString(ht)
	}

	key := datastore.NameKey(KindPost, p.ID, nil)
	k, err := ds.Put(ctx, key, &p)
	if err != nil {
		return "", err
	}
	err = updateUserPostedStatistics(ctx, p, p.UserID)

	return k.Name, err
}

func ReadPostByID(ctx context.Context, id string) (Post, error) {
	key := datastore.NameKey(KindPost, id, nil)
	p := Post{}
	err := ds.Get(ctx, key, &p)
	populatePost(&p)
	return p, err
}

func ReadPostAndRepliesByID(ctx context.Context, id string) ([]Post, error) {
	q := datastore.NewQuery(KindPost).Filter("TopParent =", id)

	posts := []Post{}
	_, err := ds.GetAll(ctx, q, &posts)
	if err == datastore.ErrNoSuchEntity || len(posts) == 0 {
		p, err := ReadPostByID(ctx, id)
		if err != nil {
			return []Post{}, err
		}

		return ReadPostAndRepliesByID(ctx, p.TopParent)
	}
	if err != nil {
		return []Post{}, err
	}
	for ii, p := range posts {
		populatePost(&p)
		posts[ii] = p
	}

	return posts, nil
}

func ReadPostsByTime(ctx context.Context, cursor string, limit int) ([]Post, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindPost).Filter("IsReply =", false).Order("-Date").Start(c).Limit(limit)

	posts := []Post{}
	it := ds.Run(ctx, q)
	for {
		p := Post{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Post{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []Post{}, "", err
		}
		populatePost(&p)
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		return []Post{}, "", ErrNotFound
	}

	return posts, c.String(), nil

}

func ReadPostsByUser(ctx context.Context, userID, cursor string, limit int) ([]Post, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindPost).Filter("IsReply =", false).Filter("UserID = ", userID).Order("-Date").Start(c).Limit(limit)

	posts := []Post{}

	it := ds.Run(ctx, q)
	for {
		p := Post{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Post{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []Post{}, "", err
		}
		populatePost(&p)
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		return []Post{}, "", ErrNotFound
	}

	return posts, c.String(), nil
}

func ReadPostsByUsers(ctx context.Context, userIDs []string, cursors map[string]string, limit int) ([]Post, map[string]string, error) {
	var posts []Post
	if cursors == nil {
		cursors = map[string]string{}
	}

	newCursors := map[string]string{}
	for _, userID := range userIDs {
		ups, c, err := ReadPostsByUser(ctx, userID, cursors[userID], limit)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return []Post{}, newCursors, err
		}

		if err == datastore.ErrNoSuchEntity {
			continue
		}
		posts = append(posts, ups...)

		newCursors[userID] = c
	}

	return posts, newCursors, nil
}

func ReadPostsBySubTime(ctx context.Context, sub, cursor string, limit int) ([]Post, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindPost).Filter("Sub = ", sub).Filter("IsReply =", false).Order("-Date").Start(c).Limit(limit)

	posts := []Post{}
	it := ds.Run(ctx, q)
	for {
		p := Post{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Post{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []Post{}, "", err
		}
		populatePost(&p)
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		return []Post{}, "", ErrNotFound
	}

	return posts, c.String(), nil
}

func SearchPostByMention(ctx context.Context, mention, cursor string, limit int) ([]Post, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindPost).Filter("Mentions = ", mention).Filter("IsReply =", false).Order("-Date").Start(c).Limit(limit)
	posts := []Post{}
	it := ds.Run(ctx, q)
	for {
		p := Post{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Post{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []Post{}, "", err
		}
		populatePost(&p)
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		return []Post{}, "", ErrNotFound
	}

	return posts, c.String(), nil
}

func ReadPostByHashtag(ctx context.Context, tag, cursor string, limit int) ([]Post, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindPost).Filter("Hashtags = ", tag).Filter("IsReply =", false).Order("-Date").Start(c).Limit(limit)
	posts := []Post{}
	it := ds.Run(ctx, q)
	for {
		p := Post{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Post{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []Post{}, "", err
		}
		populatePost(&p)
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		return []Post{}, "", ErrNotFound
	}

	return posts, c.String(), nil
}

func ReadPostByEntities(ctx context.Context, tag, cursor string, limit int) ([]Post, string, error) {
	c, err := datastore.DecodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	q := datastore.NewQuery(KindPost).Filter("EntitiesString = ", tag).Filter("IsReply =", false).Order("-Date").Start(c).Limit(limit)
	posts := []Post{}
	it := ds.Run(ctx, q)
	for {
		p := Post{}
		_, err := it.Next(&p)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Post{}, "", err
		}

		c, err = it.Cursor()
		if err != nil {
			return []Post{}, "", err
		}
		populatePost(&p)
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		return []Post{}, "", ErrNotFound
	}

	return posts, c.String(), nil
}

func DeletePost(ctx context.Context, postID string) error {
	key := datastore.NameKey(KindPost, postID, nil)
	return ds.Delete(ctx, key)
}

func BlockExistingPost(ctx context.Context, post Post) error {
	if err := DeletePost(ctx, post.ID); err != nil {
		return err
	}

	_, err := CreateRottenPost(ctx, RottenPost{
		RottenPost: domain.RottenPost{
			Date:   post.Date,
			Text:   string(post.Text),
			UserID: post.UserID,
		},
	})

	return err
}

func UpdatePostLinkDetails(ctx context.Context, post Post) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindPost, post.ID, nil)

	p := &Post{}
	if err := tx.Get(key, p); err != nil {
		tx.Rollback()
		return err
	}
	p.Post.LinkDetails = post.LinkDetails
	if _, err := tx.Put(key, p); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func updateLikePost(ctx context.Context, postID string, update int64) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	key := datastore.NameKey(KindPost, postID, nil)

	p := &Post{}
	if err := tx.Get(key, p); err != nil {
		tx.Rollback()
		return err
	}
	p.Post.Likes += update
	if _, err := tx.Put(key, p); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func updateReplyCount(ctx context.Context, parentID string, update int64) error {
	tx, err := ds.NewTransaction(ctx)
	if err != nil {
		return err
	}

	for {
		key := datastore.NameKey(KindPost, parentID, nil)

		p := &Post{}
		if err := tx.Get(key, p); err != nil {
			tx.Rollback()
			log.Printf("error reading 1, %s: %v", parentID, err)
			return err
		}

		p.Post.ReplyCount += update
		if _, err := tx.Put(key, p); err != nil {
			tx.Rollback()
			log.Printf("error reading 2  %s: %v", parentID, err)
			return err
		}
		// Found the top parent
		if p.Post.TopParent == p.Post.ID {
			break
		}
		parentID = p.Parent
	}
	if _, err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func populatePost(p *Post) error {
	analysis := &languagepb.AnnotateTextResponse{}
	err := json.Unmarshal(p.AnalysisBytes, analysis)

	p.Analysis = analysis

	return err
}
