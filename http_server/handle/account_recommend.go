package handle

import (
	"context"
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
	//Page    int    `json:"page" binding:"required,gte=1"`
	//Size    int    `json:"size" binding:"required,gte=5"`
}
type RepAccountRecommend struct {
	//TotalPage int      `json:"total_page"`
	//Page      int      `json:"page"`
	AccList []string `json:"acc_list"`
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
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx.Request.Context())
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx.Request.Context())

	if err = h.doAccountRecommend(ctx.Request.Context(), &req, &apiResp); err != nil {
		log.Error("doAccountRecommend err:", err.Error(), funcName, clientIp, ctx.Request.Context())
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doAccountRecommend(ctx context.Context, req *ReqAccountRecommend, apiResp *http_api.ApiResp) error {
	var resp RepAccountRecommend
	//check top level acc
	acc := req.Account
	count := strings.Count(acc, ".")
	if count != 1 || !strings.HasSuffix(acc, common.DasAccountSuffix) {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	acc = strings.ToLower(strings.ReplaceAll(acc, common.DasAccountSuffix, ""))

	//separate token
	tokens, separateType, err := h.separateToken(acc)
	if err != nil {
		err = fmt.Errorf("separateToken err: %s", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeError500, fmt.Sprintf("separateToken err"))
		return err
	}
	log.Info(ctx, "tokens: ", tokens)
	//token recommend
	recommendTokens, err := h.tokenRecommend(tokens)
	if err != nil {
		err = fmt.Errorf("tokenRecommend err: %s", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeError500, fmt.Sprintf("tokenRecommend err"))
		return err
	}
	log.Info(ctx, "recommendTokens: ", recommendTokens)
	//Combine recommend
	length := len(tokens)
	recommendAcc := make([]string, 0)
	for i := 0; i < length; i++ {
		if i > 2 {
			break
		}
		for _, v := range recommendTokens[i] {
			tempToken := make([]string, length)
			copy(tempToken, tokens)
			tempToken[i] = v
			newWord := strings.ToLower(strings.Join(tempToken, separateType))
			fmt.Println("newword: ", newWord)
			available, err := h.checkRecommendAvailable(ctx, acc, newWord)
			if err != nil {
				log.Errorf("checkRecommendAvailable err: %s", err.Error(), ctx)
				continue
			}
			if !available {
				continue
			}
			recommendAcc = append(recommendAcc, newWord+common.DasAccountSuffix)
		}
	}

	//paging
	//data, totalPage := h.pagingAcc(req.Page, req.Size, recommendAcc)
	//resp.Page = req.Page
	//resp.TotalPage = totalPage
	if len(recommendAcc) == 0 {
		apiResp.ApiRespErr(http_api.ApiCodeRecommendAccEmpty, fmt.Sprintf("recommend acc is empty"))
		return nil
	}
	resp.AccList = recommendAcc
	apiResp.ApiRespOK(resp)
	return nil
}

func (h *HttpHandle) checkRecommendAvailable(ctx context.Context, acc, recommendAcc string) (available bool, err error) {
	charSet, err := h.dasCore.GetAccountCharSetList(recommendAcc + common.DasAccountSuffix)
	if err != nil {
		err = fmt.Errorf("GetAccountCharSetList err: %s", err.Error())
		return false, err
	}
	//check available
	if strings.Contains(recommendAcc, ".") {
		return false, nil
	}
	if recommendAcc == acc {
		return false, nil
	}
	tempReq := &ReqAccountSearch{
		Account:        recommendAcc + common.DasAccountSuffix,
		AccountCharStr: charSet,
	}
	var tempRep http_api.ApiResp
	_, status, _, _ := h.checkAccountBase(ctx, tempReq, &tempRep)
	if tempRep.ErrNo != http_api.ApiCodeSuccess {
		return false, nil
	}
	if status != tables.SearchStatusRegisterAble {
		return false, nil
	}
	return true, nil
}
func (h *HttpHandle) tokenRecommend(tokens []string) (recommendTokens [][]string, err error) {
	//recommendTokens := make([][]string, 0)
	for _, v := range tokens {
		recommendToken, err := h.es.FuzzyQueryAcc(v, len(v), 0)
		if err != nil {
			err = fmt.Errorf("FuzzyQueryAcc err: ", err.Error())
			return recommendTokens, err
		}
		recommendToken = lowerAndUnique(recommendToken)
		recommendTokens = append(recommendTokens, recommendToken)

	}
	if len(recommendTokens) == 0 {
		for _, v := range tokens {
			recommendToken, err := h.es.FuzzyQueryAcc(v, 0, 0)
			if err != nil {
				//continue
				err = fmt.Errorf("FuzzyQueryAcc err: ", err.Error())
				return recommendTokens, err
			}
			recommendToken = lowerAndUnique(recommendToken)
			recommendTokens = append(recommendTokens, recommendToken)

		}
	}
	return
}
func (h *HttpHandle) separateToken(acc string) (tokens []string, separateTag string, err error) {

	res, err := h.es.TermQueryAcc(acc)
	if err != nil {
		err = fmt.Errorf("TermQuery err: %s", err.Error())
		return tokens, separateTag, err
	}
	if res.Acc != "" {
		tokens = append(tokens, acc)
	} else {
		if strings.Contains(acc, "-") {
			separateTag = "-"
			tokens = strings.Split(acc, "-")

		} else {
			for i := 1; i < len(acc); i++ {
				prefix := acc[:i]
				suffix := acc[i:]
				fmt.Println(prefix, suffix)
				prefixAcc, err := h.es.TermQueryAcc(prefix)
				if err != nil {
					err = fmt.Errorf("TermQuery err:", err.Error())
					return tokens, separateTag, err
				}
				if prefixAcc.Acc != "" {
					suffixAcc, err := h.es.TermQueryAcc(suffix)
					if err != nil {
						err = fmt.Errorf("TermQuery err:", err.Error())
						return tokens, separateTag, err
					}
					if suffixAcc.Acc != "" {
						tokens = append(tokens, prefix, suffix)
						break
					}
				}
			}
		}
	}
	fmt.Println("tokens: ", tokens)
	return

}
func (h *HttpHandle) pagingAcc(page, size int, acc []string) (data []string, totalPage int) {
	data = make([]string, 0)
	if len(acc) == 0 {
		return
	}
	totalPage = len(acc) / size
	if len(acc)%size != 0 {
		totalPage += 1
	}
	start := (page - 1) * size

	if page < totalPage {
		data = acc[start : start+size]
	} else if page == totalPage {
		data = acc[start:]
	}
	return
}
func lowerAndUnique(data []string) []string {
	uniqueMap := make(map[string]bool)
	var newData []string
	for _, str := range data {
		lowerStr := strings.ToLower(str)
		if !uniqueMap[lowerStr] {
			uniqueMap[lowerStr] = true
			newData = append(newData, lowerStr)
		}
	}
	return newData
}
