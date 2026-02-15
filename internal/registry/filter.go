package registry

// FilterOptions specifies criteria for filtering sites.
type FilterOptions struct {
	Group    string // Filter by group name
	TagKey   string // Filter by tag key
	TagValue string // Filter by tag value (requires TagKey)
	HostType string // Filter by host type
}

// FilterSites returns sites matching all specified filter criteria.
// Empty filter fields are ignored (match all).
func FilterSites(sites []*Site, opts FilterOptions, userGroups map[string][]string) []*Site {
	var result []*Site
	for _, s := range sites {
		if matchesFilter(s, opts, userGroups) {
			result = append(result, s)
		}
	}
	return result
}

func matchesFilter(site *Site, opts FilterOptions, userGroups map[string][]string) bool {
	if opts.Group != "" && !MatchGroup(site, opts.Group, userGroups) {
		return false
	}
	if opts.HostType != "" && site.HostType != opts.HostType {
		return false
	}
	if opts.TagKey != "" {
		val, ok := site.Tags[opts.TagKey]
		if !ok {
			return false
		}
		if opts.TagValue != "" && val != opts.TagValue {
			return false
		}
	}
	return true
}
