package topology

func mergeTags (tags ...map[string]string) map[string]string{
	merged := make(map[string]string)

	for _, t := range tags {
		for k,v := range t {
			merged[k]=v
		}
	}

	return merged
}

