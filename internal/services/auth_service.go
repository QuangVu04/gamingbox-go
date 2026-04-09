package services

import (
	"errors"
	"time"

	"vault/be/config"
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
        Password: 	  hashedPassword,
        Username:     input.Username,
        Role:         models.RoleUser,
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo tài khoản")
    }

    return s.buildAuthResponse(user)
}

func (s *authService) Login(input dto.LoginInput) (*dto.AuthResponse, error) {
    user, err := s.userRepo.FindByEmail(input.Email)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, dto.NewServiceError("INVALID_CREDENTIALS", "email hoặc mật khẩu không đúng")
        }
        return nil, dto.NewServiceError("SERVER_ERROR", "Lỗi server")
    }

    if !utils.CheckPassword(input.Password, user.Password) {
        return nil, dto.NewServiceError("INVALID_CREDENTIALS", "email hoặc mật khẩu không đúng")
    }

    return s.buildAuthResponse(user)
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

    res, err := s.buildAuthResponse(user)
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

    return s.buildAuthResponse(user)
}

func (s *authService) Logout(tokenString string) {
    if tokenString != "" {
        _ = s.tokenRepo.Revoke(tokenString)
    }
}

func (s *authService) buildAuthResponse(user *models.User) (*dto.AuthResponse, error) {
    tokens, err := utils.GenerateTokenPair(user.ID, user.Username, string(user.Role))

    if err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo token")
    }

    rt := &models.RefreshToken{
        UserID:    user.ID,
        Token:     tokens.RefreshToken,
        ExpiresAt: time.Now().Add(config.App.JWTRefreshExpires),
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