package domain

import "time"

type VoteType string

const (
	Like    VoteType = "like"
	Dislike VoteType = "dislike"
)

var VoteTypes = []VoteType{Like, Dislike}

type Vote struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	ProductID string    `json:"product_id"`
	VoteType  VoteType  `json:"vote_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Created   bool      `json:"-"`
}

type UpsertVoteInput struct {
	SessionID string   `json:"session_id" binding:"required"`
	ProductID string   `json:"product_id" binding:"required"`
	VoteType  VoteType `json:"vote_type" binding:"required"`
}

type GetVoteFilter struct {
	SessionID string `json:"session_id"`
}

type VoteReportItem struct {
	ProductID    string  `json:"product_id"`
	LikeCount    int     `json:"like_count"`
	LikeRatio    float64 `json:"like_ratio"`
	DislikeCount int     `json:"dislike_count"`
	DislikeRatio float64 `json:"dislike_ratio"`
}

type VoteReportResponse []VoteReportItem
