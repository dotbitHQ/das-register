package handle

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func (h *HttpHandle) GithubCallback(ctx *gin.Context) {
	queryParams := ctx.Request.URL.Query()

	// 打印所有 URL 参数
	fmt.Println("URL Parameters:")
	for key, values := range queryParams {
		fmt.Printf("%s: %s\n", key, values)
	}
}
