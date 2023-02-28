package domain

type FollowRequest struct {
	ID        string `json:"id,omitempty"`
	Receiver  string `json:"receiver"`
	Requester string `json:"requester"`
	Status    Status `json:"status,omitempty"`
}

type User struct {
	ID        string `json:"id,omitempty"`
	Username  string `json:"username"`
	Age       int    `json:"age"`
	Residence string `json:"residence"`
	Gender    string `json:"gender"`
}

type Ad struct {
	TweetID   string `json:"tweet_id"`
	AgeFrom   int    `json:"age_from"`
	AgeTo     int    `json:"age_to"`
	Gender    string `json:"gender"`
	Residence string `json:"residence"`
}

type FeedInfo struct {
	Usernames []string `json:"usernames"`
	AdIds     []string `json:"ad_ids"`
}

type Status int

const (
	Pending Status = iota + 1
	Declined
	Accepted
)

func (status Status) String() string {
	return [...]string{"Pending", "Declined", "Accepted"}[status-1]
}

func (status Status) EnumIndex() int {
	return int(status)
}
