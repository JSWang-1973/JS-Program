#include <algorithm>
#include <vector>
#include <iostream>
#include <string>
using namespace std;
int main(void){
int a[3][3] = {
	{1, 2, 3},
	{4, 5, 6},
	{7, 8, 9}
	};
int b[3][3];
	  // a[0][0]
  printf("Address of a in main = %p\n", a);
  printf("Address of *a in main = %p\n", *a);
  printf("Value of **a in main = %d\n", **(a));
  
  cout <<"Address of a in main" << a << endl;
  cout <<"Address of *a in main" <<*(a) <<endl;
  cout <<"Value of a[0][0]" << *(*(a))<<endl;
  
   // a[1][1]
  printf("Address of a + 1 in main = %p\n", a + 1);
  printf("Address of *(a + 1) + 1 in main = %p\n", *(a + 1) + 1);
  cout <<"Address of *(a + 1) + 1 in main" << *(a + 1) + 1 << endl;
  cout <<"Value of a[1][1] in main = " << *(*(a + 1) + 1)<< endl;
   // a[2][2]
  cout <<"Value of a[2][2] in main = " << *(*(a + 2) + 2)<< endl;

	for(int row=0 ; row < 3 ; row++){
		for(int clumn=0 ; clumn < 3 ; clumn++ ){
			
			int temp = *(*(a+row)+clumn);
			*(*(b+row)+clumn)  =temp;
			cout << *(*(a+row)+clumn);
			
		}
		cout << endl; 
	}
		cout << endl; 
	for(int i=0 ; i < 3 ; i++){
		for(int j=0 ; j < 3 ; j++ ){
			cout << *(*(b+i)+j);

		}
		cout << endl; 
	}
    int **c = NULL;
    int row = 3;//用於表示行數
    int col = 3;//用於表示列數
    c = new int*[row];//開闢一塊記憶體來存放每一行的地址
    for (int i = 0; i < col; i++)//分別為每一行開闢記憶體
 	 c[i] = new int[col];
 
	for(int row=0 ;row < 3; row++)
	{
		for(int clumn =0 ; clumn< 3 ; clumn++)
		{
			*(*(c+row)+clumn) = *(*(b+row)+clumn);
		}
	}
	cout << "c" <<endl;
	for(int i=0 ; i < 3 ; i++){
		for(int j=0 ; j < 3 ; j++ ){
			cout << *(*(c+i)+j);
		}
		cout << endl; 
	} 
 
}
