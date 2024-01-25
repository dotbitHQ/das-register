package handle

import (
	"das_register_server/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"net/http"
	"strings"
)

type ReqAccountRecommend struct {
	Account string `json:"account" binding:"required"`
}

func (h *HttpHandle) AccountRecommend(ctx *gin.Context) {
	var (
		funcName = "AccountRecommend"
		clientIp = GetClientIp(ctx)
		req      ReqAccountRecommend
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doAccountRecommend(&req, &apiResp); err != nil {
		log.Error("doAccountRecommend err:", err.Error(), funcName, clientIp, ctx)
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountRecommend(req *ReqAccountRecommend, apiResp *http_api.ApiResp) error {

	//check top level acc
	acc := req.Account
	count := strings.Count(acc, ".")
	if count != 1 || !strings.HasSuffix(acc, common.DasAccountSuffix) {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return nil
	}

	count = strings.Count(acc, ".")
	if count != 1 {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return nil

	}
	acc = strings.ReplaceAll(acc, common.DasAccountSuffix, "")

	tokens := strings.Split(strings.ToLower(acc), "-")
	recommendTokens := make([][]string, 0)
	for _, v := range tokens {
		//
		recommendToken, err := h.es.FuzzyQueryAcc(v)
		if err != nil {
			//continue
			apiResp.ApiRespErr(http_api.ApiCodeError500, fmt.Sprintf("FuzzyQueryAcc err: %s", err.Error()))
			return fmt.Errorf("fuzzyQuery err : %s", err.Error())
		}
		recommendTokens = append(recommendTokens, recommendToken)
	}
	fmt.Println(recommendTokens)
	//推荐组合
	length := len(tokens)
	recommendAcc := make([]string, 0)
	for i := 0; i < length; i++ {
		//0,1,2
		if i > 2 {
			break
		}
		//recommend token,
		//变i，其他不变
		for _, v := range recommendTokens[i] {
			tempToken := make([]string, length)
			copy(tempToken, tokens)
			tempToken[i] = v
			newWord := strings.ToLower(strings.Join(tempToken, "-"))
			fmt.Println("newword: ", newWord)
			charSet, err := h.dasCore.GetAccountCharSetList(acc + common.DasAccountSuffix)
			if err != nil {
				fmt.Println("GetAccountCharSetList err: ", err.Error())
				continue
			}
			//check available

			tempReq := &ReqAccountSearch{
				Account:        newWord,
				AccountCharStr: charSet,
			}
			var tempRep http_api.ApiResp
			_, status, _, _ := h.checkAccountBase(tempReq, &tempRep)
			if tempRep.ErrNo != http_api.ApiCodeSuccess {
				continue
			}
			if status != tables.SearchStatusRegisterAble {
				continue
			}
			recommendAcc = append(recommendAcc, newWord+common.DasAccountSuffix)
		}
	}

	//check
	apiResp.ApiRespOK(recommendAcc)
	return nil
}
