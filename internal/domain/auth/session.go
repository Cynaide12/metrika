package auth

type Session struct {
    ID        uint
    UserID    uint
    UserAgent string
    RefreshToken string
}
