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

type ForgotPasswordRequest struct {
    Email string `json:"email" binding:"required,email"`
}

type VerifyCodeRequest struct {
    Email string `json:"email" binding:"required,email"`
    Code  string `json:"code" binding:"required,len=6"`
}

type VerifyCodeResponse struct {
    ResetToken string `json:"reset_token"`
}

type ResetPasswordRequest struct {
    ResetToken  string `json:"reset_token" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=8"`
}

type AuthResponse struct {
    User         *UserResponse `json:"user"`
    AccessToken  string        `json:"access_token"`
    RefreshToken string        `json:"refresh_token"`
    ExpiresIn    int           `json:"expires_in"`
}