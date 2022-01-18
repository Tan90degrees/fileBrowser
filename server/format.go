package server

func formatHead(in string, outLen int) string {
	inLen := len(in)
	if inLen < outLen {
		tmp := make([]byte, outLen-inLen)
		for i := range tmp {
			tmp[i] = ' '
		}
		return string(append(tmp, []byte(in)...))
	} else {
		return string(append([]byte(in[0:outLen-3]), []byte("...")...))
	}
}

func formatTail(in string, outLen int) string {
	inLen := len(in)
	if inLen < outLen {
		tmp := make([]byte, outLen-inLen)
		for i := range tmp {
			tmp[i] = ' '
		}
		return string(append([]byte(in), tmp...))
	} else {
		return string(append([]byte(in[0:outLen-3]), []byte("...")...))
	}
}

func caclEmpty(in string, outLen int) int {
	inLen := len(in)
	if inLen < outLen {
		return outLen - inLen + 28
	} else {
		return 28
	}
}
