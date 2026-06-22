package share

type CreateBody struct {
	Password string `json:"password"`
	Expires  string `json:"expires"`
	Unit     string `json:"unit"`
	// Type is the share type: "preview" (open the online preview page) or
	// "download" (open directly triggers a download). Empty is treated as
	// "preview" for backward compatibility.
	Type string `json:"type"`
}

// Link is the information needed to build a shareable link.
type Link struct {
	Hash   string `json:"hash" storm:"id,index"`
	Path   string `json:"path" storm:"index"`
	UserID uint   `json:"userID"`
	Expire int64  `json:"expire"`
	// Type is the share type: "preview" or "download". Empty is treated as
	// "preview" for backward compatibility with links created before this field
	// existed.
	Type         string `json:"type"`
	PasswordHash string `json:"password_hash,omitempty"`
	// Token is a random value that will only be set when PasswordHash is set. It is
	// URL-Safe and is used to download links in password-protected shares via a
	// query arg.
	Token string `json:"token,omitempty"`
	// CreatedAt is the Unix timestamp (seconds) when the share was created. Links
	// created before this field existed have CreatedAt == 0, which should be
	// treated as unknown/earliest for backward compatibility.
	CreatedAt int64 `json:"createdAt"`
}
