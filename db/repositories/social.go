package repositories

import (
	"context"
	"errors"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SocialProfileRepo struct {
	pool *pgxpool.Pool
}

func NewSocialProfileRepo(pool *pgxpool.Pool) *SocialProfileRepo {
	return &SocialProfileRepo{pool: pool}
}

func (r *SocialProfileRepo) Create(ctx context.Context, sp *types.SocialProfile) error {
	query := `INSERT INTO social_profiles (id, user_id, platform, handle, is_public) 
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, sp.ID, sp.UserID, sp.Platform, sp.Handle, sp.IsPublic)
	return err
}

func (r *SocialProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]types.SocialProfile, error) {
	query := `SELECT id, user_id, platform, handle, is_public, created_at 
		FROM social_profiles WHERE user_id = $1 ORDER BY platform`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []types.SocialProfile
	for rows.Next() {
		var sp types.SocialProfile
		if err := rows.Scan(&sp.ID, &sp.UserID, &sp.Platform, &sp.Handle, &sp.IsPublic, &sp.CreatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, sp)
	}
	return profiles, nil
}

func (r *SocialProfileRepo) GetByUserIDAndPlatform(ctx context.Context, userID uuid.UUID, platform types.SocialPlatform) (*types.SocialProfile, error) {
	query := `SELECT id, user_id, platform, handle, is_public, created_at 
		FROM social_profiles WHERE user_id = $1 AND platform = $2`
	var sp types.SocialProfile
	err := r.pool.QueryRow(ctx, query, userID, platform).Scan(
		&sp.ID, &sp.UserID, &sp.Platform, &sp.Handle, &sp.IsPublic, &sp.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func (r *SocialProfileRepo) UpsertByPlatform(ctx context.Context, userID uuid.UUID, platform types.SocialPlatform, handle string, isPublic bool) error {
	existing, err := r.GetByUserIDAndPlatform(ctx, userID, platform)
	if err != nil {
		return err
	}
	if existing != nil {
		_, err := r.Update(ctx, existing.ID, types.SocialProfileUpdate{Handle: &handle, IsPublic: &isPublic})
		return err
	}
	sp := &types.SocialProfile{
		ID:       uuid.New(),
		UserID:   userID,
		Platform: platform,
		Handle:   handle,
		IsPublic: isPublic,
	}
	return r.Create(ctx, sp)
}

func (r *SocialProfileRepo) GetPublicByUserID(ctx context.Context, userID uuid.UUID) ([]types.SocialProfile, error) {
	query := `SELECT id, user_id, platform, handle, is_public, created_at 
		FROM social_profiles WHERE user_id = $1 AND is_public = TRUE ORDER BY platform`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []types.SocialProfile
	for rows.Next() {
		var sp types.SocialProfile
		if err := rows.Scan(&sp.ID, &sp.UserID, &sp.Platform, &sp.Handle, &sp.IsPublic, &sp.CreatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, sp)
	}
	return profiles, nil
}

func (r *SocialProfileRepo) Update(ctx context.Context, id uuid.UUID, upd types.SocialProfileUpdate) (*types.SocialProfile, error) {
	query := `UPDATE social_profiles SET 
		handle = COALESCE($2, handle), 
		is_public = COALESCE($3, is_public)
		WHERE id = $1
		RETURNING id, user_id, platform, handle, is_public, created_at`
	var sp types.SocialProfile
	err := r.pool.QueryRow(ctx, query, id, upd.Handle, upd.IsPublic).Scan(
		&sp.ID, &sp.UserID, &sp.Platform, &sp.Handle, &sp.IsPublic, &sp.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &sp, err
}

func (r *SocialProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM social_profiles WHERE id = $1`, id)
	return err
}
