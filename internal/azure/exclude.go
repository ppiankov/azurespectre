package azure

// shouldExcludeTags returns true if the resource should be excluded based on tag filters.
func shouldExcludeTags(resourceTags, excludeTags map[string]string) bool {
	if len(excludeTags) == 0 {
		return false
	}
	for k, v := range excludeTags {
		resVal, exists := resourceTags[k]
		if v == "" {
			if exists {
				return true
			}
		} else {
			if resVal == v {
				return true
			}
		}
	}
	return false
}
