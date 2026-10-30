package app

import (
	"encoding/json"
	"testing"
)

func TestRemovedVideoPluginsAreReconciledOnStartup(t *testing.T) {
	dataDir := t.TempDir()
	center, err := newPluginRuntime(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	removed := []string{"vinted-seedance-video", "200t-seedance-video", "vinted-video", "200t.vinted.mi"}
	for _, id := range removed {
		if _, exists := center.plugins[id]; exists {
			t.Fatalf("removed plugin %s is still bundled", id)
		}
	}
	records, err := center.readRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range removed {
		records = append(records, pluginRegistryRecord{
			ID: id, Source: PluginOriginOfficial,
			Raw: json.RawMessage(`{"metadata":{"id":"` + id + `","enabled":true}}`),
		})
	}
	if err := center.writeRegistry(records); err != nil {
		t.Fatal(err)
	}
	center, err = newPluginRuntime(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	records, err = center.readRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range removed {
		if _, exists := center.plugins[id]; exists {
			t.Fatalf("removed plugin %s survived restart", id)
		}
		if _, exists := center.registry.Resolve(id); exists {
			t.Fatalf("removed provider %s is still available", id)
		}
		for _, record := range records {
			if record.ID == id {
				t.Fatalf("removed plugin %s remains in registry", id)
			}
		}
	}
	if _, exists := center.registry.Resolve("vinted-main-video"); exists {
		t.Fatal("removed Vinted main video provider is still available")
	}
	if _, exists := center.registry.Resolve("200t-vinted-mi"); exists {
		t.Fatal("removed Vinted MI provider is still available")
	}
}
