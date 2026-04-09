package dto

type RegisterInput struct {
    Email    string
    Username string
    Password string
}

type LoginInput struct {
    Email      string
    Password   string
}

type RefreshTokenInput struct {
    RefreshToken string
}

type AuthResponse struct {
    User         *UserResponse `json:"user"`
    AccessToken  string        `json:"access_token"`
    RefreshToken string        `json:"refresh_token"`
    ExpiresIn    int           `json:"expires_in"`
}