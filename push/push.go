package push

import (
	"encoding/json"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jonomacd/forcedhappyness/site/dao"
)

var privateKeyVAPID = ""

type Notify struct {
	Body               string                 `json:"body,omitempty"`
	Icon               string                 `json:"icon,omitempty"`
	Image              string                 `json:"image,omitempty"`
	Badge              string                 `json:"badge,omitempty"`
	Vibrate            string                 `json:"vibrate,omitempty"`
	Sound              string                 `json:"sound,omitempty"`
	Dir                string                 `json:"dir,omitempty"`
	Tag                string                 `json:"tag,omitempty"`
	Data               map[string]interface{} `json:"data,omitempty"`
	RequireInteraction string                 `json:"requireInteraction,omitempty"`
	Renotify           string                 `json:"renotify,omitempty"`
	Silent             bool                   `json:"silent,omitempty"`
	Actions            string                 `json:"actions,omitempty"`
	Timestamp          int64                  `json:"timestamp,omitempty"`
	Title              string                 `json:"title,omitempty"`
}

func Init() string {
	keys := dao.GetVAPIDKey()
	privateKeyVAPID = keys.PrivateKey

	return keys.PublicKey
}

func SendPush(notify Notify, notification dao.Notification) error {
	bb, err := json.Marshal(notify)
	if err != nil {
		return err
	}

	_, err = webpush.SendNotification(bb, notification.Subscription, &webpush.Options{
		VAPIDPrivateKey: privateKeyVAPID,
	})
	if err != nil {
		return err
	}

	return nil
}
