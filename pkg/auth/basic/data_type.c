#include <stdio.h>
#include <float.h>

void pr();

int main() {
    pr();

    return 0;
}

/**
 * 1 byte = 1字节 = 8bit = 8位
 * +-----+----------------+----------++-----+----------------+----------+
 * |type | range          | size     | 
 * 
 * |char | -128~127/0~255 | 1 byte   | 
 * 
 * |unsigned char | 0~255 | 1 byte   |
 * |signed char | -128~127 | 1 byte  |
 * 
 * |int  | -32768~32767 / -2147483648~2147483647 | 2/4 byte  |
 * |unsigned int  | 0~65535 / 0 到 4294967295 | 2/4 byte  |
 * 
 * |short  | -32768~32767 | 2 byte  | 
 * |unsigned short  | 0~65535 / 0 到 4294967295 | 2 byte  |
 * 
 * |long  | -2147483648 ~ 2147483647 | 4 byte  | 
 * |unsigned long  | 0 ~ 4294967295 | 4 byte  |b
 * +-----+----------------+----------++-----+----------------+----------+
 *   
 * 
 * %d 十进制有符号整数(正数不带符号)
 * %o 八进制无符号整数(不输出前缀0)
 * %x %X 十六禁止无符号整数(无前缀Ox)
 * %u 十进制无符号整数
 * %f 小数形式输出单,双精度实数
 * %e %E 指数形式输出单,双精度实数
 * %c 单个字符
 * %s 字符串
 * %p 指针地址
 * %lu 32位无符号整数
 * %llu 64位无符号整数
 */
void pr() {
    printf("float 最大字节数 %lu \n",sizeof(float));
    printf("float 最小值 %E \n",FLT_MIN);
    printf("float 最大值 %E \n",FLT_MAX);
    printf("精度值 %d \n",FLT_DIG);

    int ch;
    for (ch = 97; ch <= 97+25; ch++) {
        printf(" %c ",ch);
    }
    printf("\n");

    for (ch = 65; ch <= 65+25; ch++) {
        printf(" %c ",ch);
    }
    printf("\n");

    /** flags(跟%一起使用)
     * %#x : 带0x前缀输出
     * %+d : 正数也带符号输出
     * %-x : 字段左对齐
     */
    char i = 1;
    printf("i: %#x \n",i);
}