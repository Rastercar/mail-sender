package array

func RemoveDuplicates(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, item := range s {
		if _, value := keys[item]; !value {
			keys[item] = true
			list = append(list, item)
		}
	}

	return list
}
