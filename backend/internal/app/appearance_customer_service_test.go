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
		{Title: "客服", SecondContactName: strings.Repeat("字", 81)},
		{Title: "客服", SecondPhone: "invalid\x00"},
		{Title: "客服", SecondQRCodeResourceID: strings.Repeat("x", 81)},
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

func TestCustomerServiceSecondContactLifecycle(t *testing.T) {
	svc, db, dir, admin := newAppearanceTestService(t)
	resources := []model.Resource{
		{ID: "first-qr", UserID: admin.ID, Kind: "image", Status: model.ResourceStatusReady, Provider: "local", ObjectKey: "support/first.png", MimeType: "image/png"},
		{ID: "second-qr", UserID: admin.ID, Kind: "image", Status: model.ResourceStatusReady, Provider: "local", ObjectKey: "support/second.png", MimeType: "image/png"},
		{ID: "foreign-second-qr", UserID: "other-admin", Kind: "image", Status: model.ResourceStatusReady, Provider: "local", ObjectKey: "support/foreign.png", MimeType: "image/png"},
	}
	if err := db.Create(&resources).Error; err != nil {
		t.Fatal(err)
	}
	writeAppearanceResourceFixtures(t, dir, resources)
	value := defaultAppearanceSetting()
	value.CustomerService.Enabled = true
	value.CustomerService.SecondContactName = " 第二位客服 "
	value.CustomerService.SecondPhone = "400-000-0002"
	value.CustomerService.SecondWechat = "support-two"
	value.CustomerService.SecondQRCodeResourceID = "foreign-second-qr"
	if _, err := svc.UpdateAppearance(admin, value); err == nil {
		t.Fatal("accepted foreign second QR")
	}
	value.CustomerService.SecondQRCodeResourceID = "second-qr"
	// 第一位留空时也允许仅发布第二位客服。
	saved, err := svc.UpdateAppearance(admin, value)
	if err != nil {
		t.Fatal(err)
	}
	if saved.CustomerService.SecondContactName != "第二位客服" || saved.Public.CustomerService.SecondWechat != "support-two" || !strings.Contains(saved.Public.CustomerService.SecondQRCodeURL, "/api/public/appearance/assets/customer-service-second-qr?v=") {
		t.Fatalf("second contact projection: %#v", saved)
	}
	value = saved.AppearanceSetting
	value.CustomerService.QRCodeResourceID = "first-qr"
	if _, err := svc.UpdateAppearance(admin, value); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range []struct{ slot, id string }{{AppearanceAssetCustomerServiceQR, "first-qr"}, {AppearanceAssetCustomerServiceSecondQR, "second-qr"}} {
		stream, err := svc.OpenAppearanceAsset(candidate.slot, "")
		if err != nil {
			t.Fatal(err)
		}
		stream.Body.Close()
		if refs := svc.appearanceResourceReferences([]string{candidate.id}); len(refs[candidate.id]) != 1 {
			t.Fatalf("missing reference: %#v", refs)
		}
	}
	value.CustomerService.Enabled = false
	saved, err = svc.UpdateAppearance(admin, value)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Public.CustomerService.SecondQRCodeURL != "" || saved.Public.CustomerService.SecondPhone != "" || saved.CustomerService.SecondQRCodeResourceID != "second-qr" {
		t.Fatal("disabled contact lost its draft or leaked publicly")
	}
	if _, err := svc.PrepareAppearanceAssetDelivery(AppearanceAssetCustomerServiceSecondQR, ResourceAccessOptions{}, ""); err == nil {
		t.Fatal("disabled second QR was public")
	}
	if refs := svc.appearanceResourceReferences([]string{"second-qr"}); len(refs["second-qr"]) != 1 {
		t.Fatal("disabled second QR was unprotected")
	}
	value.CustomerService.Enabled = true
	value.CustomerService.SecondQRCodeResourceID = ""
	if _, err := svc.UpdateAppearance(admin, value); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.OpenAppearanceAsset(AppearanceAssetCustomerServiceSecondQR, ""); err == nil {
		t.Fatal("removed second QR was public")
	}
	if refs := svc.appearanceResourceReferences([]string{"first-qr", "second-qr"}); len(refs["second-qr"]) != 0 || len(refs["first-qr"]) != 1 {
		t.Fatal("removing second QR changed first reference")
	}
	if max, err := AppearanceAssetMaxBytes(AppearanceAssetCustomerServiceSecondQR); err != nil || max != 5<<20 {
		t.Fatalf("second upload limit: %d, %v", max, err)
	}
	if _, err := svc.UploadAppearanceAsset(admin, AppearanceAssetCustomerServiceSecondQR, multipartFileHeader(t, "fake.png", "image/png", []byte("not an image"))); err == nil {
		t.Fatal("accepted forged second QR")
	}
}
