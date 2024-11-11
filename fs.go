package regex

import (
	"io"
	"os"
)

// RepFileStr replaces a regex match with a new []byte in a file
//
// @all: if true, will replace all text matching @re,
// if false, will only replace the first occurrence
func (reg *Regexp) RepFileStr(file *os.File, rep []byte, all bool, maxReSize ...int64) error {
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
func (reg *Regexp) RepFileFunc(file *os.File, rep func(data func(int) []byte) []byte, all bool, maxReSize ...int64) error {
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

// MatchFile returns true if a file contains a regex match
func (reg *Regexp) MatchFile(file *os.File, maxReSize ...int64) bool {
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
			return true
		}

		i++
		buf = make([]byte, l)
		size, err = file.ReadAt(buf, i)
		buf = buf[:size]
	}

	return reg.Match(buf)
}
