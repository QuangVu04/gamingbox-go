package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"vault/be/config"
	_ "vault/be/internal/dto"
	"vault/be/internal/services"
	"vault/be/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/yohcop/openid-go"
)

var (
	nonceStore     = openid.NewSimpleNonceStore()
	discoveryCache = openid.NewSimpleDiscoveryCache() 
)

type SteamHandler struct{
	authService services.AuthService 
}

func NewSteamHandler(authService services.AuthService) *SteamHandler {
	return &SteamHandler{
		authService: authService,
	}
}

// LoginHandle godoc
// @Summary      Đăng nhập bằng Steam
// @Description  Chuyển hướng người dùng đến trang đăng nhập Steam OpenID
// @Tags         Authentication (Steam)
// @Success      302  "Redirect"
// @Router       /auth/steam [get]
func (h *SteamHandler) LoginHandle(c *gin.Context) {
	url, err := openid.RedirectURL(config.App.SteamOpenIdUrl, config.App.ReturnUrl, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể khởi tạo OpenID"})
		return
	}

	c.Redirect(http.StatusFound, url)
}

// CallbackHandle godoc
// @Summary      Steam Callback
// @Description  Xử lý callback trả về từ Steam OpenID sau khi đăng nhập thành công
// @Tags         Authentication (Steam)
// @Success      200  {string}  string "HTML Script trả về Frontend"
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /auth/steam/callback [get]
func (h *SteamHandler) CallbackHandle(c *gin.Context) {
    scheme := "http"
    if c.Request.TLS != nil {
        scheme = "https"
    }
    fullURL := fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, c.Request.RequestURI)
    fmt.Printf("Full callback URL: %s\n", fullURL)
    id, err := openid.Verify(fullURL, discoveryCache, nonceStore)
    if err != nil {
        utils.Unauthorized(c, "Xác thực thất bại: "+err.Error())
        return
    }

	steamID64 := h.extractSteamID(id)

    result, err := h.authService.LoginWithSteam(steamID64) 
    if err != nil {
        handleAuthServiceError(c, err)
		fmt.Printf("LoginWithSteam error: %v\n", err)
        return
    } 

respBytes, _ := json.Marshal(result)
	respData := string(respBytes)

	//tạm hardacode localhost 5173
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
        <html>
            <body>
                <script>
                    // Gửi dữ liệu về cửa sổ mẹ (Vite)
                    window.opener.postMessage(%s, "http://localhost:5173");
                    
                    // Tự đóng cửa sổ popup này lại
                    window.close();
                </script>
            </body>
        </html>
    `, respData))
}


func (h *SteamHandler) extractSteamID(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}