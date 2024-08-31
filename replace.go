package regex

import "strconv"

// RepFunc replaces a string with the result of a function
//
// similar to JavaScript .replace(/re/, function(data){})
func (reg *Regexp) RepFunc(str []byte, rep func(data func(int) []byte) []byte) []byte {
	ind := reg.RE.FindAllIndex(str, 0)

	res := []byte{}
	trim := 0
	for _, pos := range ind {
		v := str[pos[0]:pos[1]]
		m := reg.RE.NewMatcher(v, 0)

		if trim == 0 {
			res = append(res, str[:pos[0]]...)
		} else {
			res = append(res, str[trim:pos[0]]...)
		}
		trim = pos[1]

		gCache := map[int][]byte{}
		r := rep(func(g int) []byte {
			if v, ok := gCache[g]; ok {
				return v
			}
			v := m.Group(g)
			gCache[g] = v
			return v
		})

		if []byte(r) == nil {
			res = append(res, str[trim:]...)
			return res
		}

		res = append(res, r...)
	}

	res = append(res, str[trim:]...)

	return res
}

// RepStrLit replaces a string with another string
//
// note: this function is optimized for performance, and the replacement string does not accept replacements like $1
func (reg *Regexp) RepStrLit(str []byte, rep []byte) []byte {
	return reg.RE.ReplaceAll(str, rep, 0)
}

// RepStr is a more complex version of the RepStrLit method
//
// this function will replace things in the result like $1 with your capture groups
//
// use $0 to use the full regex capture group
//
// use ${123} to use numbers with more than one digit
func (reg *Regexp) RepStr(str []byte, rep []byte) []byte {
	ind := reg.RE.FindAllIndex(str, 0)

	res := []byte{}
	trim := 0
	for _, pos := range ind {
		v := str[pos[0]:pos[1]]
		m := reg.RE.NewMatcher(v, 0)

		if trim == 0 {
			res = append(res, str[:pos[0]]...)
		} else {
			res = append(res, str[trim:pos[0]]...)
		}
		trim = pos[1]

		r := regComplexSel.RepFunc(rep, func(data func(int) []byte) []byte {
			if len(data(1)) != 0 {
				return data(0)
			}
			n := data(2)
			if len(n) > 1 {
				n = n[1 : len(n)-1]
			}
			if i, err := strconv.Atoi(string(n)); err == nil {
				return m.Group(i)
			}
			return []byte{}
		})

		if r == nil {
			res = append(res, str[trim:]...)
			return res
		}

		res = append(res, r...)
	}

	res = append(res, str[trim:]...)

	return res
}
