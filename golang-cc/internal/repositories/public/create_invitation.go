package public

import (
	"context"
	"time"
)

func (r *Repository) CreateInvitation(ctx context.Context, id, email string, tokenHash []byte, invitedBy string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO invitations(id,email,token_hash,invited_by,expires_at) VALUES($1,$2,$3,$4,$5)`,
		id, email, tokenHash, invitedBy, expiresAt)
	return err
}
