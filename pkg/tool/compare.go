package tool

func versionOrdinal(version string) string {
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

func LessThan(version, target string) bool {
	v1, v2 := versionOrdinal(version), versionOrdinal(target)
	return v1 < v2
}

func GreaterThan(version, target string) bool {
	v1, v2 := versionOrdinal(version), versionOrdinal(target)
	return v1 > v2
}
