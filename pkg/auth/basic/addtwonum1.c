
#include <stdio.h>

//引用另一个文件test.c中的变量, 外部输入变量
extern int x1;
extern int y1;

int addtwonum1() {
    printf("addtwnum1() x1:%d y1:%d \n",x1,y1);
    return x1+y1;
}