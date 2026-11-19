package app

import "strings"

type CustomerServiceSetting struct {
	Enabled                bool   `json:"enabled"`
	Title                  string `json:"title"`
	Description            string `json:"description"`
	ContactName            string `json:"contactName"`
	Phone                  string `json:"phone"`
	Wechat                 string `json:"wechat"`
	QRCodeResourceID       string `json:"qrCodeResourceId"`
	SecondContactName      string `json:"secondContactName"`
	SecondPhone            string `json:"secondPhone"`
	SecondWechat           string `json:"secondWechat"`
	SecondQRCodeResourceID string `json:"secondQrCodeResourceId"`
}

type PublicCustomerServiceSetting struct {
	Enabled           bool   `json:"enabled"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	ContactName       string `json:"contactName"`
	Phone             string `json:"phone"`
	Wechat            string `json:"wechat"`
	QRCodeURL         string `json:"qrCodeUrl"`
	SecondContactName string `json:"secondContactName"`
	SecondPhone       string `json:"secondPhone"`
	SecondWechat      string `json:"secondWechat"`
	SecondQRCodeURL   string `json:"secondQrCodeUrl"`
}

func defaultCustomerServiceSetting() CustomerServiceSetting {
	return CustomerServiceSetting{Title: "联系客服"}
}

func normalizeCustomerService(value CustomerServiceSetting) CustomerServiceSetting {
	value.Title = normalizeAppearanceSingleLine(value.Title)
	value.Description = normalizeAppearanceCopy(value.Description)
	value.ContactName = normalizeAppearanceSingleLine(value.ContactName)
	value.Phone = normalizeAppearanceSingleLine(value.Phone)
	value.Wechat = normalizeAppearanceSingleLine(value.Wechat)
	value.QRCodeResourceID = strings.TrimSpace(value.QRCodeResourceID)
	value.SecondContactName = normalizeAppearanceSingleLine(value.SecondContactName)
	value.SecondPhone = normalizeAppearanceSingleLine(value.SecondPhone)
	value.SecondWechat = normalizeAppearanceSingleLine(value.SecondWechat)
	value.SecondQRCodeResourceID = strings.TrimSpace(value.SecondQRCodeResourceID)
	return value
}

func validateCustomerService(value CustomerServiceSetting) error {
	for _, field := range []struct {
		value string
		label string
		max   int
	}{
		{value.Title, "客服弹窗标题", 40},
		{value.Description, "客服说明", 1000},
		{value.ContactName, "客服名称", 80},
		{value.Phone, "客服电话", 80},
		{value.Wechat, "客服微信号", 80},
		{value.SecondContactName, "第二位客服名称", 80},
		{value.SecondPhone, "第二位客服电话", 80},
		{value.SecondWechat, "第二位客服微信号", 80},
	} {
		if err := validateAppearanceCopy(field.value, field.label, field.max, value.Enabled && field.label == "客服弹窗标题"); err != nil {
			return err
		}
	}
	if len(value.QRCodeResourceID) > 80 || len(value.SecondQRCodeResourceID) > 80 {
		return BadAuthRequest("客服二维码资源 ID 无效")
	}
	if value.Enabled && value.Description == "" && value.ContactName == "" && value.Phone == "" && value.Wechat == "" && value.QRCodeResourceID == "" && value.SecondContactName == "" && value.SecondPhone == "" && value.SecondWechat == "" && value.SecondQRCodeResourceID == "" {
		return BadAuthRequest("启用客服前请至少填写一项内容或上传二维码")
	}
	return nil
}

func publicCustomerService(value CustomerServiceSetting, revision string) PublicCustomerServiceSetting {
	if !value.Enabled {
		return PublicCustomerServiceSetting{}
	}
	result := PublicCustomerServiceSetting{
		Enabled: true, Title: value.Title, Description: value.Description,
		ContactName: value.ContactName, Phone: value.Phone, Wechat: value.Wechat,
		SecondContactName: value.SecondContactName, SecondPhone: value.SecondPhone, SecondWechat: value.SecondWechat,
	}
	if value.QRCodeResourceID != "" {
		result.QRCodeURL = appearanceAssetURL(AppearanceAssetCustomerServiceQR, revision)
	}
	if value.SecondQRCodeResourceID != "" {
		result.SecondQRCodeURL = appearanceAssetURL(AppearanceAssetCustomerServiceSecondQR, revision)
	}
	return result
}
