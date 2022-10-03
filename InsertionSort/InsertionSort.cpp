#include <algorithm>
#include <vector>
#include <iostream>
#include <string>
using namespace std;
/*
void insertionSort(int array[], int size){
	int key , j ;
	for(int index =1 ; index < size ; index++){
		key = array[index];
		j = index -1 ;
		while(key < array[j] && j >= 0){
			array[j+1] = array[j];
			j--;
		}
		array[j+1] =key;
	}

}
*/
void insertionSort(int array[], int size){
	int key , i , j;
	for(i = 1 ; i < size ; i++){
		key = array[i];
		for(j=i-1;(key < array[j] && j >=0) ; j-- ){
			array[j+1] =array[j];
		}
		array[j+1] =key;
	}
}

int main(void){
	int size = 10;
	int array[size]={5, 2 , 6 , 0 , 7, 1 , 11, 8, 16, 15};
	for(int i=0 ; i < size ; i++)
		cout << "the data : " << array[i] << endl ;
	cout <<"+++++++++++++++++++"<<endl;
	cout <<"after insertion sort" << endl; 
	insertionSort(array, size);
	for(int i=0 ; i < size ; i++)
		cout << "the data : " << array[i] << endl ;
}
