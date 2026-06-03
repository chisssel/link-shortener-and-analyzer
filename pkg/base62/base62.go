package base62

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Encode(id uint64) string {
	if id == 0 {
		return string(alphabet[0])
	}
	var buf [12]byte
	var i = len(buf)
	for id > 0 {
		i--
		buf[i] = alphabet[id%62]
		id /= 62
	}
	return string(buf[i:])
}

func Decode(s string) uint64 {
	var n uint64
	for _, c := range []byte(s) {
		n *= 62
		switch {
		case c >= '0' && c <= '9':
			n += uint64(c - '0')
		case c >= 'A' && c <= 'Z':
			n += uint64(c - 'A' + 10)
		case c >= 'a' && c <= 'z':
			n += uint64(c - 'a' + 36)
		default:
			return 0
		}
	}
	return n
}
