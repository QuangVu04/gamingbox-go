package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"vault/be/config"
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

func (h *SteamHandler) LoginHandle(c *gin.Context) {
	url, err := openid.RedirectURL(config.App.SteamOpenIdUrl, config.App.ReturnUrl, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể khởi tạo OpenID"})
		return
	}

	c.Redirect(http.StatusFound, url)
}

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

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
        <html>
            <body>
                <script>
                    // Gửi dữ liệu về cửa sổ mẹ (Angular)
                    window.opener.postMessage(%s, "%s");
                    
                    // Tự đóng cửa sổ popup này lại
                    window.close();
                </script>
            </body>
        </html>
    `, respData, config.App.FrontEndUrl))
}


func (h *SteamHandler) extractSteamID(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}