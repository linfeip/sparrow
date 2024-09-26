package utils

func AssertLength(n int, err error) int {
	if err != nil {
		panic(err)
	}
	return n
}

func Assert(err error) {
	if err != nil {
		panic(err)
	}
}
