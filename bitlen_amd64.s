// +build !gccgo

// func bitlen(x Word) (n int)
TEXT ·bitlen(SB),4,$0
	BSRQ x+0(FP), AX
	JZ z1
	ADDQ $1, AX
	MOVQ AX, n+8(FP)
	RET

z1:	MOVQ $0, n+8(FP)
	RET
