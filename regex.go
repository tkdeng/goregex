package regex

import (
	"bytes"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/GRbit/go-pcre"
	"github.com/tkdeng/goregex/common"
)

type PCRE pcre.Regexp
type RE2 *regexp.Regexp

type Regexp struct {
	RE  pcre.Regexp
	len int64
}

type RegexpRE2 struct {
	RE  *regexp.Regexp
	len int64
}

type bgPart struct {
	ref []byte
	b   []byte
}

var regCompCommentAndChars *regexp.Regexp = regexp.MustCompile(`(\\|)\(\?#.*?\)|%!|!%|\\[\\']`)
var regCompParam *regexp.Regexp = regexp.MustCompile(`(\\|)%(\{[0-9]+\}|[0-9])`)
var regCompBG *regexp.Regexp = regexp.MustCompile(`\[^?(\\[\\\]]|[^\]])+\]`)
var regCompBGRef *regexp.Regexp = regexp.MustCompile(`%!([0-9]+|o|c)!%`)

var regComplexSel *Regexp
var regEscape *Regexp

var cache common.CacheMap[*Regexp] = common.NewCache[*Regexp]()
var cacheRE2 common.CacheMap[*RegexpRE2] = common.NewCache[*RegexpRE2]()
var compCache common.CacheMap[[]byte] = common.NewCache[[]byte]()

func init() {
	regComplexSel = Comp(`(\\|)\$([0-9]|\{[0-9]+\})`)
	regEscape = Comp(`[\\\^\$\.\|\?\*\+\(\)\[\]\{\}\%]`)

	go func() {
		for {
			time.Sleep(10 * time.Minute)

			// default: remove cache items have not been accessed in over 2 hours
			cacheTime := 2 * time.Hour

			// SysFreeMemory returns the total free system memory in megabytes
			mb := common.SysFreeMemory()
			if mb < 200 && mb != 0 {
				// low memory: remove cache items have not been accessed in over 10 minutes
				cacheTime = 10 * time.Minute
			} else if mb < 500 && mb != 0 {
				// low memory: remove cache items have not been accessed in over 30 minutes
				cacheTime = 30 * time.Minute
			} else if mb < 2000 && mb != 0 {
				// low memory: remove cache items have not been accessed in over 1 hour
				cacheTime = 1 * time.Hour
			} else if mb > 64000 {
				// high memory: remove cache items have not been accessed in over 12 hour
				cacheTime = 12 * time.Hour
			} else if mb > 32000 {
				// high memory: remove cache items have not been accessed in over 6 hour
				cacheTime = 6 * time.Hour
			} else if mb > 16000 {
				// high memory: remove cache items have not been accessed in over 3 hour
				cacheTime = 3 * time.Hour
			}

			cache.DelOld(cacheTime)
			cacheRE2.DelOld(cacheTime)
			compCache.DelOld(cacheTime)

			time.Sleep(10 * time.Second)

			// clear cache if were still critically low on available memory
			if mb := common.SysFreeMemory(); mb < 10 && mb != 0 {
				cache.DelOld(0)
				cacheRE2.DelOld(0)
				compCache.DelOld(0)
			}
		}
	}()
}

// this method compiles the RE string to add more functionality to it
func compRE(re string, params []string) string {
	if val, err := compCache.Get(re); val != nil || err != nil {
		if err != nil {
			return ""
		}

		return string(regCompParam.ReplaceAllFunc(val, func(b []byte) []byte {
			if b[1] == '{' && b[len(b)-1] == '}' {
				b = b[2 : len(b)-1]
			} else {
				b = b[1:]
			}

			if n, e := strconv.Atoi(string(b)); e == nil && n > 0 && n <= len(params) {
				return []byte(Escape(params[n-1]))
			}
			return []byte{}
		}))
	}

	reB := []byte(re)

	reB = regCompCommentAndChars.ReplaceAllFunc(reB, func(b []byte) []byte {
		if bytes.Equal(b, []byte("%!")) {
			return []byte("%!o!%")
		} else if bytes.Equal(b, []byte("!%")) {
			return []byte("%!c!%")
		} else if b[0] == '\\' {
			if b[1] == '\'' {
				return []byte{'`'}
			}
			return b
		}
		return []byte{}
	})

	bgList := [][]byte{}
	reB = regCompBG.ReplaceAllFunc(reB, func(b []byte) []byte {
		bgList = append(bgList, b)
		return common.JoinBytes('%', '!', len(bgList)-1, '!', '%')
	})

	for ind, bgItem := range bgList {
		charS := []byte{'['}
		if bgItem[1] == '^' {
			bgItem = bgItem[2 : len(bgItem)-1]
			charS = append(charS, '^')
		} else {
			bgItem = bgItem[1 : len(bgItem)-1]
		}

		newBG := []bgPart{}
		for i := 0; i < len(bgItem); i++ {
			if i+1 < len(bgItem) {
				if bgItem[i] == '\\' {
					newBG = append(newBG, bgPart{ref: []byte{bgItem[i+1]}, b: []byte{bgItem[i], bgItem[i+1]}})
					i++
					continue
				} else if bgItem[i+1] == '-' && i+2 < len(bgItem) {
					newBG = append(newBG, bgPart{ref: []byte{bgItem[i], bgItem[i+2]}, b: []byte{bgItem[i], bgItem[i+1], bgItem[i+2]}})
					i += 2
					continue
				}
			}
			newBG = append(newBG, bgPart{ref: []byte{bgItem[i]}, b: []byte{bgItem[i]}})
		}

		sort.Slice(newBG, func(i, j int) bool {
			if len(newBG[i].ref) > len(newBG[j].ref) {
				return true
			} else if len(newBG[i].ref) < len(newBG[j].ref) {
				return false
			}

			for k := 0; k < len(newBG[i].ref); k++ {
				if newBG[i].ref[k] < newBG[j].ref[k] {
					return true
				} else if newBG[i].ref[k] > newBG[j].ref[k] {
					return false
				}
			}

			return false
		})

		bgItem = charS
		for i := 0; i < len(newBG); i++ {
			bgItem = append(bgItem, newBG[i].b...)
		}
		bgItem = append(bgItem, ']')

		bgList[ind] = bgItem
	}

	reB = regCompBGRef.ReplaceAllFunc(reB, func(b []byte) []byte {
		b = b[2 : len(b)-2]

		if b[0] == 'o' {
			return []byte(`%!`)
		} else if b[0] == 'c' {
			return []byte(`!%`)
		}

		if n, e := strconv.Atoi(string(b)); e == nil && n < len(bgList) {
			return bgList[n]
		}
		return []byte{}
	})

	compCache.Set(re, reB, nil)

	return string(regCompParam.ReplaceAllFunc(reB, func(b []byte) []byte {
		if b[1] == '{' && b[len(b)-1] == '}' {
			b = b[2 : len(b)-1]
		} else {
			b = b[1:]
		}

		if n, e := strconv.Atoi(string(b)); e == nil && n > 0 && n <= len(params) {
			return []byte(Escape(params[n-1]))
		}
		return []byte{}
	}))
}

//* regex compile methods

// Comp compiles a regular expression and store it in the cache
func Comp(re string, params ...string) *Regexp {
	re = compRE(re, params)

	if val, err := cache.Get(re); val != nil || err != nil {
		if err != nil {
			panic(err)
		}

		return val
	}

	reg := pcre.MustCompile(re, pcre.UTF8)

	// commented below methods compiled 10000 times in 0.1s (above method being used finished in half of that time)
	// reg := pcre.MustCompileParse(re)
	// reg := pcre.MustCompileJIT(re, pcre.UTF8, pcre.STUDY_JIT_COMPILE)
	// reg := pcre.MustCompileJIT(re, pcre.EXTRA, pcre.STUDY_JIT_COMPILE)
	// reg := pcre.MustCompileJIT(re, pcre.JAVASCRIPT_COMPAT, pcre.STUDY_JIT_COMPILE)
	// reg := pcre.MustCompileParseJIT(re, pcre.STUDY_JIT_COMPILE)

	compRe := Regexp{RE: reg, len: int64(len(re))}

	cache.Set(re, &compRe, nil)
	return &compRe
}

// CompTry tries to compile or returns an error
func CompTry(re string, params ...string) (*Regexp, error) {
	re = compRE(re, params)

	if val, err := cache.Get(re); val != nil || err != nil {
		if err != nil {
			return &Regexp{}, err
		}

		return val, nil
	}

	reg, err := pcre.Compile(re, pcre.UTF8)
	if err != nil {
		cache.Set(re, nil, err)
		return &Regexp{}, err
	}

	// commented below methods compiled 10000 times in 0.1s (above method being used finished in half of that time)
	// reg := pcre.MustCompileParse(re)
	// reg := pcre.MustCompileJIT(re, pcre.UTF8, pcre.STUDY_JIT_COMPILE)
	// reg := pcre.MustCompileJIT(re, pcre.EXTRA, pcre.STUDY_JIT_COMPILE)
	// reg := pcre.MustCompileJIT(re, pcre.JAVASCRIPT_COMPAT, pcre.STUDY_JIT_COMPILE)
	// reg := pcre.MustCompileParseJIT(re, pcre.STUDY_JIT_COMPILE)

	compRe := Regexp{RE: reg, len: int64(len(re))}

	cache.Set(re, &compRe, nil)
	return &compRe, nil
}

//* other regex methods

// Escape will escape regex special chars
func Escape(re string) string {
	return string(regEscape.RepStr([]byte(re), []byte(`\$1`)))
}

// IsValid will return true if a regex is valid and can be compiled by this module
func IsValid(re string) bool {
	re = compRE(re, []string{})
	if _, err := pcre.Compile(re, pcre.UTF8); err == nil {
		return true
	}
	return false
}

// IsValidPCRE will return true if a regex is valid and can be compiled by the PCRE module
func IsValidPCRE(re string) bool {
	if _, err := pcre.Compile(re, pcre.UTF8); err == nil {
		return true
	}
	return false
}

// IsValidRE2 will return true if a regex is valid and can be compiled by the builtin RE2 module
func IsValidRE2(re string) bool {
	if _, err := regexp.Compile(re); err == nil {
		return true
	}
	return false
}

// JoinBytes is an easy way to join multiple values into a single []byte
func JoinBytes(bytes ...interface{}) []byte {
	return common.JoinBytes(bytes...)
}
