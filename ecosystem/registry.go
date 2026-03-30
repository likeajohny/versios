package ecosystem

var registry []Ecosystem

func Register(e Ecosystem) {
	registry = append(registry, e)
}

func DetectAll(dir string) []Ecosystem {
	var found []Ecosystem
	for _, e := range registry {
		if e.Detect(dir) {
			found = append(found, e)
		}
	}
	return found
}
