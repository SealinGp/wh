package self_report

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

func StrEnc(data, firstKey, secondKey, thirdKey string) string {
	leng := len(data)
	encData := ""

	if leng <= 0 {
		return encData
	}

	firstKeyBt := GetKeyBytes(firstKey)
	firstLength := len(firstKeyBt)
	secondKeyBt := GetKeyBytes(secondKey)
	secondLength := len(secondKeyBt)
	thirdKeyBt := GetKeyBytes(thirdKey)
	thirdLength := len(thirdKeyBt)

	if leng < 4 {
		var bt = StrToBt(data)
		var encByte []int
		if firstKey != "" && secondKey != "" && thirdKey != "" {
			var tempBt []int
			var x, y, z int
			tempBt = bt
			for x = 0; x < firstLength; x++ {
				tempBt = Enc(tempBt, firstKeyBt[x])
			}
			for y = 0; y < secondLength; y++ {
				tempBt = Enc(tempBt, secondKeyBt[y])
			}
			for z = 0; z < thirdLength; z++ {
				tempBt = Enc(tempBt, thirdKeyBt[z])
			}
			encByte = tempBt
		} else if firstKey != "" && secondKey != "" {
			var tempBt []int
			var x, y int
			tempBt = bt
			for x = 0; x < firstLength; x++ {
				tempBt = Enc(tempBt, firstKeyBt[x])
			}
			for y = 0; y < secondLength; y++ {
				tempBt = Enc(tempBt, secondKeyBt[y])
			}
			encByte = tempBt
		} else if firstKey != "" {
			var tempBt []int
			var x = 0
			tempBt = bt
			for x = 0; x < firstLength; x++ {
				tempBt = Enc(tempBt, firstKeyBt[x])
			}
			encByte = tempBt
		}

		encData = Bt64ToHex(encByte)
	} else {
		iterator := leng / 4
		remainder := leng % 4
		var i = 0
		for i = 0; i < iterator; i++ {
			var tempData = data[i*4+0 : i*4+4]
			var tempByte = StrToBt(tempData)
			var encByte []int
			if firstKey != "" && secondKey != "" && thirdKey != "" {
				var tempBt []int
				var x, y, z int
				tempBt = tempByte
				for x = 0; x < firstLength; x++ {
					tempBt = Enc(tempBt, firstKeyBt[x])
				}
				for y = 0; y < secondLength; y++ {
					tempBt = Enc(tempBt, secondKeyBt[y])
				}
				for z = 0; z < thirdLength; z++ {
					tempBt = Enc(tempBt, thirdKeyBt[z])
				}

				encByte = tempBt
			} else if firstKey != "" && secondKey != "" {
				var tempBt []int
				var x, y int
				tempBt = tempByte
				for x = 0; x < firstLength; x++ {
					tempBt = Enc(tempBt, firstKeyBt[x])
				}
				for y = 0; y < secondLength; y++ {
					tempBt = Enc(tempBt, secondKeyBt[y])
				}
				encByte = tempBt
			} else if firstKey != "" {
				var tempBt []int
				var x int
				tempBt = tempByte
				for x = 0; x < firstLength; x++ {
					tempBt = Enc(tempBt, firstKeyBt[x])
				}
				encByte = tempBt
			}
			encData += Bt64ToHex(encByte)
		}
		if remainder > 0 {
			var remainderData = data[iterator*4+0 : leng]
			var tempByte = StrToBt(remainderData)
			var encByte []int

			if firstKey != "" && secondKey != "" && thirdKey != "" {
				var tempBt []int
				var x, y, z int
				tempBt = tempByte
				for x = 0; x < firstLength; x++ {
					tempBt = Enc(tempBt, firstKeyBt[x])
				}
				for y = 0; y < secondLength; y++ {
					tempBt = Enc(tempBt, secondKeyBt[y])
				}
				for z = 0; z < thirdLength; z++ {
					tempBt = Enc(tempBt, thirdKeyBt[z])
				}
				encByte = tempBt
			} else if firstKey != "" && secondKey != "" {
				var tempBt []int
				var x, y int
				tempBt = tempByte
				for x = 0; x < firstLength; x++ {
					tempBt = Enc(tempBt, firstKeyBt[x])
				}
				for y = 0; y < secondLength; y++ {
					tempBt = Enc(tempBt, secondKeyBt[y])
				}
				encByte = tempBt
			} else if firstKey != "" {
				var tempBt []int
				var x int
				tempBt = tempByte
				for x = 0; x < firstLength; x++ {
					tempBt = Enc(tempBt, firstKeyBt[x])
				}
				encByte = tempBt
			}

			encData += Bt64ToHex(encByte)
		}
	}

	return encData
}

func Bt64ToHex(byteData []int) string {
	var hex = ""
	for i := 0; i < 16; i++ {
		var bt = ""

		for j := 0; j < 4; j++ {
			bt = fmt.Sprintf("%s%d", bt, byteData[i*4+j])
		}

		hex += Bt4ToHex(bt)
	}
	return hex
}

func Bt4ToHex(binary string) string {
	var hex string
	switch binary {
	case "0000":
		hex = "0"
		break
	case "0001":
		hex = "1"
		break
	case "0010":
		hex = "2"
		break
	case "0011":
		hex = "3"
		break
	case "0100":
		hex = "4"
		break
	case "0101":
		hex = "5"
		break
	case "0110":
		hex = "6"
		break
	case "0111":
		hex = "7"
		break
	case "1000":
		hex = "8"
		break
	case "1001":
		hex = "9"
		break
	case "1010":
		hex = "A"
		break
	case "1011":
		hex = "B"
		break
	case "1100":
		hex = "C"
		break
	case "1101":
		hex = "D"
		break
	case "1110":
		hex = "E"
		break
	case "1111":
		hex = "F"
		break
	}
	return hex
}

func Enc(dataByte []int, keyByte []int) []int {
	var keys = GenerateKeys(keyByte)
	var ipByte = InitPermute(dataByte)
	var ipLeft = make([]int, 32)
	var ipRight = make([]int, 32)
	var tempLeft = make([]int, 32)

	i, j, k, m, n := 0, 0, 0, 0, 0

	for k = 0; k < 32; k++ {
		ipLeft[k] = ipByte[k]
		ipRight[k] = ipByte[32+k]
	}

	for i = 0; i < 16; i++ {
		for j = 0; j < 32; j++ {
			tempLeft[j] = ipLeft[j]
			ipLeft[j] = ipRight[j]
		}

		var key = make([]int, 48)
		for m = 0; m < 48; m++ {
			key[m] = keys[i][m]
		}
		var tempRight = Xor(PPermute(SBoxPermute(Xor(ExpandPermute(ipRight), key))), tempLeft)

		for n = 0; n < 32; n++ {
			ipRight[n] = tempRight[n]
		}
	}

	var finalData = make([]int, 64)
	for i = 0; i < 32; i++ {
		finalData[i] = ipRight[i]
		finalData[32+i] = ipLeft[i]
	}
	return FinallyPermute(finalData)
}

func Dec(dataByte []int, keyByte []int) []int {
	keys := GenerateKeys(keyByte)
	ipByte := InitPermute(dataByte)

	var ipLeft = make([]int, 32)
	var ipRight = make([]int, 32)
	var tempLeft = make([]int, 32)

	i := 0
	j := 0
	k := 0
	m := 0
	n := 0

	for k = 0; k < 32; k++ {
		ipLeft[k] = ipByte[k]
		ipRight[k] = ipByte[32+k]
	}
	for i = 15; i >= 0; i-- {
		for j = 0; j < 32; j++ {
			tempLeft[j] = ipLeft[j]
			ipLeft[j] = ipRight[j]
		}

		key := make([]int, 48)
		for m = 0; m < 48; m++ {
			key[m] = keys[i][m]
		}

		var tempRight = Xor(PPermute(SBoxPermute(Xor(ExpandPermute(ipRight), key))), tempLeft)
		for n = 0; n < 32; n++ {
			ipRight[n] = tempRight[n]
		}
	}

	var finalData = make([]int, 64)
	for i = 0; i < 32; i++ {
		finalData[i] = ipRight[i]
		finalData[32+i] = ipLeft[i]
	}

	return FinallyPermute(finalData)
}

func FinallyPermute(endByte []int) []int {
	var fpByte = make([]int, 64)
	fpByte[0] = endByte[39]
	fpByte[1] = endByte[7]
	fpByte[2] = endByte[47]
	fpByte[3] = endByte[15]
	fpByte[4] = endByte[55]
	fpByte[5] = endByte[23]
	fpByte[6] = endByte[63]
	fpByte[7] = endByte[31]
	fpByte[8] = endByte[38]
	fpByte[9] = endByte[6]
	fpByte[10] = endByte[46]
	fpByte[11] = endByte[14]
	fpByte[12] = endByte[54]
	fpByte[13] = endByte[22]
	fpByte[14] = endByte[62]
	fpByte[15] = endByte[30]
	fpByte[16] = endByte[37]
	fpByte[17] = endByte[5]
	fpByte[18] = endByte[45]
	fpByte[19] = endByte[13]
	fpByte[20] = endByte[53]
	fpByte[21] = endByte[21]
	fpByte[22] = endByte[61]
	fpByte[23] = endByte[29]
	fpByte[24] = endByte[36]
	fpByte[25] = endByte[4]
	fpByte[26] = endByte[44]
	fpByte[27] = endByte[12]
	fpByte[28] = endByte[52]
	fpByte[29] = endByte[20]
	fpByte[30] = endByte[60]
	fpByte[31] = endByte[28]
	fpByte[32] = endByte[35]
	fpByte[33] = endByte[3]
	fpByte[34] = endByte[43]
	fpByte[35] = endByte[11]
	fpByte[36] = endByte[51]
	fpByte[37] = endByte[19]
	fpByte[38] = endByte[59]
	fpByte[39] = endByte[27]
	fpByte[40] = endByte[34]
	fpByte[41] = endByte[2]
	fpByte[42] = endByte[42]
	fpByte[43] = endByte[10]
	fpByte[44] = endByte[50]
	fpByte[45] = endByte[18]
	fpByte[46] = endByte[58]
	fpByte[47] = endByte[26]
	fpByte[48] = endByte[33]
	fpByte[49] = endByte[1]
	fpByte[50] = endByte[41]
	fpByte[51] = endByte[9]
	fpByte[52] = endByte[49]
	fpByte[53] = endByte[17]
	fpByte[54] = endByte[57]
	fpByte[55] = endByte[25]
	fpByte[56] = endByte[32]
	fpByte[57] = endByte[0]
	fpByte[58] = endByte[40]
	fpByte[59] = endByte[8]
	fpByte[60] = endByte[48]
	fpByte[61] = endByte[16]
	fpByte[62] = endByte[56]
	fpByte[63] = endByte[24]
	return fpByte
}

func PPermute(sBoxByte []int) []int {
	var pBoxPermute = make([]int, 32)
	pBoxPermute[0] = sBoxByte[15]
	pBoxPermute[1] = sBoxByte[6]
	pBoxPermute[2] = sBoxByte[19]
	pBoxPermute[3] = sBoxByte[20]
	pBoxPermute[4] = sBoxByte[28]
	pBoxPermute[5] = sBoxByte[11]
	pBoxPermute[6] = sBoxByte[27]
	pBoxPermute[7] = sBoxByte[16]
	pBoxPermute[8] = sBoxByte[0]
	pBoxPermute[9] = sBoxByte[14]
	pBoxPermute[10] = sBoxByte[22]
	pBoxPermute[11] = sBoxByte[25]
	pBoxPermute[12] = sBoxByte[4]
	pBoxPermute[13] = sBoxByte[17]
	pBoxPermute[14] = sBoxByte[30]
	pBoxPermute[15] = sBoxByte[9]
	pBoxPermute[16] = sBoxByte[1]
	pBoxPermute[17] = sBoxByte[7]
	pBoxPermute[18] = sBoxByte[23]
	pBoxPermute[19] = sBoxByte[13]
	pBoxPermute[20] = sBoxByte[31]
	pBoxPermute[21] = sBoxByte[26]
	pBoxPermute[22] = sBoxByte[2]
	pBoxPermute[23] = sBoxByte[8]
	pBoxPermute[24] = sBoxByte[18]
	pBoxPermute[25] = sBoxByte[12]
	pBoxPermute[26] = sBoxByte[29]
	pBoxPermute[27] = sBoxByte[5]
	pBoxPermute[28] = sBoxByte[21]
	pBoxPermute[29] = sBoxByte[10]
	pBoxPermute[30] = sBoxByte[3]
	pBoxPermute[31] = sBoxByte[24]
	return pBoxPermute

}

func SBoxPermute(expandByte []int) []int {
	var sBoxByte = make([]int, 32)
	var binary = ""
	var s1 = [][]int{
		{14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7},
		{0, 15, 7, 4, 14, 2, 13, 1, 10, 6, 12, 11, 9, 5, 3, 8},
		{4, 1, 14, 8, 13, 6, 2, 11, 15, 12, 9, 7, 3, 10, 5, 0},
		{15, 12, 8, 2, 4, 9, 1, 7, 5, 11, 3, 14, 10, 0, 6, 13},
	}

	var s2 = [][]int{
		{15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10},
		{3, 13, 4, 7, 15, 2, 8, 14, 12, 0, 1, 10, 6, 9, 11, 5},
		{0, 14, 7, 11, 10, 4, 13, 1, 5, 8, 12, 6, 9, 3, 2, 15},
		{13, 8, 10, 1, 3, 15, 4, 2, 11, 6, 7, 12, 0, 5, 14, 9},
	}

	var s3 = [][]int{
		{10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8},
		{13, 7, 0, 9, 3, 4, 6, 10, 2, 8, 5, 14, 12, 11, 15, 1},
		{13, 6, 4, 9, 8, 15, 3, 0, 11, 1, 2, 12, 5, 10, 14, 7},
		{1, 10, 13, 0, 6, 9, 8, 7, 4, 15, 14, 3, 11, 5, 2, 12},
	}

	var s4 = [][]int{
		{7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15},
		{13, 8, 11, 5, 6, 15, 0, 3, 4, 7, 2, 12, 1, 10, 14, 9},
		{10, 6, 9, 0, 12, 11, 7, 13, 15, 1, 3, 14, 5, 2, 8, 4},
		{3, 15, 0, 6, 10, 1, 13, 8, 9, 4, 5, 11, 12, 7, 2, 14},
	}

	/* Table - s5 */
	var s5 = [][]int{
		{2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9},
		{14, 11, 2, 12, 4, 7, 13, 1, 5, 0, 15, 10, 3, 9, 8, 6},
		{4, 2, 1, 11, 10, 13, 7, 8, 15, 9, 12, 5, 6, 3, 0, 14},
		{11, 8, 12, 7, 1, 14, 2, 13, 6, 15, 0, 9, 10, 4, 5, 3},
	}

	/* Table - s6 */
	var s6 = [][]int{
		{12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11},
		{10, 15, 4, 2, 7, 12, 9, 5, 6, 1, 13, 14, 0, 11, 3, 8},
		{9, 14, 15, 5, 2, 8, 12, 3, 7, 0, 4, 10, 1, 13, 11, 6},
		{4, 3, 2, 12, 9, 5, 15, 10, 11, 14, 1, 7, 6, 0, 8, 13},
	}

	var s7 = [][]int{
		{4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1},
		{13, 0, 11, 7, 4, 9, 1, 10, 14, 3, 5, 12, 2, 15, 8, 6},
		{1, 4, 11, 13, 12, 3, 7, 14, 10, 15, 6, 8, 0, 5, 9, 2},
		{6, 11, 13, 8, 1, 4, 10, 7, 9, 5, 0, 15, 14, 2, 3, 12},
	}

	var s8 = [][]int{
		{13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7},
		{1, 15, 13, 8, 10, 3, 7, 4, 12, 5, 6, 11, 0, 14, 9, 2},
		{7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 3, 5, 8},
		{2, 1, 14, 7, 4, 10, 8, 13, 15, 12, 9, 0, 3, 5, 6, 11},
	}

	for m := 0; m < 8; m++ {
		i, j := 0, 0

		i = expandByte[m*6+0]*2 + expandByte[m*6+5]
		j = expandByte[m*6+1]*2*2*2 +
			expandByte[m*6+2]*2*2 +
			expandByte[m*6+3]*2 +
			expandByte[m*6+4]

		switch m {
		case 0:
			binary = GetBoxBinary(s1[i][j])
			break
		case 1:
			binary = GetBoxBinary(s2[i][j])
			break
		case 2:
			binary = GetBoxBinary(s3[i][j])
			break
		case 3:
			binary = GetBoxBinary(s4[i][j])
			break
		case 4:
			binary = GetBoxBinary(s5[i][j])
			break
		case 5:
			binary = GetBoxBinary(s6[i][j])
			break
		case 6:
			binary = GetBoxBinary(s7[i][j])
			break
		case 7:
			binary = GetBoxBinary(s8[i][j])
			break
		}

		b1, _ := strconv.ParseInt(binary[0:1], 10, 64)
		sBoxByte[m*4+0] = int(b1)

		b1, _ = strconv.ParseInt(binary[1:2], 10, 64)
		sBoxByte[m*4+1] = int(b1)

		b1, _ = strconv.ParseInt(binary[2:3], 10, 64)
		sBoxByte[m*4+2] = int(b1)

		b1, _ = strconv.ParseInt(binary[3:4], 10, 64)
		sBoxByte[m*4+3] = int(b1)
	}

	return sBoxByte
}

func GetBoxBinary(i int) string {
	var binary = ""
	switch i {
	case 0:
		binary = "0000"
		break
	case 1:
		binary = "0001"
		break
	case 2:
		binary = "0010"
		break
	case 3:
		binary = "0011"
		break
	case 4:
		binary = "0100"
		break
	case 5:
		binary = "0101"
		break
	case 6:
		binary = "0110"
		break
	case 7:
		binary = "0111"
		break
	case 8:
		binary = "1000"
		break
	case 9:
		binary = "1001"
		break
	case 10:
		binary = "1010"
		break
	case 11:
		binary = "1011"
		break
	case 12:
		binary = "1100"
		break
	case 13:
		binary = "1101"
		break
	case 14:
		binary = "1110"
		break
	case 15:
		binary = "1111"
		break
	}
	return binary
}

func ExpandPermute(rightData []int) []int {
	var epByte = make([]int, 48)

	for i := 0; i < 8; i++ {
		if i == 0 {
			epByte[i*6+0] = rightData[31]
		} else {
			epByte[i*6+0] = rightData[i*4-1]
		}
		epByte[i*6+1] = rightData[i*4+0]
		epByte[i*6+2] = rightData[i*4+1]
		epByte[i*6+3] = rightData[i*4+2]
		epByte[i*6+4] = rightData[i*4+3]
		if i == 7 {
			epByte[i*6+5] = rightData[0]
		} else {
			epByte[i*6+5] = rightData[i*4+4]
		}
	}
	return epByte
}

func Xor(byteOne, byteTwo []int) []int {
	var xorByte = make([]int, len(byteOne))
	for i := 0; i < len(byteOne); i++ {
		xorByte[i] = byteOne[i] ^ byteTwo[i]
	}

	return xorByte
}

func InitPermute(originalData []int) []int {
	var ipByte = make([]int, 64)
	i, m, n := 0, 1, 0
	for i < 4 {

		j, k := 7, 0
		for j >= 0 {
			ipByte[i*8+k] = originalData[j*8+m]
			ipByte[i*8+k+32] = originalData[j*8+n]

			j--
			k++
		}

		i++
		m += 2
		n += 2
	}
	return ipByte
}

func GenerateKeys(keyByte []int) [][]int {
	key := make([]int, 56)
	keys := make([][]int, 16)
	for i := range keys {
		keys[i] = make([]int, 48)
	}

	var loop = []int{1, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1}

	for i := 0; i < 7; i++ {
		j, k := 0, 7

		for j < 8 {
			key[i*8+j] = keyByte[8*k+i]
			j++
			k--
		}
	}

	for i := 0; i < 16; i++ {
		var tempLeft = 0
		var tempRight = 0

		for j := 0; j < loop[i]; j++ {
			tempLeft = key[0]
			tempRight = key[28]
			for k := 0; k < 27; k++ {
				key[k] = key[k+1]
				key[28+k] = key[29+k]
			}
			key[27] = tempLeft
			key[55] = tempRight
		}

		tempKey := make([]int, 48)
		tempKey[0] = key[13]
		tempKey[1] = key[16]
		tempKey[2] = key[10]
		tempKey[3] = key[23]
		tempKey[4] = key[0]
		tempKey[5] = key[4]
		tempKey[6] = key[2]
		tempKey[7] = key[27]
		tempKey[8] = key[14]
		tempKey[9] = key[5]
		tempKey[10] = key[20]
		tempKey[11] = key[9]
		tempKey[12] = key[22]
		tempKey[13] = key[18]
		tempKey[14] = key[11]
		tempKey[15] = key[3]
		tempKey[16] = key[25]
		tempKey[17] = key[7]
		tempKey[18] = key[15]
		tempKey[19] = key[6]
		tempKey[20] = key[26]
		tempKey[21] = key[19]
		tempKey[22] = key[12]
		tempKey[23] = key[1]
		tempKey[24] = key[40]
		tempKey[25] = key[51]
		tempKey[26] = key[30]
		tempKey[27] = key[36]
		tempKey[28] = key[46]
		tempKey[29] = key[54]
		tempKey[30] = key[29]
		tempKey[31] = key[39]
		tempKey[32] = key[50]
		tempKey[33] = key[44]
		tempKey[34] = key[32]
		tempKey[35] = key[47]
		tempKey[36] = key[43]
		tempKey[37] = key[48]
		tempKey[38] = key[38]
		tempKey[39] = key[55]
		tempKey[40] = key[33]
		tempKey[41] = key[52]
		tempKey[42] = key[45]
		tempKey[43] = key[41]
		tempKey[44] = key[49]
		tempKey[45] = key[35]
		tempKey[46] = key[28]
		tempKey[47] = key[31]

		switch i {
		case 0:
			for m := 0; m < 48; m++ {
				keys[0][m] = tempKey[m]
			}
			break
		case 1:
			for m := 0; m < 48; m++ {
				keys[1][m] = tempKey[m]
			}
			break
		case 2:
			for m := 0; m < 48; m++ {
				keys[2][m] = tempKey[m]
			}
			break
		case 3:
			for m := 0; m < 48; m++ {
				keys[3][m] = tempKey[m]
			}
			break
		case 4:
			for m := 0; m < 48; m++ {
				keys[4][m] = tempKey[m]
			}
			break
		case 5:
			for m := 0; m < 48; m++ {
				keys[5][m] = tempKey[m]
			}
			break
		case 6:
			for m := 0; m < 48; m++ {
				keys[6][m] = tempKey[m]
			}
			break
		case 7:
			for m := 0; m < 48; m++ {
				keys[7][m] = tempKey[m]
			}
			break
		case 8:
			for m := 0; m < 48; m++ {
				keys[8][m] = tempKey[m]
			}
			break
		case 9:
			for m := 0; m < 48; m++ {
				keys[9][m] = tempKey[m]
			}
			break
		case 10:
			for m := 0; m < 48; m++ {
				keys[10][m] = tempKey[m]
			}
			break
		case 11:
			for m := 0; m < 48; m++ {
				keys[11][m] = tempKey[m]
			}
			break
		case 12:
			for m := 0; m < 48; m++ {
				keys[12][m] = tempKey[m]
			}
			break
		case 13:
			for m := 0; m < 48; m++ {
				keys[13][m] = tempKey[m]
			}
			break
		case 14:
			for m := 0; m < 48; m++ {
				keys[14][m] = tempKey[m]
			}
			break
		case 15:
			for m := 0; m < 48; m++ {
				keys[15][m] = tempKey[m]
			}
			break
		}

	}
	return keys
}

func HexToBt64(hex string) string {
	binaryStr := ""
	for i := 0; i < 16; i++ {
		binaryStr += HexToBt4(hex[i : i+1])
	}

	return binaryStr
}

func GetKeyBytes(key string) [][]int {
	var keyBytes = make([][]int, len(key)/4+1)

	var leng = len(key)
	var iterator = leng / 4
	var remainder = leng % 4

	for i := 0; i < iterator; i++ {
		keyBytes[i] = StrToBt(string(key[i*4+0 : i*4+4]))
	}
	if remainder > 0 {
		keyBytes[iterator] = StrToBt(string(key[iterator*4+0 : leng]))
	}
	return keyBytes
}

func HexToBt4(hex string) string {
	var binary string
	switch hex {
	case "0":
		binary = "0000"
		break
	case "1":
		binary = "0001"
		break
	case "2":
		binary = "0010"
		break
	case "3":
		binary = "0011"
		break
	case "4":
		binary = "0100"
		break
	case "5":
		binary = "0101"
		break
	case "6":
		binary = "0110"
		break
	case "7":
		binary = "0111"
		break
	case "8":
		binary = "1000"
		break
	case "9":
		binary = "1001"
		break
	case "A":
		binary = "1010"
		break
	case "B":
		binary = "1011"
		break
	case "C":
		binary = "1100"
		break
	case "D":
		binary = "1101"
		break
	case "E":
		binary = "1110"
		break
	case "F":
		binary = "1111"
		break
	}
	return binary
}

func StrToBt(str string) []int {
	leng := len(str)
	bt := make([]int, 64)

	if leng < 4 {
		for i := 0; i < leng; i++ {
			k := str[i]

			for j := 0; j < 16; j++ {

				pow := 1
				for m := 15; m > j; m-- {
					pow *= 2
				}
				bt[16*i+j] = (int(k) / pow) % 2
			}
		}

		for p := leng; p < 4; p++ {
			k := 0
			for q := 0; q < 16; q++ {

				pow := 1
				for m := 15; m > q; m-- {
					pow *= 2
				}
				bt[16*p+q] = (int(k) / pow) % 2
			}
		}

	} else {
		for i := 0; i < 4; i++ {
			k := str[i]

			for j := 0; j < 16; j++ {
				pow := 1
				for m := 15; m > j; m-- {
					pow *= 2
				}
				bt[16*i+j] = (int(k) / pow) % 2
			}
		}
	}

	return bt
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetRandCode() int {
	rand.Seed(time.Now().UnixNano())
	num1 := math.Floor(rand.Float64() * 100)
	num2 := math.Floor(rand.Float64() * 100)

	return int(num1) + int(num2)
}
