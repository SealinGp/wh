#include <stdio.h>


/**
 * 宏定义,常量
 */

#define LENGTH 10
#define WIDTH 5
#define NEWLINE '\n'


int main() {
    int area;
    
    area = LENGTH*WIDTH;

    printf("area:%d %c",area,NEWLINE);

    return 0;
}