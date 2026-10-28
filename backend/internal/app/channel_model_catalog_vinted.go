package app

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Vinted 的主视频目录使用独立合同；首次拉取时渠道尚未配置任何模型协议。
func vintedModelCatalogURL(baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil || !strings.EqualFold(base.Hostname(), "vinted.cam") {
		return ""
	}
	if path := strings.TrimRight(base.Path, "/"); path != "" && path != "/v1" {
		return ""
	}
	base.Path, base.RawPath, base.RawQuery, base.Fragment = "/v1/catalog", "", "", ""
	return base.String()
}

func parseVintedModelCatalog(data []byte) ([]ChannelModelCatalogItem, error) {
	var payload struct {
		Status     string    `json:"status"`
		Revision   string    `json:"revision"`
		ValidUntil time.Time `json:"valid_until"`
		Models     []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Available bool   `json:"available"`
		} `json:"models"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, WrapAppError(http.StatusBadGateway, "Vinted 模型目录不是有效 JSON 或字段格式有误", err)
	}
	if payload.Status != "ready" || strings.TrimSpace(payload.Revision) == "" || !payload.ValidUntil.After(time.Now()) {
		return nil, NewAppError(http.StatusBadGateway, "Vinted 模型目录尚未就绪或已过期，请稍后重试")
	}
	seen := make(map[string]bool)
	catalog := make([]ChannelModelCatalogItem, 0, 3)
	for _, item := range payload.Models {
		// 仅导入当前插件支持且实时可用的精确 ID，不补旧别名或缓存条目。
		switch item.ID {
		case "multiref:sd-2-c4", "multiref:ali-wan3", "qsqs:wan3-720p-deal":
		default:
			continue
		}
		if !item.Available || seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		catalog = append(catalog, ChannelModelCatalogItem{ID: item.ID, DisplayName: strings.TrimSpace(item.Name), ModelType: "video"})
	}
	sort.Slice(catalog, func(i, j int) bool { return catalog[i].ID < catalog[j].ID })
	return catalog, nil
}
