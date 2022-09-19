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

	for(int i=0 ; i < 3 ; i++){
		for(int j=0 ; j < 3 ; j++ ){
			cout << *(*(a+i)+j);
		}
		cout << endl; 
	}
}
