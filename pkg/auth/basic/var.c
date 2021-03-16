
#include <stdio.h>

void basic_var();
int addtwonum();


int addtwonum1();
int x1=3;
int y1=4;


int main() {

    addtwonum();
    basic_var();
    addtwonum1();

    return 0;
}


int x,y;
int addtwonum() {
    //声明x,y为外部变量
    extern int x;
    extern int y;

    x = 1;
    y = 2;
    printf("addtwnum() x:%d y:%d \n",x,y);
    return x+y;
}

void basic_var() {
    char c1;
    float f1;
    double d1;

    //extern
    int i1;   //声明,建立存储空间
    extern int i2; //声明,不建立存储空间

    printf("basic_var() x:%d y:%d \n",x,y);
}