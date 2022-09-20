/*
 * You are given two Matries and to product it . 
 
 * 
 * Example 1:
 * Input: A = [1,1,1
 *             1,1,1
 *             1,1,1]
 * 
 * Input: B = [2,2,2
 *             2,2,2
 *             2,2,2]

 * Output: A*B = C =[6,6,6
 *                   6,6,6
 *                   6,6,6]

 */
 
#include <vector>
#include <iostream>
#include <string>
using namespace std;
void Mul(int** A, int**B , int** C){
    
    for(int row=0 ; row < 3 ; row++){
        //int row1 = *(*(A+row)+0);
        //int row2 = *(*(A+row)+1);
        //int row3 = *(*(A+row)+2);
    for(int clumn=0 ; clumn < 3 ; clumn++ ){
            //int c1 = (*(*(B)+clumn));
            //int c2 = (*(*(B+1)+clumn));
            //int c3 = (*(*(B+2)+clumn));
            int temp = *(*(A+row)+0) * (*(*(B+0)+clumn)) +
                       *(*(A+row)+1) * (*(*(B+1)+clumn)) +
                       *(*(A+row)+2) * (*(*(B+2)+clumn)) ;
            //cout <<"row1: " <<row1 <<"\tclumn1 :" <<c1 <<endl;
            //cout <<"row2: " <<row2 <<"\tclumn1 :" <<c2 <<endl;
            //cout <<"row3: " <<row3 <<"\tclumn2 :" <<c3 <<endl;
            *(*(C+row)+clumn)  =temp;
        }
    }
}
int** Allocate2D_Array(int row , int clumn)
{
    int** Array_2D = new int*[row];
    for(int i = 0 ;i < row ; i++){
        Array_2D[i] = new int[clumn];
    }
    return Array_2D;
}

void Free2D_Array(int** array, int row , int clumn)
{
    for (int i = 0 ; i < row ; i++)
        delete array[i];
    delete array;
}

int main(void){
    int a[3][3] = {
    {1, 1, 1},
    {1, 1, 1},
    {1, 1, 1}
    };
    int b[3][3] = {
        {2,2,2},
        {2,2,2},
        {2,2,2},
    };
    int** pt=NULL;
    pt = Allocate2D_Array(3 , 3);
    for(int i =0 ; i < 3; i++)
        for(int j =0 ; j < 3 ; j++)
        *(*(pt+i)+j)= a[i][j];

    for(int row=0 ; row < 3 ; row++){
    for(int clumn=0 ; clumn < 3 ; clumn++ ){
            int temp = *(*(pt+row)+clumn);
            cout << temp<<" ";
        }
        cout << endl;
    }
    Free2D_Array(pt, 3,3);
    
    /*有一個陣列可以放3個地址 (int**)
    就是A變數可以放3個 一維陣列(int*) 啦*/
    int *A[3];
    for(int i=0 ; i < 3 ; i++)
        A[i] = &a[i][0];
    /*有一個陣列可以放3個地址 (int**)
    就是B變數可以放3個 一維陣列(int*) 啦*/
    int *B[3];
    for(int i=0 ; i < 3 ; i++)
        B[i] = &b[i][0];
        
    int **C=NULL;
    int row =3, clumn =3;
    C = new int*[3];

    for(int i=0 ; i < row ; i++)
        C[i] = new int[3];
        
    Mul(A, B ,C);
    
    for(int row=0 ; row < 3 ; row++){
    for(int clumn=0 ; clumn < 3 ; clumn++ ){
            int temp = *(*(C+row)+clumn);
            cout << temp<<" ";  
        }
        cout << endl;
    }

    //To delete C 
    for(int i =0 ; i < row ; i++)
        delete C[i];
    delete C ; 
    return 0;
}
