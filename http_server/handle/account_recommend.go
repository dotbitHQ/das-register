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
	Page    int    `json:"page" binding:"required,gte=1"`
	Size    int    `json:"size" binding:"required,gte=5"`
}
type RepAccountRecommend struct {
	TotalPage int      `json:"total_page"`
	Page      int      `json:"page"`
	AccList   []string `json:"acc_list"`
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
	var resp RepAccountRecommend
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
	acc = strings.ToLower(strings.ReplaceAll(acc, common.DasAccountSuffix, ""))

	//tokens := strings.Split(strings.ToLower(acc), "-")
	//separate token
	tokens, separateType, err := h.separateToken(acc)
	if err != nil {
		err = fmt.Errorf("separateToken err: %s", err.Error())
		apiResp.ApiRespErr(http_api.ApiCodeError500, fmt.Sprintf("separateToken err"))
		return err
	}
	log.Info("tokens: ", tokens)
	//token recommend
	recommendTokens, err := h.tokenRecommend(tokens)
	log.Info("recommendTokens: ", recommendTokens)
	//Combine recommend
	length := len(tokens)
	recommendAcc := make([]string, 0)
	for i := 0; i < length; i++ {
		//0,1,2
		if i > 2 {
			break
		}
		//变i，其他不变
		for _, v := range recommendTokens[i] {
			tempToken := make([]string, length)
			copy(tempToken, tokens)
			tempToken[i] = v
			newWord := strings.ToLower(strings.Join(tempToken, separateType))
			fmt.Println("newword: ", newWord)

			charSet, err := h.dasCore.GetAccountCharSetList(newWord + common.DasAccountSuffix)
			if err != nil {
				fmt.Println("GetAccountCharSetList err: ", err.Error())
				continue
			}

			//check available
			if strings.Contains(newWord, ".") {
				continue
			}
			if newWord == acc {
				continue
			}
			tempReq := &ReqAccountSearch{
				Account:        newWord + common.DasAccountSuffix,
				AccountCharStr: charSet,
			}
			fmt.Println(222222222, newWord, charSet)
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

	//paging
	data, totalPage := h.pagingAcc(req.Page, req.Size, recommendAcc)
	resp.Page = req.Page
	resp.TotalPage = totalPage
	resp.AccList = data
	apiResp.ApiRespOK(resp)
	return nil
}

//	func (h *HttpHandle) checkRecommendAvailable(acc) bool{
//		charSet, err := h.dasCore.GetAccountCharSetList(acc + common.DasAccountSuffix)
//		if err != nil {
//			fmt.Println("GetAccountCharSetList err: ", err.Error())
//			continue
//		}
//		//check available
//		tempReq := &ReqAccountSearch{
//			Account:        newWord,
//			AccountCharStr: charSet,
//		}
//		var tempRep http_api.ApiResp
//		_, status, _, _ := h.checkAccountBase(tempReq, &tempRep)
//		if tempRep.ErrNo != http_api.ApiCodeSuccess {
//			continue
//		}
//		if status != tables.SearchStatusRegisterAble {
//			continue
//		}
//	}
func (h *HttpHandle) tokenRecommend(tokens []string) (recommendTokens [][]string, err error) {
	//recommendTokens := make([][]string, 0)
	for _, v := range tokens {
		recommendToken, err := h.es.FuzzyQueryAcc(v, len(v), 0)
		if err != nil {
			err = fmt.Errorf("FuzzyQueryAcc err: ", err.Error())
			return recommendTokens, err
		}
		recommendTokens = append(recommendTokens, recommendToken)
	}
	if len(recommendTokens) == 0 {
		fmt.Println("add length 000000000")
		for _, v := range tokens {
			recommendToken, err := h.es.FuzzyQueryAcc(v, 0, 0)
			if err != nil {
				//continue
				err = fmt.Errorf("FuzzyQueryAcc err: ", err.Error())
				return recommendTokens, err
			}
			recommendTokens = append(recommendTokens, recommendToken)
		}
	}
	return
}
func (h *HttpHandle) separateToken(acc string) (tokens []string, separateTag string, err error) {

	res, err := h.es.TermQueryAcc(acc)
	if err != nil {
		fmt.Println("TermQuery err:", err.Error())
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
					fmt.Println("TermQuery err:", err.Error())
				}
				if prefixAcc.Acc != "" {
					suffixAcc, err := h.es.TermQueryAcc(suffix)
					if err != nil {
						fmt.Println("TermQuery err:", err.Error())
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
