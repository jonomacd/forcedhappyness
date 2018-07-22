package events

import (
	"context"
	"log"

	"github.com/jonomacd/forcedhappyness/site/dao"
	"github.com/jonomacd/forcedhappyness/site/domain"
	"github.com/jonomacd/forcedhappyness/site/push"
	"github.com/jonomacd/forcedhappyness/site/sentiment"
)

func EventSubmitPost(post domain.Post) {
	go func() {
		users := map[string]struct{}{}
		postUser, err := dao.ReadUserByID(context.Background(), post.UserID)
		if err != nil {
			log.Printf("Error reading post user: %v", err)
		}

		linkDetails, block, err := sentiment.CheckLinks(post, postUser.User)
		if err != nil {
			log.Printf("Unable to do sentiment on image %v. Letting through: %v", post.Text, err)
		}
		if block && err == nil {
			if err := dao.BlockExistingPost(context.Background(), dao.Post{Post: post}); err != nil {
				log.Printf("Error blockingpost %s: %v", post.ID, err)
			}
		}

		if len(linkDetails) > 0 {
			post.LinkDetails = linkDetails
			if err := dao.UpdatePostLinkDetails(context.Background(), dao.Post{Post: post}); err != nil {
				log.Printf("Error updating link details on post %s: %v", post.ID, err)
			}
		}

		// Mentions
		for _, mention := range post.Mentions {
			_, ok := users[mention]
			if !ok {

				n, err := dao.ReadNotifications(context.Background(), mention)
				if err != nil && err != dao.ErrNotFound {
					log.Printf("Error reading notification: %v", err)
				}

				for _, notification := range n {
					if !notification.Config.Mentions {
						continue
					}
					users[mention] = struct{}{}
					// TODO: Figure out push defaults
					err := push.SendPush(push.Notify{
						Body:  postUser.Name + " (@" + postUser.Username + ") has mentioned you in a post",
						Icon:  "/static/img/fh-logo.png",
						Badge: "/static/img/fh-logo.png",
						Title: "Mention",
						Data: map[string]interface{}{
							"url": "/post/" + post.ID + "#" + post.ID,
						},
					}, notification)
					if err != nil {
						log.Printf("Unable to send push notification: %v", err)
					}
				}
			}
		}

		// Followed
		followers, err := dao.ReadFollowers(context.Background(), post.UserID)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Unable to read followers: %v", err)
		}

		for _, f := range followers {
			_, ok := users[f]
			if !ok {
				n, err := dao.ReadNotifications(context.Background(), f)
				if err != nil && err != dao.ErrNotFound {
					log.Printf("Error reading notification: %v", err)
				}
				for _, notification := range n {
					if !notification.Config.FollowerPost {
						continue
					}
					users[f] = struct{}{}
					// TODO: Figure out push defaults
					err := push.SendPush(push.Notify{
						Body:  postUser.Name + " (@" + postUser.Username + ") has posted a new nice",
						Icon:  "/static/img/fh-logo.png",
						Badge: "/static/img/fh-logo.png",
						Title: "New Post",
						Data: map[string]interface{}{
							"url": "/post/" + post.ID + "#" + post.ID,
						},
					}, notification)
					if err != nil {
						log.Printf("Unable to send push notification: %v", err)
					}
				}
			}

		}

		// Replies
		if post.Parent != "" {
			p, err := dao.ReadPostByID(context.Background(), post.Parent)
			if err != nil {
				log.Printf("Error reading parent post: %v", err)
			}
			_, ok := users[p.UserID]
			if !ok {
				n, err := dao.ReadNotifications(context.Background(), p.UserID)
				if err != nil && err != dao.ErrNotFound {
					log.Printf("Error reading notification: %v", err)
				}
				for _, notification := range n {
					if !notification.Config.Replies {
						continue
					}
					users[p.UserID] = struct{}{}
					// TODO: Figure out push defaults
					err := push.SendPush(push.Notify{
						Body:  postUser.Name + " (@" + postUser.Username + ") has replied to your post",
						Icon:  "/static/img/fh-logo.png",
						Badge: "/static/img/fh-logo.png",
						Title: "Reply",
						Data: map[string]interface{}{
							"url": "/post/" + post.ID + "#" + post.ID,
						},
					}, notification)
					if err != nil {
						log.Printf("Unable to send push notification: %v", err)
					}
				}
			}

		}

		// 	if post.TopParent != "" && post.TopParent != post.Parent {
		// 		n, ok := users[post.TopParent]
		// 		if !ok {
		// 			n, err = dao.ReadNotifications(context.Background(), post.TopParent)
		// 			if err != nil {
		// 				log.Printf("Error reading notification: %v", err)
		// 			}
		// 		}
		// 		for _, notification := range n {
		// 			if !notification.Config.Replies {
		// 				continue
		// 			}
		// 			// TODO: Figure out push defaults
		// 			err := push.SendPush(push.Notify{
		// 				Body:  postUser.Name + " (@" + postUser.Username + ") has replied to your post",
		// 				Icon:  "/static/img/fh-logo.png",
		// 				Badge: "/static/img/fh-logo.png",
		// 				Title: "Reply",
		// 				Data: map[string]interface{}{
		// 					"url": "/post/" + post.ID + "#" + post.ID,
		// 				},
		// 			}, notification)
		// 			if err != nil {
		// 				log.Printf("Unable to send push notification: %v", err)
		// 			}
		// 		}
		// 	}

	}()
}

type LikeEvent struct {
	Post    string
	LikedBy string
}

func EventLike(like LikeEvent) {
	go func() {

		u, err := dao.ReadUserByID(context.Background(), like.LikedBy)
		if err != nil {
			log.Printf("Error reading user: %v", err)
		}

		p, err := dao.ReadPostByID(context.Background(), like.Post)

		notifications, err := dao.ReadNotifications(context.Background(), p.UserID)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Error reading notification: %v", err)
		}
		for _, notification := range notifications {
			if !notification.Config.Likes {
				continue
			}
			// TODO: Figure out push defaults
			err := push.SendPush(push.Notify{
				Body:  u.Name + " (@" + u.Username + ") has liked your post!",
				Icon:  "/static/img/fh-logo.png",
				Badge: "/static/img/fh-logo.png",
				Title: "Liked!",
				Data: map[string]interface{}{
					"url": "/post/" + like.Post + "#" + like.Post,
				},
			}, notification)
			if err != nil {
				log.Printf("Unable to send push notification: %v", err)
			}
		}

	}()
}

type FollowEvent struct {
	FollowBy string
	Followed string
}

func EventFollow(fe FollowEvent) {
	go func() {
		ctx := context.Background()
		followedBy, err := dao.ReadUserByID(ctx, fe.FollowBy)
		if err != nil {
			log.Printf("Cannot read followed by %s: %v", fe.FollowBy, err)
		}

		notifications, err := dao.ReadNotifications(ctx, fe.Followed)
		if err != nil && err != dao.ErrNotFound {
			log.Printf("Cannot read notifications: %v", err)
		}
		for _, notification := range notifications {
			if !notification.Config.FollowerGet {
				continue
			}

			// TODO: Figure out push defaults
			err := push.SendPush(push.Notify{
				Body:  followedBy.Name + " (@" + followedBy.Username + ") has followed you!",
				Icon:  "/static/img/fh-logo.png",
				Badge: "/static/img/fh-logo.png",
				Title: "Followed!",
				Data: map[string]interface{}{
					"url": "/user/" + followedBy.Username,
				},
			}, notification)
			if err != nil {
				log.Printf("Unable to send push notification: %v", err)
			}
		}

	}()
}
