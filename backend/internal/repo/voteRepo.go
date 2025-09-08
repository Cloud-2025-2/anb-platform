package repo

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/Cloud-2025-2/anb-platform/internal/domain"
)

type VoteRepository interface {
	CastOnce(userID, videoID uuid.UUID) error
	CountByVideo(videoID uuid.UUID) (int64, error)
	TopByCity(limit int, city string) ([]struct{ VideoID uuid.UUID; Votes int64 }, error)
}

type voteRepo struct{ db *gorm.DB }

func NewVoteRepo(db *gorm.DB) VoteRepository { return &voteRepo{db} }

func (r *voteRepo) CastOnce(userID, videoID uuid.UUID) error {
	v := domain.Vote{UserID: userID, VideoID: videoID}
	return r.db.Create(&v).Error // UNIQUE(user_id,video_id) protege duplicados
}

func (r *voteRepo) CountByVideo(videoID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.Model(&domain.Vote{}).Where("video_id = ?", videoID).Count(&n).Error
	return n, err
}

func (r *voteRepo) TopByCity(limit int, city string) ([]struct{ VideoID uuid.UUID; Votes int64 }, error) {
	var rows []struct{ VideoID uuid.UUID; Votes int64 }
	q := r.db.Table("votes v").
		Select("v.video_id, COUNT(*) as votes").
		Joins("JOIN videos vd ON vd.id = v.video_id AND vd.is_public_for_vote = true").
		Group("v.video_id").
		Order("votes DESC").
		Limit(limit)
	if city != "" { q = q.Joins("JOIN users u ON u.id = vd.user_id AND u.city = ?", city) }
	return rows, q.Scan(&rows).Error
}
