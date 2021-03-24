#include <stdio.h>
#include <string.h>

#define NEWLINE '\n'
void arr1();
void arr2();

int main(int argc, char *argv[]) {
    if(argc <= 1)  {
        printf("need func args! %c",NEWLINE);
        return 0;
    }
    
    if (strncmp(argv[1],"arr1",strlen(argv[1])) == 0) {
        arr1();
        return 0;
    }

    if (strncmp(argv[1],"arr2",strlen(argv[1])) == 0) {
        arr2();
        return 0;
    }
    
    return 0;
}

void arr1() {
    double balance[10];
    double balance1[5] = {1000.0, 2.0, 3.4, 7.0, 50.0};
    double balance2[] = {1000.0, 2.0, 3.4, 7.0, 50.0};

    int n[10];
    for (int i = 0; i < 10; i++)
    {
        n[i] = i + 100;
    }

    for (int i = 0; i < 10; i++)
    {
        printf("ele[%d] = %d %c",i,n[i],NEWLINE);
    }
    
    
    return ;
}

void arr2() {
    /* 带有 5 个元素的整型数组 */
   double balance[5] = {1000.0, 2.0, 3.4, 17.0, 50.0};
   double *p;
   int i;
 
   p = balance;
 
   /* 输出数组中每个元素的值 */
   printf( "使用指针的数组值\n");
   for ( i = 0; i < 5; i++ )
   {
       printf("*(p + %d) : %f\n",  i, *(p + i) );
   }
 
   printf( "使用 balance 作为地址的数组值\n");
   for ( i = 0; i < 5; i++ )
   {
       printf("*(balance + %d) : %f\n",  i, *(balance + i) );
   }

    return;
}