/*
 * Given two sorted arrays nums1 and nums2 of size m and n respectively, return the median of the two sorted arrays.
 * The overall run time complexity should be O(log (m+n)).
 * Example 1:
 * Input: nums1 = [1,3], nums2 = [2]
 * Output: 2.00000
 * Explanation: merged array = [1,2,3] and median is 2.
 * Input: nums1 = [1,2], nums2 = [3,4]
 * Output: 2.50000
 * Explanation: merged array = [1,2,3,4] and median is (2 + 3) / 2 = 2.5.
 * */
 
#include <algorithm>
#include <vector>
#include <iostream>
#include <string>
using namespace std;

double findMedianSortedArrays(vector<int>& nums1, vector<int>& nums2){
    vector<int> mergedV ;
    int size1= nums1.size();
    int size2= nums2.size();
    int M = 0 ;
    int big ,small;
    bool flagEvent = false;
    double Rvalue=0;
    int count = size1 + size2 ;
    M = (size1 + size2)/2;
    if ((size1 + size2) % 2 == 0){
        flagEvent = true;
    }
    vector<int>::iterator it1 = nums1.begin();
    vector<int>::iterator it2 = nums2.begin();
    do{
        int tempA = *(it1);
        int tempB = *(it2);
        if (it1 ==nums1.end()){
            it2++;
            mergedV.push_back(tempB);
        }
        else if (it2 ==nums2.end()){
            it1++;
            mergedV.push_back(tempA);
        }
        else{
            if (tempA < tempB){
                it1++;
                mergedV.push_back(tempA);
            }else{
                it2++;
                mergedV.push_back(tempB);
            }           
        }
        count--;
    }while(count!=0);
    
    if(flagEvent){
        Rvalue = (double)(mergedV.at(M-1) + mergedV.at(M))/2;
        
    }else{
        Rvalue = (double)mergedV.at(M);
    }
    return Rvalue;
}

int main(void){
    vector<int> number1V ={1,2};
    vector<int> number2V ={3,4};
    double media =0 ;
    media = findMedianSortedArrays(number1V , number2V);
    cout <<"the meida :" << media << endl;
    return 0;
}
