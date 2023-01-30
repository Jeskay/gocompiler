package lexer

import (
	"errors"
	"unsafe"
)

type decimal struct {
	digits     [800]byte
	length     int
	pointIndex int
	truncated  bool
}

const bias = -127
const exponentBits = 8
const mantissBits = 23

// decimal power of ten to binary power of two.
var powtab = []int{1, 3, 6, 9, 13, 16, 19, 23, 26}

// Maximum shift that we can do in one pass without overflow.
// A uint has 32 or 64 bits, and we have to be able to accommodate 9<<k.
const uintSize = 32 << (^uint(0) >> 63)
const maxShift = uintSize - 4

type leftCheat struct {
	delta  int    // number of new digits
	cutoff string // minus one digit if original < a.
}

var leftcheats = []leftCheat{
	// Leading digits of 1/2^i = 5^i.
	// 5^23 is not an exact 64-bit floating point number,
	// so have to use bc for the math.
	// Go up to 60 to be large enough for 32bit and 64bit platforms.
	/*
		seq 60 | sed 's/^/5^/' | bc |
		awk 'BEGIN{ print "\t{ 0, \"\" }," }
		{
			log2 = log(2)/log(10)
			printf("\t{ %d, \"%s\" },\t// * %d\n",
				int(log2*NR+1), $0, 2**NR)
		}'
	*/
	{0, ""},
	{1, "5"},                                           // * 2
	{1, "25"},                                          // * 4
	{1, "125"},                                         // * 8
	{2, "625"},                                         // * 16
	{2, "3125"},                                        // * 32
	{2, "15625"},                                       // * 64
	{3, "78125"},                                       // * 128
	{3, "390625"},                                      // * 256
	{3, "1953125"},                                     // * 512
	{4, "9765625"},                                     // * 1024
	{4, "48828125"},                                    // * 2048
	{4, "244140625"},                                   // * 4096
	{4, "1220703125"},                                  // * 8192
	{5, "6103515625"},                                  // * 16384
	{5, "30517578125"},                                 // * 32768
	{5, "152587890625"},                                // * 65536
	{6, "762939453125"},                                // * 131072
	{6, "3814697265625"},                               // * 262144
	{6, "19073486328125"},                              // * 524288
	{7, "95367431640625"},                              // * 1048576
	{7, "476837158203125"},                             // * 2097152
	{7, "2384185791015625"},                            // * 4194304
	{7, "11920928955078125"},                           // * 8388608
	{8, "59604644775390625"},                           // * 16777216
	{8, "298023223876953125"},                          // * 33554432
	{8, "1490116119384765625"},                         // * 67108864
	{9, "7450580596923828125"},                         // * 134217728
	{9, "37252902984619140625"},                        // * 268435456
	{9, "186264514923095703125"},                       // * 536870912
	{10, "931322574615478515625"},                      // * 1073741824
	{10, "4656612873077392578125"},                     // * 2147483648
	{10, "23283064365386962890625"},                    // * 4294967296
	{10, "116415321826934814453125"},                   // * 8589934592
	{11, "582076609134674072265625"},                   // * 17179869184
	{11, "2910383045673370361328125"},                  // * 34359738368
	{11, "14551915228366851806640625"},                 // * 68719476736
	{12, "72759576141834259033203125"},                 // * 137438953472
	{12, "363797880709171295166015625"},                // * 274877906944
	{12, "1818989403545856475830078125"},               // * 549755813888
	{13, "9094947017729282379150390625"},               // * 1099511627776
	{13, "45474735088646411895751953125"},              // * 2199023255552
	{13, "227373675443232059478759765625"},             // * 4398046511104
	{13, "1136868377216160297393798828125"},            // * 8796093022208
	{14, "5684341886080801486968994140625"},            // * 17592186044416
	{14, "28421709430404007434844970703125"},           // * 35184372088832
	{14, "142108547152020037174224853515625"},          // * 70368744177664
	{15, "710542735760100185871124267578125"},          // * 140737488355328
	{15, "3552713678800500929355621337890625"},         // * 281474976710656
	{15, "17763568394002504646778106689453125"},        // * 562949953421312
	{16, "88817841970012523233890533447265625"},        // * 1125899906842624
	{16, "444089209850062616169452667236328125"},       // * 2251799813685248
	{16, "2220446049250313080847263336181640625"},      // * 4503599627370496
	{16, "11102230246251565404236316680908203125"},     // * 9007199254740992
	{17, "55511151231257827021181583404541015625"},     // * 18014398509481984
	{17, "277555756156289135105907917022705078125"},    // * 36028797018963968
	{17, "1387778780781445675529539585113525390625"},   // * 72057594037927936
	{18, "6938893903907228377647697925567626953125"},   // * 144115188075855872
	{18, "34694469519536141888238489627838134765625"},  // * 288230376151711744
	{18, "173472347597680709441192448139190673828125"}, // * 576460752303423488
	{19, "867361737988403547205962240695953369140625"}, // * 1152921504606846976
}

func prefixIsLessThan(b []byte, s string) bool {
	for i := 0; i < len(s); i++ {
		if i >= len(b) {
			return true
		}
		if b[i] != s[i] {
			return b[i] < s[i]
		}
	}
	return false
}

func rightShift(d *decimal, k uint) {
	r := 0 // read pointer
	w := 0 // write pointer

	// Pick up enough leading digits to cover first shift.
	var n uint
	for ; n>>k == 0; r++ {
		if r >= d.length {
			if n == 0 {
				// a == 0; shouldn't get here, but handle anyway.
				d.length = 0
				return
			}
			for n>>k == 0 {
				n = n * 10
				r++
			}
			break
		}
		c := uint(d.digits[r])
		n = n*10 + c - '0'
	}
	d.pointIndex -= r - 1

	var mask uint = (1 << k) - 1

	// Pick up a digit, put down a digit.
	for ; r < d.length; r++ {
		c := uint(d.digits[r])
		dig := n >> k
		n &= mask
		d.digits[w] = byte(dig + '0')
		w++
		n = n*10 + c - '0'
	}

	// Put down extra digits.
	for n > 0 {
		dig := n >> k
		n &= mask
		if w < len(d.digits) {
			d.digits[w] = byte(dig + '0')
			w++
		} else if dig > 0 {
			d.truncated = true
		}
		n = n * 10
	}

	d.length = w
	trim(d)
}

func leftShift(d *decimal, k uint) {
	delta := leftcheats[k].delta
	if prefixIsLessThan(d.digits[0:d.length], leftcheats[k].cutoff) {
		delta--
	}

	r := d.length         // read index
	w := d.length + delta // write index

	// Pick up a digit, put down a digit.
	var n uint
	for r--; r >= 0; r-- {
		n += (uint(d.digits[r]) - '0') << k
		quo := n / 10
		rem := n - 10*quo
		w--
		if w < len(d.digits) {
			d.digits[w] = byte(rem + '0')
		} else if rem != 0 {
			d.truncated = true
		}
		n = quo
	}

	// Put down extra digits.
	for n > 0 {
		quo := n / 10
		rem := n - 10*quo
		w--
		if w < len(d.digits) {
			d.digits[w] = byte(rem + '0')
		} else if rem != 0 {
			d.truncated = true
		}
		n = quo
	}

	d.length += delta
	if d.length >= len(d.digits) {
		d.length = len(d.digits)
	}
	d.pointIndex += delta
	trim(d)
}

func trim(d *decimal) {
	for d.length > 0 && d.digits[d.length-1] == '0' {
		d.length--
	}
	if d.length == 0 {
		d.pointIndex = 0
	}
}

func (d *decimal) Shift(k int) {
	switch {
	case k > 0:
		for k > maxShift {
			leftShift(d, maxShift)
			k -= maxShift
		}
		leftShift(d, uint(k))
	case k < 0:
		for k < -maxShift {
			rightShift(d, maxShift)
			k += maxShift
		}
		rightShift(d, uint(-k))
	}
}

func shouldRoundUp(d *decimal, numberOfDigits int) bool {
	if numberOfDigits < 0 || numberOfDigits >= d.length {
		return false
	}
	if d.digits[numberOfDigits] == '5' && numberOfDigits+1 == d.length { // exactly halfway - round to even
		// if we truncated, a little higher than what's recorded - always round up
		if d.truncated {
			return true
		}
		return numberOfDigits > 0 && (d.digits[numberOfDigits-1]-'0')%2 != 0
	}
	// not halfway - digit tells all
	return d.digits[numberOfDigits] >= '5'
}

func (d *decimal) RoundedInteger() uint64 {
	if d.pointIndex > 20 {
		return 0xFFFFFFFFFFFFFFFF
	}
	var i int
	n := uint64(0)
	for i = 0; i < d.pointIndex && i < d.length; i++ {
		n = n*10 + uint64(d.digits[i]-'0')
	}
	for ; i < d.pointIndex; i++ {
		n *= 10
	}
	if shouldRoundUp(d, d.pointIndex) {
		n++
	}
	return n
}

func (d *decimal) fromString(s string) error {
	i := 0
	for ; i < len(s); i++ {
		switch {
		case s[i] == '_':
			continue
		case s[i] == '.':
			d.pointIndex = d.length
			continue
		case RuneInBase(10, rune(s[i])):
			if s[i] == '0' && d.length == 0 {
				d.pointIndex--
				continue
			}
			if d.length < len(d.digits) {
				d.digits[d.length] = s[i]
				d.length++
			} else if s[i] != '0' {
				d.truncated = true
			}
			continue
		}
		break
	}

	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i >= len(s) {
			return nil
		}
		esign := 1
		if s[i] == '+' {
			i++
		} else if s[i] == '-' {
			i++
			esign = -1
		}
		if i >= len(s) || !IsDigit(rune(s[i])) {
			return nil
		}
		e := 0
		for ; i < len(s) && (IsDigit(rune(s[i])) || s[i] == '_'); i++ {
			if s[i] == '_' {
				continue
			}
			if e < 10000 {
				e = e*10 + int(s[i]) - '0'
			}
		}
		d.pointIndex += e * esign
	}
	if i != len(s) {
		return errors.New("invalid format")
	}
	return nil
}

func (d *decimal) floatBits() (b uint64, overflow bool) {
	var exp int
	var mant uint64

	if d.length == 0 {
		mant = 0
		exp = bias
		goto out
	}

	if d.pointIndex > 310 {
		goto overflow
	}
	if d.pointIndex < -330 {
		// zero
		mant = 0
		exp = bias
		goto out
	}

	exp = 0
	for d.pointIndex > 0 {
		var n int
		if d.pointIndex >= len(powtab) {
			n = 27
		} else {
			n = powtab[d.pointIndex]
		}
		d.Shift(-n)
		exp += n
	}
	for d.pointIndex < 0 || d.pointIndex == 0 && d.digits[0] < '5' {
		var n int
		if -d.pointIndex >= len(powtab) {
			n = 27
		} else {
			n = powtab[-d.pointIndex]
		}
		d.Shift(n)
		exp -= n
	}

	// Our range is [0.5,1) but floating point range is [1,2).
	exp--

	// Minimum representable exponent is flt.bias+1.
	// If the exponent is smaller, move it up and
	// adjust d accordingly.
	if exp < bias+1 {
		n := bias + 1 - exp
		d.Shift(-n)
		exp += n
	}

	if exp-bias >= 1<<exponentBits-1 {
		goto overflow
	}

	// Extract 1+flt.mantbits bits.
	d.Shift(int(1 + mantissBits))
	mant = d.RoundedInteger()

	// Rounding might have added a bit; shift down.
	if mant == 2<<mantissBits {
		mant >>= 1
		exp++
		if exp-bias >= 1<<exponentBits-1 {
			goto overflow
		}
	}

	// Denormalized?
	if mant&(1<<mantissBits) == 0 {
		exp = bias
	}
	goto out

overflow:
	// ±Inf
	mant = 0
	exp = 1<<exponentBits - 1 + bias
	overflow = true

out:
	// Assemble bits.
	bits := mant & (uint64(1)<<mantissBits - 1)
	bits |= uint64((exp-bias)&(1<<exponentBits-1)) << mantissBits
	return bits, overflow
}

func bitsFromHex(s string, mantissa uint64, exponent int, truncate bool) (float64, error) {

	maxExp := 1<<exponentBits + bias - 2
	minExp := bias + 1
	exponent += int(mantissBits)

	for mantissa != 0 && mantissa>>(mantissBits+2) == 0 {
		mantissa <<= 1
		exponent--
	}
	if truncate {
		mantissa |= 1
	}
	for mantissa>>(1+mantissBits+2) != 0 {
		mantissa = mantissa>>1 | mantissa&1
		exponent++
	}

	for mantissa > 1 && exponent < minExp-2 {
		mantissa = mantissa>>1 | mantissa&1
		exponent++
	}

	round := mantissa & 3
	mantissa >>= 2
	round |= mantissa & 1 // round to even (round up if mantissa is odd)
	exponent += 2
	if round == 3 {
		mantissa++
		if mantissa == 1<<(1+mantissBits) {
			mantissa >>= 1
			exponent++
		}
	}

	if mantissa>>mantissBits == 0 { // Denormal or zero.
		exponent = bias
	}
	var err error = nil
	if exponent > maxExp { // infinity and range error
		mantissa = 1 << mantissBits
		exponent = maxExp + 1
		err = errors.New("Parse float")
	}
	bits := mantissa & (1<<mantissBits - 1)
	bits |= uint64((exponent-bias)&(1<<exponentBits-1)) << mantissBits

	return float64(Float32FromBits(uint32(bits))), err
}

func Float32FromBits(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }
