package example

import (
	"fmt"
	"github.com/scorpiotzh/toolib"
	"testing"
)

func TestAllowOriginFunc(t *testing.T) {
	list := []string{
		"https:\\/\\/[^.]*\\.bestdas\\.com",
		"https:\\/\\/[^.]*\\.da\\.systems",
		"https:\\/\\/bestdas\\.com",
		"https:\\/\\/da\\.systems",
		"https:\\/\\/app\\.gogodas\\.com",
	}
	fmt.Println(list)
	toolib.AllowOriginList = append(toolib.AllowOriginList, list...)
	ori := "https://da.systems"
	fmt.Println(toolib.AllowOriginFunc(ori))
}
