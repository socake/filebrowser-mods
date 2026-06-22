package fbhttp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenExpiryRequest is the body of POST /api/token. expiry selects how long
// the generated MCP access token stays valid.
type tokenExpiryRequest struct {
	Expiry string `json:"expiry"`
}

// tokenResponse is returned by tokenPostHandler. ExpiresAt is the Unix second
// at which the token expires, or 0 for a "permanent" token.
type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

// permanentTokenDuration is used for "permanent" tokens. We can't omit the
// ExpiresAt claim because withUser parses with jwt.WithExpirationRequired(),
// so we set a far-future expiry (~100 years) instead.
const permanentTokenDuration = time.Hour * 24 * 365 * 100

// tokenExpiryDuration maps an expiry keyword to a duration. The bool reports
// whether the keyword is "permanent" (so the response can report expiresAt=0).
func tokenExpiryDuration(expiry string) (time.Duration, bool, bool) {
	switch expiry {
	case "permanent":
		return permanentTokenDuration, true, true
	case "7d":
		return time.Hour * 24 * 7, false, true
	case "30d":
		return time.Hour * 24 * 30, false, true
	case "90d":
		return time.Hour * 24 * 90, false, true
	case "1y":
		return time.Hour * 24 * 365, false, true
	default:
		return 0, false, false
	}
}

// tokenPostHandler issues a JWT access token for the authenticated user, using
// the exact same claims structure and signing key as loginHandler/printToken,
// differing only in the chosen expiry. This lets the web UI generate long-lived
// MCP access tokens without going through the lumen CLI.
var tokenPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if r.Body == nil {
		return http.StatusBadRequest, nil
	}

	req := &tokenExpiryRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return http.StatusBadRequest, err
	}

	duration, permanent, ok := tokenExpiryDuration(req.Expiry)
	if !ok {
		return http.StatusBadRequest, nil
	}

	user := d.user
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := &authToken{
		User: userInfo{
			ID:                    user.ID,
			Locale:                user.Locale,
			ViewMode:              user.ViewMode,
			SingleClick:           user.SingleClick,
			RedirectAfterCopyMove: user.RedirectAfterCopyMove,
			Perm:                  user.Perm,
			LockPassword:          user.LockPassword,
			Commands:              user.Commands,
			HideDotfiles:          user.HideDotfiles,
			DisableThumbnails:     user.DisableThumbnails,
			DateFormat:            user.DateFormat,
			Username:              user.Username,
			AceEditorTheme:        user.AceEditorTheme,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "File Browser",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(d.settings.Key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	resp := tokenResponse{Token: signed}
	if !permanent {
		resp.ExpiresAt = expiresAt.Unix()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
})
