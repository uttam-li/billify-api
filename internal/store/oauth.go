package store

import (
	"context"
	"database/sql"
)

type OAuthProviderStore struct {
	db *sql.DB
}

func (o *OAuthProviderStore) CreateOrUpdate(ctx context.Context, authProvider *OAuthProvider) error {
	query := `
		INSERT INTO oauth_providers (user_id, provider_name, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, provider_name) DO UPDATE
		SET access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := o.db.ExecContext(
        ctx,
        query,
        authProvider.UserID,
        authProvider.Provider,
        authProvider.AccessToken,
        authProvider.RefreshToken,
		authProvider.ExpiresAt,
    )

    return err
}
