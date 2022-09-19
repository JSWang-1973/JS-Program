/*
 * You are given two Matries and to add it . 
 
 * 
 * Example 1:
 * Input: A = [1,1,1
 *             1,1,1
 *             1,1,1]
 * 
 * Input: B = [2,2,2
 *             2,2,2
 *             2,2,2]

 * Output: C = [3,3,3
 *              3,3,3
 *              3,3,3]

 */
 
#include <vector>
#include <iostream>
#include <string>
using namespace std;
void Add(int** A, int**B , int** C){
	for(int row=0 ; row < 3 ; row++){
	for(int clumn=0 ; clumn < 3 ; clumn++ ){
			int temp = *(*(A+row)+clumn) + *(*(B+row)+clumn) ;
			*(*(C+row)+clumn)  =temp;
		}
	}
}

int main(void){
	int a[3][3] = {
	{1, 1, 1},
	{1, 1, 1},
	{1, 1, 1}
	};
	int b[3][3] = {
		{1,1,1},
		{1,1,1},
		{1,1,1},
	};

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
		
	Add(A, B ,C);
	
	for(int row=0 ; row < 3 ; row++){
	for(int clumn=0 ; clumn < 3 ; clumn++ ){
			int temp = *(*(C+row)+clumn);
			cout << temp;
		}
		cout << endl;
	}

	//To delete C 
	for(int i =0 ; i < row ; i++)
		delete C[i];
	return 0;
}
