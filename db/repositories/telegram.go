package repositories

import (
	"context"
	"errors"

	"qr-parking/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TelegramRepo struct {
	pool *pgxpool.Pool
}

func NewTelegramRepo(pool *pgxpool.Pool) *TelegramRepo {
	return &TelegramRepo{pool: pool}
}

func (r *TelegramRepo) Create(ctx context.Context, acc *types.TelegramAccount) error {
	query := `INSERT INTO telegram_accounts (id, user_id, tg_user_id, tg_username) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET tg_user_id = $3, tg_username = $4, linked_at = NOW()`
	_, err := r.pool.Exec(ctx, query, acc.ID, acc.UserID, acc.TgUserID, acc.TgUsername)
	return err
}

func (r *TelegramRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*types.TelegramAccount, error) {
	query := `SELECT id, user_id, tg_user_id, tg_username, linked_at FROM telegram_accounts WHERE user_id = $1`
	var acc types.TelegramAccount
	err := r.pool.QueryRow(ctx, query, userID).Scan(&acc.ID, &acc.UserID, &acc.TgUserID, &acc.TgUsername, &acc.LinkedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &acc, err
}

func (r *TelegramRepo) GetByTgUserID(ctx context.Context, tgUserID int64) (*types.TelegramAccount, error) {
	query := `SELECT id, user_id, tg_user_id, tg_username, linked_at FROM telegram_accounts WHERE tg_user_id = $1`
	var acc types.TelegramAccount
	err := r.pool.QueryRow(ctx, query, tgUserID).Scan(&acc.ID, &acc.UserID, &acc.TgUserID, &acc.TgUsername, &acc.LinkedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &acc, err
}

func (r *TelegramRepo) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM telegram_accounts WHERE user_id = $1`, userID)
	return err
}
