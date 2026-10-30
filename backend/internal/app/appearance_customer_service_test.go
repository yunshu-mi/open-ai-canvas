package app

import (
	"strings"
	"testing"

	"yingce/backend/internal/model"
)

func TestCustomerServicePublicationOwnershipAndReferences(t *testing.T) {
	svc, db, dir, admin := newAppearanceTestService(t)
	resources := []model.Resource{
		{ID: "support-qr", UserID: admin.ID, Kind: "image", Status: model.ResourceStatusReady, Provider: "local", ObjectKey: "support/qr.png", MimeType: "image/png"},
		{ID: "foreign-qr", UserID: "other-admin", Kind: "image", Status: model.ResourceStatusReady, Provider: "local", ObjectKey: "support/foreign.png", MimeType: "image/png"},
	}
	if err := db.Create(&resources).Error; err != nil {
		t.Fatal(err)
	}
	writeAppearanceResourceFixtures(t, dir, resources)
	value := defaultAppearanceSetting()
	value.CustomerService = CustomerServiceSetting{Enabled: true, Title: " 联系客服 ", Description: "扫码咨询\r\n工作日在线", Phone: "400-000-0000", Wechat: "support", QRCodeResourceID: "foreign-qr"}
	if _, err := svc.UpdateAppearance(admin, value); err == nil {
		t.Fatal("accepted another administrator's resource")
	}
	value.CustomerService.QRCodeResourceID = "support-qr"
	if _, err := svc.UpdateAppearance(&model.User{ID: "user", Role: model.UserRoleUser, Status: model.UserStatusActive}, value); err == nil {
		t.Fatal("non-admin changed customer service")
	}
	saved, err := svc.UpdateAppearance(admin, value)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Public.CustomerService.Title != "联系客服" || saved.Public.CustomerService.Description != "扫码咨询\n工作日在线" || !strings.Contains(saved.Public.CustomerService.QRCodeURL, "/api/public/appearance/assets/customer-service-qr?v=") {
		t.Fatalf("public: %#v", saved.Public.CustomerService)
	}
	stream, err := svc.OpenAppearanceAsset(AppearanceAssetCustomerServiceQR, "")
	if err != nil {
		t.Fatal(err)
	}
	stream.Body.Close()
	value = saved.AppearanceSetting
	value.CustomerService.Enabled = false
	saved, err = svc.UpdateAppearance(admin, value)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Public.CustomerService.Enabled || saved.Public.CustomerService.Phone != "" || saved.Public.CustomerService.QRCodeURL != "" || saved.CustomerService.QRCodeResourceID != "support-qr" {
		t.Fatalf("disabled: %#v", saved)
	}
	if _, err := svc.OpenAppearanceAsset(AppearanceAssetCustomerServiceQR, ""); err == nil {
		t.Fatal("disabled QR was public")
	}
	if _, err := svc.PrepareAppearanceAssetDelivery(AppearanceAssetCustomerServiceQR, ResourceAccessOptions{}, ""); err == nil {
		t.Fatal("disabled QR delivery was public")
	}
	if refs := svc.appearanceResourceReferences([]string{"support-qr"}); len(refs["support-qr"]) != 1 {
		t.Fatalf("QR reference not protected: %#v", refs)
	}
	value.CustomerService.Enabled = true
	saved, err = svc.UpdateAppearance(admin, value)
	if err != nil || !saved.Public.CustomerService.Enabled {
		t.Fatalf("reenable: %v", err)
	}
	if _, err := svc.ResetAppearance(admin); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.OpenAppearanceAsset(AppearanceAssetCustomerServiceQR, ""); err == nil {
		t.Fatal("reset QR was public")
	}
}

func TestCustomerServiceValidationAndUpload(t *testing.T) {
	svc, _, _, admin := newAppearanceTestService(t)
	for _, support := range []CustomerServiceSetting{
		{Enabled: true, Title: "客服"},
		{Enabled: true, Description: "帮助"},
		{Title: strings.Repeat("字", 41)},
		{Title: "客服", Phone: "invalid\x00"},
		{Title: "客服", QRCodeResourceID: strings.Repeat("x", 81)},
	} {
		value := defaultAppearanceSetting()
		value.CustomerService = support
		if _, err := svc.UpdateAppearance(admin, value); err == nil {
			t.Fatalf("accepted invalid support: %#v", support)
		}
	}
	if _, err := svc.UploadAppearanceAsset(admin, AppearanceAssetCustomerServiceQR, multipartFileHeader(t, "fake.png", "image/png", []byte("not an image"))); err == nil {
		t.Fatal("accepted forged QR image")
	}
	if max, err := AppearanceAssetMaxBytes(AppearanceAssetCustomerServiceQR); err != nil || max != 5<<20 {
		t.Fatalf("upload limit: %d, %v", max, err)
	}
	if err := validateAppearanceResourceType(AppearanceAssetCustomerServiceQR, &model.Resource{Kind: "video", Status: model.ResourceStatusReady, MimeType: "video/mp4"}); err == nil {
		t.Fatal("accepted video as QR")
	}
}
