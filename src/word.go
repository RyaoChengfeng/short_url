package main

var elements = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// base10ToBase62 十进制转换成62进制
func base10ToBase62(n int64) string {
	var str string
	for n != 0 {
		str += string(elements[n % 62])
		n /= 62
	}
	return str
}
