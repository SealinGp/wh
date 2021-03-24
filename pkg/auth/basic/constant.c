#include <stdio.h>

#include <string.h>

/**
 * 宏定义,常量
 */
#define NEWLINE '\n'

void constantFunc();
void saveClass();
void yunsuanfu();
void loop();

static int count = 10;

int main(int argc, char *argv[]) {
    if(argc <= 1)  {
        printf("need func args! %c",NEWLINE);
        return 0;
    }
    
    if (strncmp(argv[1],"constantFunc",strlen(argv[1])) == 0) {
        constantFunc();
        return 0;
    }

    if (strncmp(argv[1],"saveClass",strlen(argv[1])) == 0) {
        while (count--)
        {
            saveClass();
        }
        return 0;
    }

    if (strncmp(argv[1],"yunsuanfu",strlen(argv[1])) == 0) {
        yunsuanfu();
        return 0;
    }
    
    printf("invalid args[1]: %s %c",argv[1],NEWLINE);
    return 0;
}


//静态变量
#define LENGTH 10
#define WIDTH 5
const int I1 = 4;
void constantFunc() {
    int area;
    area = LENGTH*WIDTH;

    printf("area:%d %c",area,NEWLINE);
    printf("const I1:%d %c",I1,NEWLINE);
    return;
}

//存储类
//extern 别的文件 引入此函数,并初始化全局变量int I2 = 1; 则此文件中I2 = 1
extern int I2;
void saveClass() {
    //auto: 所有局部变量默认的存储类       
    auto int mount; //  = int mount;

    //register: 寄存器中,最大尺寸=寄存器大小 不能用&运算符,用于快速访问的变量,如计数器
    register int miles;

    //static 每次调用该函数,局部static变量不会重置
    static int thingy = 5;
    thingy++;
    printf("thingy:%d, count:%d %c",thingy,count,NEWLINE);

    //extern: 提供一个全局变量的引用,也就是当前文件的输入全局变量    
    printf("I2:%d %c",count,NEWLINE);

    return;
}


//运算符
void yunsuanfu() {
    int a = 21;
    int b = 10;
    int c;

    c = a + b;
    printf("21 + 10 = %d %c",c,NEWLINE);

    c = a - b;
    printf("21 - 10 = %d %c",c,NEWLINE);

    c = a * b;
    printf("21 * 10 = %d %c",c,NEWLINE);

    c = a / b;
    printf("21 / 10 = %d %c",c,NEWLINE);

    c = a % b;
    printf("21 %% 10 = %d %c",c,NEWLINE);

    c = a++;
    printf("21++ = %d %c",c,NEWLINE);

    c = a--;
    printf("21-- = %d %c",c,NEWLINE);

    return;    
}

void loop() {
    for (; ;)
    {
        /* code */
    }
    

    return;
}