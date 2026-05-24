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
	"strconv"

	"gorm.io/gorm"
)

type AuthService interface {
    RequestRegisterOTP(input dto.RegisterInput) error
    VerifyRegisterOTP(input dto.VerifyRegisterOTPInput) (*dto.AuthResponse, error)
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
    db          *gorm.DB
}

func NewAuthService(
    userRepo repositories.UserRepository,
    tokenRepo repositories.RefreshTokenRepository,
    db *gorm.DB,
) AuthService {
    return &authService{
        userRepo:  userRepo,
        tokenRepo: tokenRepo,
        db:        db,
    }
}

func (s *authService) RequestRegisterOTP(input dto.RegisterInput) error {
    if !utils.IsValidUsername(input.Username) {
        return dto.NewFieldError(
            "USERNAME_INVALID",
            "username chỉ được chứa chữ cái, số và dấu gạch dưới",
            "username",
        )
    }

    if _, err := s.userRepo.FindByEmail(input.Email); err == nil {
        return dto.NewFieldError("EMAIL_EXISTS", "email này đã được sử dụng", "email")
    }

    if _, err := s.userRepo.FindByUsername(input.Username); err == nil {
        return dto.NewFieldError("USERNAME_EXISTS", "username này đã được sử dụng", "username")
    }

    // Generate 6-digit code
    code := fmt.Sprintf("%06d", mathRand.Intn(1000000))
    
    // Store in Redis with 5 mins expiration
    ctx := context.Background()
    err := database.RDB.Set(ctx, "register_otp:"+input.Email, code, 5*time.Minute).Err()
    if err != nil {
        return dto.NewServiceError("SERVER_ERROR", "Không thể tạo mã xác nhận")
    }

    fmt.Printf("====[TESTING] Mã xác nhận đăng ký cho %s là: %s ====\n", input.Email, code)

    // Send email
    go func() {
        body := utils.GenerateOTPEmailTemplate(
            "Mã xác nhận đăng ký",
            "Cảm ơn bạn đã đăng ký tài khoản tại GamingBox. Vui lòng nhập mã xác nhận gồm 6 chữ số bên dưới để hoàn tất quá trình đăng ký:",
            code,
        )
        err := utils.SendEmail(input.Email, "Mã xác nhận đăng ký - GamingBox", body)
        if err != nil {
            fmt.Printf("====[LỖI GỬI EMAIL] Không thể gửi email tới %s: %v ====\n", input.Email, err)
        } else {
            fmt.Printf("====[THÀNH CÔNG] Đã gửi email tới %s ====\n", input.Email)
        }
    }()

    return nil
}

func (s *authService) VerifyRegisterOTP(input dto.VerifyRegisterOTPInput) (*dto.AuthResponse, error) {
    ctx := context.Background()

    // Verify code from Redis
    storedCode, err := database.RDB.Get(ctx, "register_otp:"+input.Email).Result()
    if err != nil || storedCode != input.Code {
        return nil, dto.NewFieldError("INVALID_CODE", "Mã xác nhận không đúng hoặc đã hết hạn", "code")
    }

    // Double check email/username uniqueness just in case
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
        Password:     &hashedPassword,
        Username:     input.Username,
        Role:         models.RoleUser,
    }

    if err := s.userRepo.Create(user); err != nil {
        return nil, dto.NewServiceError("SERVER_ERROR", "không thể tạo tài khoản")
    }

    // Delete the OTP code
    database.RDB.Del(ctx, "register_otp:"+input.Email)

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

    if user.Status == "banned" {
        return nil, dto.NewServiceError("ACCOUNT_BANNED", "Tài khoản của bạn đã bị khóa")
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

    if user.Status == "banned" {
        return nil, dto.NewServiceError("ACCOUNT_BANNED", "Tài khoản của bạn đã bị khóa")
    }

    res, err := s.buildAuthResponse(user, true)
    if err != nil {
        return nil, err
    }

    // Background sync Steam data
    go s.SyncSteamLibraryAndWishlist(steamID64, user.ID)

    return res, nil
}

func (s *authService) GetMe(userID uint) (*dto.UserResponse, error) {
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        return nil, dto.NewServiceError("USER_NOT_FOUND", "tài khoản không tồn tại")
    }

    return mapper.ToUserResponse(user, nil), nil
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

    if user.Status == "banned" {
        return nil, dto.NewServiceError("ACCOUNT_BANNED", "Tài khoản của bạn đã bị khóa")
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
        body := utils.GenerateOTPEmailTemplate(
            "Đặt lại mật khẩu",
            "Chúng tôi nhận được yêu cầu đặt lại mật khẩu cho tài khoản GamingBox của bạn. Vui lòng nhập mã xác nhận gồm 6 chữ số bên dưới để tiếp tục:",
            code,
        )
        err := utils.SendEmail(email, "Mã xác nhận đặt lại mật khẩu - GamingBox", body)
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
        User:         mapper.ToUserResponse(user, nil),
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

type SteamOwnedGamesResponse struct {
	Response struct {
		GameCount int `json:"game_count"`
		Games     []struct {
			AppID int `json:"appid"`
		} `json:"games"`
	} `json:"response"`
}

type AuthSteamAppDetails struct {
	Success bool `json:"success"`
	Data    struct {
		Name             string `json:"name"`
		ShortDescription string `json:"short_description"`
		IsFree           bool   `json:"is_free"`
		PriceOverview    struct {
			Initial int `json:"initial"` // Tính bằng cent
		} `json:"price_overview"`
		ReleaseDate struct {
			Date string `json:"date"`
		} `json:"release_date"`
		HeaderImage string `json:"header_image"`
		Developers  []string `json:"developers"`
		Genres      []struct {
			Description string `json:"description"`
		} `json:"genres"`
		Platforms struct {
			Windows bool `json:"windows"`
			Mac     bool `json:"mac"`
			Linux   bool `json:"linux"`
		} `json:"platforms"`
		Screenshots []struct {
			PathThumbnail string `json:"path_thumbnail"`
			PathFull      string `json:"path_full"`
		} `json:"screenshots"`
	} `json:"data"`
}

func (s *authService) fetchAndCreateMissingGames(missingAppIDs []int) map[int]uint {
	newGamesMap := make(map[int]uint)
	for _, appID := range missingAppIDs {
		fmt.Printf("[SteamSync] Đang cào dữ liệu game bị thiếu từ Steam: AppID %d\n", appID)
		url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&l=vietnamese", appID)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("[SteamSync] Lỗi HTTP khi fetch appID %d: %v\n", appID, err)
			continue
		}

		var result map[string]AuthSteamAppDetails
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		appIDStr := strconv.Itoa(appID)
		appData, ok := result[appIDStr]
		if !ok || !appData.Success {
			continue
		}

		data := appData.Data
		studioName := "Unknown Studio"
		if len(data.Developers) > 0 && data.Developers[0] != "" {
			studioName = data.Developers[0]
		}
		var studio models.Studio
		s.db.FirstOrCreate(&studio, models.Studio{Name: studioName})

		parsedDate, err := time.Parse("2 Jan, 2006", data.ReleaseDate.Date)
		if err != nil {
			parsedDate = time.Now()
		}

		game := models.Game{
			SteamID:     appID,
			Title:       data.Name,
			Description: data.ShortDescription,
			IsFree:      data.IsFree,
			Price:       float64(data.PriceOverview.Initial) / 100.0,
			ReleaseDate: parsedDate,
			StudioID:    studio.ID,
		}
		if err := s.db.Create(&game).Error; err != nil {
			fmt.Printf("[SteamSync] Lỗi Create Game AppID %d: %v\n", appID, err)
			continue
		}
		fmt.Printf("[SteamSync] Đã lưu game thành công AppID %d, GameID %d\n", appID, game.ID)

		for _, g := range data.Genres {
			if g.Description == "" {
				continue
			}
			var genre models.Genre
			s.db.FirstOrCreate(&genre, models.Genre{Name: g.Description})
			s.db.Model(&game).Association("Genres").Append(&genre)
		}

		if data.Platforms.Windows {
			var p models.Platform
			s.db.FirstOrCreate(&p, models.Platform{Name: "Windows"})
			s.db.Model(&game).Association("Platforms").Append(&p)
		}
		if data.Platforms.Mac {
			var p models.Platform
			s.db.FirstOrCreate(&p, models.Platform{Name: "Mac"})
			s.db.Model(&game).Association("Platforms").Append(&p)
		}
		if data.Platforms.Linux {
			var p models.Platform
			s.db.FirstOrCreate(&p, models.Platform{Name: "Linux"})
			s.db.Model(&game).Association("Platforms").Append(&p)
		}

		coverUrl := fmt.Sprintf("https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/%d/library_600x900_2x.jpg", appID)
		coverImg := models.GameImg{
			OgURL:   coverUrl,
			Thumb:   coverUrl,
			ImgType: "cover",
			GameID:  game.ID,
		}
		s.db.Create(&coverImg)

		headerImg := models.GameImg{
			OgURL:   data.HeaderImage,
			Thumb:   data.HeaderImage,
			ImgType: "header",
			GameID:  game.ID,
		}
		s.db.Create(&headerImg)

		for _, ss := range data.Screenshots {
			screenshot := models.GameImg{
				OgURL:   ss.PathFull,
				Thumb:   ss.PathThumbnail,
				ImgType: "screenshot",
				GameID:  game.ID,
			}
			s.db.Create(&screenshot)
		}

		newGamesMap[appID] = game.ID

		// Tạm ngưng 1 giây để tránh Rate Limit
		time.Sleep(1 * time.Second)
	}

	return newGamesMap
}

func (s *authService) SyncSteamLibraryAndWishlist(steamID64 string, userID uint) {
	fmt.Printf("[SteamSync] Bắt đầu đồng bộ cho user %d (SteamID: %s)\n", userID, steamID64)

	// 1. Lấy Owned Games (Bao gồm cả các game Free-to-play hoặc Share mà user đã từng chơi)
	ownedUrl := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&include_played_free_games=1&format=json", config.App.SteamApiKey, steamID64)
	resp, err := http.Get(ownedUrl)
	var ownedGames SteamOwnedGamesResponse
	if err == nil {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&ownedGames)
	}

	ownedAppIDs := []int{}
	for _, g := range ownedGames.Response.Games {
		ownedAppIDs = append(ownedAppIDs, g.AppID)
	}

	// 2. Lấy Wishlist
	wishlistUrl := fmt.Sprintf("https://store.steampowered.com/wishlist/profiles/%s/wishlistdata/", steamID64)
	wishlistResp, err := http.Get(wishlistUrl)
	wishlistAppIDs := []int{}
	if err == nil {
		defer wishlistResp.Body.Close()
		var wishlistData map[string]interface{}
		if err := json.NewDecoder(wishlistResp.Body).Decode(&wishlistData); err == nil {
			for appIDStr := range wishlistData {
				// Steam might return success: 2 for private wishlist, which is not a map of appIds
				if appIDStr != "success" {
					var appID int
					if _, err := fmt.Sscanf(appIDStr, "%d", &appID); err == nil {
						wishlistAppIDs = append(wishlistAppIDs, appID)
					}
				}
			}
		}
	}

	// 3. Lọc ra các game có trong Database
	allAppIDs := append(ownedAppIDs, wishlistAppIDs...)
	if len(allAppIDs) == 0 {
		return
	}

	// Vì list có thể dài, chia batch nếu cần, nhưng SQLite/MySQL xử lý IN dễ dàng
	var dbGames []models.Game
	if err := s.db.Where("steam_id IN ?", allAppIDs).Find(&dbGames).Error; err != nil {
		fmt.Printf("[SteamSync] Lỗi query games: %v\n", err)
		return
	}

	dbGamesMap := make(map[int]uint)
	for _, g := range dbGames {
		dbGamesMap[g.SteamID] = g.ID
	}

	// 3.5 Tìm các game bị thiếu và crawl
	var missingAppIDs []int
	for _, appID := range allAppIDs {
		if _, exists := dbGamesMap[appID]; !exists {
			// Deduplicate if owned and wishlisted
			found := false
			for _, m := range missingAppIDs {
				if m == appID {
					found = true
					break
				}
			}
			if !found {
				missingAppIDs = append(missingAppIDs, appID)
			}
		}
	}

	if len(missingAppIDs) > 0 {
		newGames := s.fetchAndCreateMissingGames(missingAppIDs)
		// Merge map
		for k, v := range newGames {
			dbGamesMap[k] = v
		}
	}

	// 4. Lưu logs
	// Chỉ insert nếu user chưa log game này
	for _, appID := range ownedAppIDs {
		if gameID, exists := dbGamesMap[appID]; exists {
			logEntry := models.GameLog{UserID: userID, GameID: gameID, Status: "completed", LoggedAt: time.Now()}
			s.db.FirstOrCreate(&logEntry, models.GameLog{UserID: userID, GameID: gameID})
		}
	}

	for _, appID := range wishlistAppIDs {
		if gameID, exists := dbGamesMap[appID]; exists {
			logEntry := models.GameLog{UserID: userID, GameID: gameID, Status: "backlog", LoggedAt: time.Now()}
			s.db.FirstOrCreate(&logEntry, models.GameLog{UserID: userID, GameID: gameID})
		}
	}

	fmt.Printf("[SteamSync] Hoàn tất đồng bộ cho user %d\n", userID)
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