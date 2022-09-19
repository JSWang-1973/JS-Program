/*
 * Given an array of integers nums and an integer target, 
 * return indices of the two numbers such that they add up to target.
 * 
 * You may assume that each input would have exactly one solution, and you may not use the same element twice.
 * You can return the answer in any order.
 * 
 * Example 1:
 * Input: nums = [2,7,11,15], target = 9
 * Output: [0,1]
 * Explanation: Because nums[0] + nums[1] == 9, we return [0, 1].
 * 
 * Example 2:
 * Input: nums = [3,2,4], target = 6
 * Output: [1,2]
 * 
 * Example 3:
 * Input: nums = [3,3], target = 6
 * Output: [0,1]
 * 
 * Constraints:
 * 2 <= nums.length <= 104
 * -109 <= nums[i] <= 109
 * -109 <= target <= 109
 * Only one valid answer exists.
*/
#include <algorithm>
#include <vector>
#include <iostream>
#include <string>
using namespace std;

vector<int> twoSum(vector<int>& nums, int target) {
    std::vector<int> Rvec;
    vector<int>::iterator Iiterator = nums.begin();
    vector<int>::iterator Jiterator = nums.begin();
    int a= 0 , b =0 ,sum =0 ;

    for (int i= 0 ; i < nums.size(); i++){
        a = *(Iiterator+i);
        for (int j=i+1; j < nums.size(); j++){
            b = *(Jiterator+j);
            sum = a + b ;
            if (sum == target){
                Rvec.push_back(i);
                Rvec.push_back(j);
                return Rvec;
            }
        }
    }
    Rvec.push_back(0);
    Rvec.push_back(0);
    return Rvec;
}

int main() {
    std::vector<int> resultV;
    std::vector<int> nums = {2,7,11,15};
    resultV = twoSum(nums , 13);
    cout << "The size of vector is :" <<   "We return ["<<resultV[0]<< " "<<resultV[1]<<"]" << endl;
   return 0 ;

}
