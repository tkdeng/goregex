package regex

import (
	"io"
	"os"
	"regexp"
	"strconv"
)

// CompRE2 compiles an re2 regular expression and store it in the cache
func CompRE2(re string, params ...string) *RegexpRE2 {
	re = compRE(re, params)

	if val, err := cacheRE2.Get(re); val != nil || err != nil {
		if err != nil {
			panic(err)
		}

		return val
	}

	reg := regexp.MustCompile(re)

	compRe := RegexpRE2{RE: reg, len: int64(len(re))}

	cacheRE2.Set(re, &compRe, nil)
	return &compRe
}

// CompTryRE2 tries to compile re2 or returns an error
func CompTryRE2(re string, params ...string) (*RegexpRE2, error) {
	re = compRE(re, params)

	if val, err := cacheRE2.Get(re); val != nil || err != nil {
		if err != nil {
			return &RegexpRE2{}, err
		}

		return val, nil
	}

	reg, err := regexp.Compile(re)
	if err != nil {
		cacheRE2.Set(re, nil, err)
		return &RegexpRE2{}, err
	}

	compRe := RegexpRE2{RE: reg, len: int64(len(re))}

	cacheRE2.Set(re, &compRe, nil)
	return &compRe, nil
}

//* regex methods

// RepFunc replaces a string with the result of a function
//
// similar to JavaScript .replace(/re/, function(data){})
func (reg *RegexpRE2) RepFunc(str []byte, rep func(data func(int) []byte) []byte, blank ...bool) []byte {
	ind := reg.RE.FindAllIndex(str, -1)

	res := []byte{}
	trim := 0
	for _, pos := range ind {
		v := str[pos[0]:pos[1]]
		m := reg.RE.FindAllSubmatch(v, -1)

		if len(blank) != 0 {
			gCache := map[int][]byte{}
			r := rep(func(g int) []byte {
				if v, ok := gCache[g]; ok {
					return v
				}
				v := []byte{}
				if len(m[0]) > g {
					v = m[0][g]
				}
				gCache[g] = v
				return v
			})

			if []byte(r) == nil {
				return []byte{}
			}
		} else {
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
				v := []byte{}
				if len(m[0]) > g {
					v = m[0][g]
				}
				gCache[g] = v
				return v
			})

			if []byte(r) == nil {
				res = append(res, str[trim:]...)
				return res
			}

			res = append(res, r...)
		}
	}

	if len(blank) != 0 {
		return []byte{}
	}

	res = append(res, str[trim:]...)

	return res
}

// RepStrLit replaces a string with another string
//
// @rep uses the literal string, and does Not use args like $1
func (reg *RegexpRE2) RepStrLit(str []byte, rep []byte) []byte {
	return reg.RE.ReplaceAllLiteral(str, rep)
}

// RepStr is a more complex version of the RepStrLit method
//
// this function will replace things in the result like $1 with your capture groups
//
// use $0 to use the full regex capture group
//
// use ${123} to use numbers with more than one digit
func (reg *RegexpRE2) RepStr(str []byte, rep []byte) []byte {
	ind := reg.RE.FindAllIndex(str, -1)

	res := []byte{}
	trim := 0
	for _, pos := range ind {
		v := str[pos[0]:pos[1]]
		m := reg.RE.FindAllSubmatch(v, -1)

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
				if len(m[0]) > i {
					return m[0][i]
				}
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

// Match returns true if a []byte matches a regex
func (reg *RegexpRE2) Match(str []byte) bool {
	return reg.RE.Match(str)
}

// Split splits a string, and keeps capture groups
//
// Similar to JavaScript .split(/re/)
func (reg *RegexpRE2) Split(str []byte) [][]byte {
	ind := reg.RE.FindAllIndex(str, -1)

	res := [][]byte{}
	trim := 0
	for _, pos := range ind {
		v := str[pos[0]:pos[1]]
		m := reg.RE.FindAllSubmatch(v, -1)

		if trim == 0 {
			res = append(res, str[:pos[0]])
		} else {
			res = append(res, str[trim:pos[0]])
		}
		trim = pos[1]

		for i := 1; i < len(m[0]); i++ {
			g := m[0][i]
			if len(g) != 0 {
				res = append(res, m[0][i])
			}
		}
	}

	e := str[trim:]
	if len(e) != 0 {
		res = append(res, str[trim:])
	}

	return res
}

//* regex fs methods

// RepFileStr replaces a regex match with a new []byte in a file
//
// @all: if true, will replace all text matching @re,
// if false, will only replace the first occurrence
func (reg *RegexpRE2) RepFileStr(file *os.File, rep []byte, all bool, maxReSize ...int64) error {
	var found bool

	l := int64(reg.len * 10)
	if l < 1024 {
		l = 1024
	}
	for _, maxRe := range maxReSize {
		if l < maxRe {
			l = maxRe
		}
	}

	i := int64(0)

	buf := make([]byte, l)
	size, err := file.ReadAt(buf, i)
	buf = buf[:size]
	for err == nil {
		if reg.Match(buf) {
			found = true

			repRes := reg.RepStr(buf, rep)

			rl := int64(len(repRes))
			if rl == l {
				file.WriteAt(repRes, i)
				file.Sync()
			} else if rl < l {
				file.WriteAt(repRes, i)
				rl = l - rl

				j := i + l

				b := make([]byte, 1024)
				s, e := file.ReadAt(b, j)
				b = b[:s]

				for e == nil {
					file.WriteAt(b, j-rl)
					j += 1024
					b = make([]byte, 1024)
					s, e = file.ReadAt(b, j)
					b = b[:s]
				}

				if s != 0 {
					file.WriteAt(b, j-rl)
					j += int64(s)
				}

				file.Truncate(j - rl)
				file.Sync()
			} else if rl > l {
				rl -= l

				dif := int64(1024)
				if rl > dif {
					dif = rl
				}

				j := i + l

				b := make([]byte, dif)
				s, e := file.ReadAt(b, j)
				bw := b[:s]

				file.WriteAt(repRes, i)
				j += rl

				for e == nil {
					b = make([]byte, dif)
					s, e = file.ReadAt(b, j+dif-rl)

					file.WriteAt(bw, j)
					bw = b[:s]

					j += dif
				}

				file.WriteAt(bw, j)
				file.Sync()
			}

			if !all {
				file.Sync()
				return nil
			}

			i += int64(len(repRes))
		}

		i++
		buf = make([]byte, l)
		size, err = file.ReadAt(buf, i)
		buf = buf[:size]
	}

	if reg.Match(buf) {
		found = true

		repRes := reg.RepStr(buf, rep)

		rl := int64(len(repRes))
		if rl == l {
			file.WriteAt(repRes, i)
			file.Sync()
		} else if rl < l {
			file.WriteAt(repRes, i)
			rl = l - rl

			j := i + l

			b := make([]byte, 1024)
			s, e := file.ReadAt(b, j)
			b = b[:s]

			for e == nil {
				file.WriteAt(b, j-rl)
				j += 1024
				b = make([]byte, 1024)
				s, e = file.ReadAt(b, j)
				b = b[:s]
			}

			if s != 0 {
				file.WriteAt(b, j-rl)
				j += int64(s)
			}

			file.Truncate(j - rl)
			file.Sync()
		} else if rl > l {
			rl -= l

			dif := int64(1024)
			if rl > dif {
				dif = rl
			}

			j := i + l

			b := make([]byte, dif)
			s, e := file.ReadAt(b, j)
			bw := b[:s]

			file.WriteAt(repRes, i)
			j += rl

			for e == nil {
				b = make([]byte, dif)
				s, e = file.ReadAt(b, j+dif-rl)

				file.WriteAt(bw, j)
				bw = b[:s]

				j += dif
			}

			file.WriteAt(bw, j)
			file.Sync()
		}
	}

	file.Sync()

	if !found {
		return io.EOF
	}
	return nil
}

// RepFileFunc replaces a regex match with the result of a callback function in a file
//
// @all: if true, will replace all text matching @re,
// if false, will only replace the first occurrence
func (reg *RegexpRE2) RepFileFunc(file *os.File, rep func(data func(int) []byte) []byte, all bool, maxReSize ...int64) error {
	var found bool

	l := int64(reg.len * 10)
	if l < 1024 {
		l = 1024
	}
	for _, maxRe := range maxReSize {
		if l < maxRe {
			l = maxRe
		}
	}

	i := int64(0)

	buf := make([]byte, l)
	size, err := file.ReadAt(buf, i)
	buf = buf[:size]
	for err == nil {
		if reg.Match(buf) {
			found = true

			repRes := reg.RepFunc(buf, rep)

			rl := int64(len(repRes))
			if rl == l {
				file.WriteAt(repRes, i)
				file.Sync()
			} else if rl < l {
				file.WriteAt(repRes, i)
				rl = l - rl

				j := i + l

				b := make([]byte, 1024)
				s, e := file.ReadAt(b, j)
				b = b[:s]

				for e == nil {
					file.WriteAt(b, j-rl)
					j += 1024
					b = make([]byte, 1024)
					s, e = file.ReadAt(b, j)
					b = b[:s]
				}

				if s != 0 {
					file.WriteAt(b, j-rl)
					j += int64(s)
				}

				file.Truncate(j - rl)
				file.Sync()
			} else if rl > l {
				rl -= l

				dif := int64(1024)
				if rl > dif {
					dif = rl
				}

				j := i + l

				b := make([]byte, dif)
				s, e := file.ReadAt(b, j)
				bw := b[:s]

				file.WriteAt(repRes, i)
				j += rl

				for e == nil {
					b = make([]byte, dif)
					s, e = file.ReadAt(b, j+dif-rl)

					file.WriteAt(bw, j)
					bw = b[:s]

					j += dif
				}

				file.WriteAt(bw, j)
				file.Sync()
			}

			if !all {
				file.Sync()
				return nil
			}

			i += int64(len(repRes))
		}

		i++
		buf = make([]byte, l)
		size, err = file.ReadAt(buf, i)
		buf = buf[:size]
	}

	if reg.Match(buf) {
		found = true

		repRes := reg.RepFunc(buf, rep)

		rl := int64(len(repRes))
		if rl == l {
			file.WriteAt(repRes, i)
			file.Sync()
		} else if rl < l {
			file.WriteAt(repRes, i)
			rl = l - rl

			j := i + l

			b := make([]byte, 1024)
			s, e := file.ReadAt(b, j)
			b = b[:s]

			for e == nil {
				file.WriteAt(b, j-rl)
				j += 1024
				b = make([]byte, 1024)
				s, e = file.ReadAt(b, j)
				b = b[:s]
			}

			if s != 0 {
				file.WriteAt(b, j-rl)
				j += int64(s)
			}

			file.Truncate(j - rl)
			file.Sync()
		} else if rl > l {
			rl -= l

			dif := int64(1024)
			if rl > dif {
				dif = rl
			}

			j := i + l

			b := make([]byte, dif)
			s, e := file.ReadAt(b, j)
			bw := b[:s]

			file.WriteAt(repRes, i)
			j += rl

			for e == nil {
				b = make([]byte, dif)
				s, e = file.ReadAt(b, j+dif-rl)

				file.WriteAt(bw, j)
				bw = b[:s]

				j += dif
			}

			file.WriteAt(bw, j)
			file.Sync()
		}
	}

	file.Sync()

	if !found {
		return io.EOF
	}
	return nil
}
