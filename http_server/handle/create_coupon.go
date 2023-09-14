package handle

import (
	"crypto/sha256"
	"das_register_server/config"
	"das_register_server/tables"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	api_code "github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var letters = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

type Coupon struct {
	CouponType tables.CouponType `json:"type"`
	Num        int               `json:"num"`
	FileName   string            `json:"file_name"`
	FileAddr   string            `json:"file_addr"`
}
type ReqCreateCoupon struct {
	Desc         string   `json:"desc"`
	ExpireTime   string   `json:"expire_time"`
	StartTime    string   `json:"start_time"`
	CouponsGroup []Coupon `json:"coupons"`
}
type RespCreateCoupon struct {
	Coupons []Coupon `json:"coupons"`
}

func (h *HttpHandle) CreateCoupon(ctx *gin.Context) {
	var (
		funcName = "CreateCoupon"
		clientIp = GetClientIp(ctx)
		req      ReqCreateCoupon
		apiResp  api_code.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req), ctx)

	if err = h.doCreateCoupon(&req, &apiResp); err != nil {
		log.Error("doCreateCoupon err:", err.Error(), funcName, clientIp, ctx)
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCreateCoupon(req *ReqCreateCoupon, apiResp *api_code.ApiResp) error {
	var resp RespCreateCoupon
	salt := config.Cfg.Server.CouponEncrySalt
	filePath := config.Cfg.Server.CouponFilePath
	qrcodePrefix := config.Cfg.Server.CouponQrcodePrefix
	codeLength := config.Cfg.Server.CouponCodeLength
	if salt == "" || filePath == "" || qrcodePrefix == "" || codeLength == 0 {
		apiResp.ApiRespErr(api_code.ApiCodeError500, "config err")
		return fmt.Errorf("coupon config error")
	}
	if len(req.CouponsGroup) == 0 || req.Desc == "" || req.ExpireTime == "" || req.StartTime == "" {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	startAt, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, time.Local)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	expireAt, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExpireTime, time.Local)
	if err != nil {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	if startAt.After(expireAt) {
		apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
		return nil
	}
	for i, _ := range req.CouponsGroup {
		num := req.CouponsGroup[i].Num
		coupon_type := req.CouponsGroup[i].CouponType
		if num <= 0 || coupon_type <= 0 {
			apiResp.ApiRespErr(api_code.ApiCodeParamsInvalid, "params invalid")
			return nil
		}

		var fileData [][]string
		var tableData []tables.TableCoupon
		for j := 0; j < num; j++ {
			code := randStr(codeLength)
			hashCode := couponEncry(code, salt)
			rowData := []string{fmt.Sprintf("%s%s", qrcodePrefix, code), fmt.Sprintf("%s", req.Desc)}
			fileData = append(fileData, rowData)
			tableData = append(tableData, tables.TableCoupon{
				Code:       hashCode,
				CouponType: coupon_type,
				Desc:       req.Desc,
				ExpiredAt:  expireAt,
				StartAt:    startAt,
			})
		}

		fileName := fmt.Sprintf("/coupon-%s.csv", req.CouponsGroup[i].FileName)
		req.CouponsGroup[i].FileAddr = filePath + fileName

		if err := writeCouponCsv(filePath, fileName, fileData); err != nil {
			log.Error("Write file err:", err.Error())
			apiResp.ApiRespErr(api_code.ApiCodeError500, "Write file err")
			return err
		}

		if err := h.dbDao.CreateCoupon(tableData); err != nil {
			log.Error("CreateCoupon err:", err.Error())
			apiResp.ApiRespErr(api_code.ApiCodeError500, "create coupon fail")
			return err
		}

	}
	resp.Coupons = req.CouponsGroup
	apiResp.ApiRespOK(resp)
	return nil
}

func writeCouponCsv(filePath, fileName string, data [][]string) error {
	if err := checkPath(filePath); err != nil {
		return err
	}
	file, err := os.Create(filePath + "/" + fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	writer.Comma = ','
	writer.Write([]string{"qrcode_content"})
	for i, _ := range data {
		writer.Write(data[i])
	}
	writer.Flush()
	return nil
}

func randStr(n uint8) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
func couponEncry(str, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(str + salt))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}
func checkPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		err := os.MkdirAll(path, 0766)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
