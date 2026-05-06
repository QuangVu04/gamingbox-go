package dto

type NotificationTask struct {
	ReceiverID uint   `json:"receiver_id"`
	SenderID   uint   `json:"sender_id"`
	ActionType string `json:"action_type"`
	TargetID   uint   `json:"target_id"`
	TargetType string `json:"target_type"`
}

type NotificationResponse struct {
	ID         uint64 `json:"id"`
	SenderID   uint   `json:"sender_id"`
	ActionType string `json:"action_type"`
	TargetID   uint   `json:"target_id"`
	TargetType string `json:"target_type"`
	IsRead     bool   `json:"is_read"`
	CreatedAt  string `json:"created_at"`
}
