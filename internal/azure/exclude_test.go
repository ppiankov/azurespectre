package azure

import "testing"

func TestShouldExcludeTags(t *testing.T) {
	tags := map[string]string{"Environment": "production", "Team": "platform"}

	if !shouldExcludeTags(tags, map[string]string{"Environment": "production"}) {
		t.Error("should exclude Environment=production")
	}
	if shouldExcludeTags(tags, map[string]string{"Environment": "staging"}) {
		t.Error("should not exclude Environment=staging")
	}
}

func TestShouldExcludeTagsKeyOnly(t *testing.T) {
	tags := map[string]string{"DoNotDelete": "true"}

	if !shouldExcludeTags(tags, map[string]string{"DoNotDelete": ""}) {
		t.Error("should exclude by key presence")
	}
	if shouldExcludeTags(tags, map[string]string{"MissingKey": ""}) {
		t.Error("should not exclude missing key")
	}
}

func TestShouldExcludeTagsEmpty(t *testing.T) {
	if shouldExcludeTags(map[string]string{"a": "b"}, nil) {
		t.Error("empty exclude should not match")
	}
	if shouldExcludeTags(nil, map[string]string{"a": "b"}) {
		t.Error("empty resource tags should not match")
	}
}
