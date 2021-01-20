// +build !gccgo

// func bitlen(x Word) (n int)
TEXT ·bitlen(SB),4,$0
	BSRL x+0(FP), AX
	JZ z1
	INCL AX
	MOVL AX, n+4(FP)
	RET

z1:	MOVL $0, n+4(FP)
	RET
