/*
 * Given an integer array nums of length n and an integer target, 
 * find three integers in nums such that the sum is closest to target.
 * Return the sum of the three integers.
 * 
 * You may assume that each input would have exactly one solution.
 * Example 1:
 * Input: nums = [-1,2,1,-4], target = 1
 * Output: 2
 * Explanation: The sum that is closest to the target is 2. (-1 + 2 + 1 = 2).
 * 
 * Example 2:
 * Input: nums = [0,0,0], target = 1
 * Output: 0
 * 
 * Constraints:
 * 3 <= nums.length <= 1000
 * -1000 <= nums[i] <= 1000
 * -104 <= target <= 104
 */
#include <algorithm>
#include <vector>
#include <iostream>
#include <string>
using namespace std;
 
int threeSumClosest(vector<int>& nums, int target) {
    int distance = 1000;
    int closest;
    int sum;
    int temp;
    int size = nums.size();
    int Ivalue ;
    int Jvalue;
    int Kvalue ;

    vector<int>::iterator Iiterator = nums.begin();
    vector<int>::iterator Jiterator = nums.begin();
    vector<int>::iterator Kiterator = nums.begin();
    for(int i=0 ; i < size ;i++){
        Ivalue = *(Iiterator+i);
        for(int j=i+1;j < size; j++ ){
            Jvalue = *(Jiterator+j);
            for(int k=j+1; k < size ; k++){
                Kvalue = *(Kiterator+k);
                sum  = Ivalue + Jvalue + Kvalue;
                temp = abs(sum - target) ;
                cout << "Ivalue\t:" << Ivalue << "\tJvalue: " << Jvalue <<"\tKvalue\t: "<<Kvalue <<"\tSum\t: " << sum<<endl;
                if(temp < distance){
                    closest = sum;
                    distance = temp;
                }
            } 
        }
    }
    return closest;
}

int main(){
    vector<int>Input={-1,2,1,-4,5};
    int target = 1 ;
    int output;
    output = threeSumClosest(Input , target);
    cout << "the result is : "<< output << endl;
    return 0 ;
}
