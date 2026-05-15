package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	mathRand "math/rand"
	"time"

	"vault/be/config"
	"vault/be/database"
	"vault/be/internal/dto"
    "vault/be/internal/dto/mapper"
	"vault/be/internal/models"
	"vault/be/internal/repositories"
	"vault/be/pkg/utils"
    "vault/be/internal/dto/response"

	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type AuthService interface {
    Register(input dto.RegisterInput) (*dto.AuthResponse, error)
    Login(input dto.LoginInput) (*dto.AuthResponse, error)
    GetMe(userID uint) (*dto.UserResponse, error)
    RefreshTokens(tokenString string) (*dto.AuthResponse, error)
    Logout(tokenString string)
    LoginWithSteam(steamID64 string) (*dto.AuthResponse, error)
    ForgotPassword(email string) error
    VerifyCode(req dto.VerifyCodeRequest) (*dto.VerifyCodeResponse, error)
    ResetPassword(req dto.ResetPasswordRequest) error
    HandleGoogleLogin(code string) (*dto.AuthResponse, error)
    HandleFacebookLogin(code string) (*dto.AuthResponse, error)
}

type authService struct {
    userRepo    repositories.UserRepository
    tokenRepo   repositories.RefreshTokenRepository
}

func NewAuthService(
    userRepo repositories.UserRepository,
    tokenRepo repositories.RefreshTokenRepository,
) AuthService {
    return &authService{
        userRepo:  userRepo,
        tokenRepo: tokenRepo,
    }
}

func (s *authService) Register(input dto.RegisterInput) (*dto.AuthResponse, error) {
    if !utils.IsValidUsername(input.Username) {
        return nil, dto.NewFieldError(
            "USERNAME_INVALID",
            "username chỉ được chứa chữ cái, số và dấu gạch dưới",
            "username",
        )
    }

    if _, err := s.userRepo.FindByEmail(input.Email); err == nil {
        return nil, dto.NewFieldError("EMAIL_EXISTS", "email này đã được sử dụng", "email")
    }

    if _, err := s.userRepo.FindByUsername(input.Username); err == nil {
        return nil, dto.NewFieldError("USERNAME_EXISTS", "username này đã được sử dụng", "username")
    }

    hashedPassword, err := utils.HashPassword(input.Password)
    if err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể xử lý mật khẩu")
    }

    user := &models.User{
        Email:        input.Email,
        Password: 	  &hashedPassword,
        Username:     input.Username,
        Role:         models.RoleUser,
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo tài khoản")
    }

    return s.buildAuthResponse(user, true)
}

func (s *authService) Login(input dto.LoginInput) (*dto.AuthResponse, error) {
    user, err := s.userRepo.FindByEmail(input.Email)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, dto.NewServiceError("INVALID_CREDENTIALS", "email hoặc mật khẩu không đúng")
        }
        return nil, dto.NewServiceError("SERVER_ERROR", "Lỗi server")
    }

    if user.Password == nil || !utils.CheckPassword(input.Password, *user.Password) {
        return nil, dto.NewServiceError("INVALID_CREDENTIALS", "email hoặc mật khẩu không đúng")
    }

    return s.buildAuthResponse(user, input.RememberMe)
}

func (s *authService) LoginWithSteam(steamID64 string) (*dto.AuthResponse, error) {
    steamProfile, err := GetSteamProfile(steamID64)
    if err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "Lỗi lấy thông tin Steam: "+err.Error())
    }

    if len(steamProfile.Response.Players) == 0 {
        return nil, dto.NewServiceError("SERVER_ERROR", "Steam không trả về thông tin người dùng")
    }

    player := steamProfile.Response.Players[0]
    user, err := s.userRepo.FindBySteamID(steamID64)

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            newUser := &models.User{
                SteamID:   steamID64,
                Username:  player.PersonaName,
                Role:      models.RoleUser,
                AvatarURL: &player.AvatarMedium, 
            }
            if err := s.userRepo.Create(newUser); err != nil {
                return nil, dto.NewServiceError("SERVER_ERROR", "Không thể tạo tài khoản: "+err.Error())
            }
            user = newUser 
        } else {
            return nil, err
        }
    }

    res, err := s.buildAuthResponse(user, true)
    if err != nil {
        return nil, err
    }
    return res, nil
}

func (s *authService) GetMe(userID uint) (*dto.UserResponse, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
    }

    return mapper.ToUserResponse(user), nil
}

func (s *authService) RefreshTokens(tokenString string) (*dto.AuthResponse, error) {
    claims, err := utils.ParseToken(tokenString)

    if err != nil || claims.Type != utils.RefreshToken {
        return nil, dto.NewServiceError("INVALID_REFRESH_TOKEN", "refresh token không hợp lệ hoặc đã hết hạn")
    }

    storedToken, err := s.tokenRepo.Find(tokenString)
    if err != nil {
        return nil, dto.NewServiceError("TOKEN_REVOKED", "token đã bị thu hồi hoặc không tồn tại")
    }

    user, err := s.userRepo.FindByID(storedToken.UserID)
    if err != nil {
        return nil, dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
    }

    _ = s.tokenRepo.Revoke(tokenString)

    return s.buildAuthResponse(user, true)
}

func (s *authService) Logout(tokenString string) {
    if tokenString != "" {
        _ = s.tokenRepo.Revoke(tokenString)
    }
}

func (s *authService) ForgotPassword(email string) error {
    // Check if user exists
    _, err := s.userRepo.FindByEmail(email)
    if err != nil {
        // We still return nil to prevent email enumeration attacks
        return nil
    }

    // Generate 6-digit code
    code := fmt.Sprintf("%06d", mathRand.Intn(1000000))
    
    // Store in Redis with 15 mins expiration
    ctx := context.Background()
    err = database.RDB.Set(ctx, "reset_code:"+email, code, 15*time.Minute).Err()
    if err != nil {
        return dto.NewServiceError("SERVER_ERROR", "Không thể tạo mã xác nhận")
    }

    fmt.Printf("====[TESTING] Mã xác nhận cho %s là: %s ====\n", email, code)

    // Send email
    go func() {
        body := fmt.Sprintf("Xin chào,\n\nMã xác nhận để đặt lại mật khẩu của bạn là: %s\nMã này sẽ hết hạn trong vòng 15 phút.\n\nTrân trọng,\nĐội ngũ GamingBox", code)
        err := utils.SendEmail(email, "Mã xác nhận đặt lại mật khẩu", body)
        if err != nil {
            fmt.Printf("====[LỖI GỬI EMAIL] Không thể gửi email tới %s: %v ====\n", email, err)
        } else {
            fmt.Printf("====[THÀNH CÔNG] Đã gửi email tới %s ====\n", email)
        }
    }()

    return nil
}

func (s *authService) VerifyCode(req dto.VerifyCodeRequest) (*dto.VerifyCodeResponse, error) {
    ctx := context.Background()

    // Verify code from Redis
    storedCode, err := database.RDB.Get(ctx, "reset_code:"+req.Email).Result()
    if err != nil || storedCode != req.Code {
        return nil, dto.NewFieldError("INVALID_CODE", "Mã xác nhận không đúng hoặc đã hết hạn", "code")
    }

    // Delete the code
    database.RDB.Del(ctx, "reset_code:"+req.Email)

    // Generate a secure reset token
    b := make([]byte, 32)
    rand.Read(b)
    resetToken := hex.EncodeToString(b)

    // Store the token with the email for 15 mins
    err = database.RDB.Set(ctx, "reset_token:"+resetToken, req.Email, 15*time.Minute).Err()
    if err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "Không thể tạo token đặt lại mật khẩu")
    }

    return &dto.VerifyCodeResponse{ResetToken: resetToken}, nil
}

func (s *authService) ResetPassword(req dto.ResetPasswordRequest) error {
    ctx := context.Background()
    
    // Get email from token
    email, err := database.RDB.Get(ctx, "reset_token:"+req.ResetToken).Result()
    if err != nil {
        return dto.NewFieldError("INVALID_TOKEN", "Token không hợp lệ hoặc đã hết hạn", "reset_token")
    }

    // Get user
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        return dto.NewServiceError("USER_NOT_FOUND", "Không tìm thấy tài khoản")
    }

    // Hash new password
    hashedPassword, err := utils.HashPassword(req.NewPassword)
    if err != nil {
        return dto.NewServiceError("SERVER_ERROR", "Không thể xử lý mật khẩu")
    }

    // Update password
    user.Password = &hashedPassword
    if err := s.userRepo.Update(user); err != nil {
        return dto.NewServiceError("SERVER_ERROR", "Không thể cập nhật mật khẩu")
    }

    // Delete the token
    database.RDB.Del(ctx, "reset_token:"+req.ResetToken)

    return nil
}

func (s *authService) buildAuthResponse(user *models.User, rememberMe bool) (*dto.AuthResponse, error) {
    refreshExpires := config.App.JWTRefreshShortExpires
    if rememberMe {
        refreshExpires = config.App.JWTRefreshExpires
    }

    tokens, err := utils.GenerateTokenPair(user.ID, user.Username, string(user.Role), refreshExpires)

    if err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo token")
    }

    rt := &models.RefreshToken{
        UserID:    user.ID,
        Token:     tokens.RefreshToken,
        ExpiresAt: time.Now().Add(refreshExpires),
    }

    if err := s.tokenRepo.Save(rt); err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể lưu token")
    }

    return &dto.AuthResponse{
        User:         mapper.ToUserResponse(user),
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        ExpiresIn:    int(config.App.JWTAccessExpires.Seconds()),
    }, nil
}

func GetSteamProfile(steamID64 string) (*response.SteamPlayerResponse, error) {
	url := fmt.Sprintf(config.App.SteamApiUrl, config.App.SteamApiKey, steamID64)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result response.SteamPlayerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Response.Players) == 0 {
		return nil, fmt.Errorf("không tìm thấy profile cho steamid: %s", steamID64)
	}

	return &result, nil
}

func (s *authService) HandleGoogleLogin(code string) (*dto.AuthResponse, error) {
	googleUser, err := utils.GetGoogleUserInfo(code)
	if err != nil {
		return nil, dto.NewServiceError("GOOGLE_AUTH_FAILED", err.Error())
	}

	// 1. Tìm theo GoogleID
	user, err := s.userRepo.FindByField("google_id", googleUser.ID)
	if err != nil {
		// 2. Nếu không thấy ID, tìm theo Email
		user, err = s.userRepo.FindByEmail(googleUser.Email)
		if err == nil {
			// Liên kết tài khoản hiện có với GoogleID
			user.GoogleID = googleUser.ID
			if user.AvatarURL == nil {
				user.AvatarURL = &googleUser.Picture
			}
			s.userRepo.Update(user)
		} else {
			// 3. Tạo mới nếu chưa có email
			user = &models.User{
				Email:     googleUser.Email,
				Username:  googleUser.Name,
				GoogleID:  googleUser.ID,
				AvatarURL: &googleUser.Picture,
				Role:      models.RoleUser,
			}
			if err := s.userRepo.Create(user); err != nil {
				return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo tài khoản")
			}
		}
	}

	return s.buildAuthResponse(user, true)
}

func (s *authService) HandleFacebookLogin(code string) (*dto.AuthResponse, error) {
	fbUser, err := utils.GetFacebookUserInfo(code)
	if err != nil {
		return nil, dto.NewServiceError("FACEBOOK_AUTH_FAILED", err.Error())
	}

	// 1. Tìm theo FacebookID
	user, err := s.userRepo.FindByField("facebook_id", fbUser.ID)
	if err != nil {
		// 2. Tìm theo Email (Facebook có thể không trả về email nếu user ko cấp quyền)
		if fbUser.Email != "" {
			user, err = s.userRepo.FindByEmail(fbUser.Email)
			if err == nil {
				user.FacebookID = fbUser.ID
				if user.AvatarURL == nil {
					user.AvatarURL = &fbUser.Picture.Data.URL
				}
				s.userRepo.Update(user)
			}
		}
		
		if user == nil {
			// 3. Tạo mới
			user = &models.User{
				Email:      fbUser.Email,
				Username:   fbUser.Name,
				FacebookID: fbUser.ID,
				AvatarURL:  &fbUser.Picture.Data.URL,
				Role:       models.RoleUser,
			}
			if err := s.userRepo.Create(user); err != nil {
				return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo tài khoản")
			}
		}
	}

	return s.buildAuthResponse(user, true)
}