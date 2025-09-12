package repo

import (
	"errors"
	"strings"
	
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
)

type RankingRow struct {
	Position int    `json:"position"`
	Username string `json:"username"`
	City     string `json:"city"`
	Votes    int64  `json:"votes"`
}

type VoteRepository interface {
	CastOnce(userID, videoID uuid.UUID) error
	CountByVideo(videoID uuid.UUID) (int64, error)
	TopByCity(limit int, city string) ([]RankingRow, error)
}

type voteRepo struct{ db *gorm.DB }

func NewVoteRepo(db *gorm.DB) VoteRepository { return &voteRepo{db} }

var ErrDuplicateVote = errors.New("user has already voted for this video")

func (r *voteRepo) CastOnce(userID, videoID uuid.UUID) error {
	v := domain.Vote{UserID: userID, VideoID: videoID}
	err := r.db.Create(&v).Error
	
	// Check if error is due to unique constraint violation
	if err != nil && (strings.Contains(err.Error(), "duplicate key") || 
		strings.Contains(err.Error(), "UNIQUE constraint") || 
		strings.Contains(err.Error(), "idx_user_video")) {
		return ErrDuplicateVote
	}
	
	return err
}

func (r *voteRepo) CountByVideo(videoID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&domain.Vote{}).Where("video_id = ?", videoID).Count(&n).Error
	return n, err
}

func (r *voteRepo) TopByCity(limit int, city string) ([]RankingRow, error) {
	var rows []RankingRow
	q := r.db.Table("votes v").
		Select("ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC) as position, CONCAT(u.first_name, ' ', u.last_name) as username, u.city, COUNT(*) as votes").
		Joins("JOIN videos vd ON vd.id = v.video_id AND vd.is_public_for_vote = true").
		Joins("JOIN users u ON u.id = vd.user_id").
		Group("u.id, u.first_name, u.last_name, u.city").
		Order("votes DESC").
		Limit(limit)
	if city != "" { 
		q = q.Where("u.city = ?", city) 
	}
	return rows, q.Scan(&rows).Error
}
