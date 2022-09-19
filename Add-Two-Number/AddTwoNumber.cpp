/*
 *You are given two non-empty linked lists representing two non-negative integers. 
 * The digits are stored in reverse order, and each of their nodes contains a single digit.
 * Add the two numbers and return the sum as a linked list.
 * You may assume the two numbers do not contain any leading zero, except the number 0 itself. 
 * 
 * Example 1:
 * Input: l1 = [2,4,3], l2 = [5,6,4]
 * Output: [7,0,8]
 * Explanation: 342 + 465 = 807.
 * 
 * Example 2:
 * Input: l1 = [0], l2 = [0]
 * Output: [0]
 * 
 * Example 3:
 * Input: l1 = [9,9,9,9,9,9,9], l2 = [9,9,9,9]
 * Output: [8,9,9,9,0,0,0,1]
 * 
 * Constraints:
 * The number of nodes in each linked list is in the range [1, 100].
 * 0 <= Node.val <= 9
 * It is guaranteed that the list represents a number that does not have leading zeros.
 */
#include <algorithm>
#include <vector>
#include <iostream>
#include <string>
using namespace std;
struct ListNode {
    int val;
    ListNode *next;
    ListNode() : val(0), next(nullptr) {}
    ListNode(int x) : val(x), next(nullptr) {}
    ListNode(int x, ListNode *next) : val(x), next(next) {}
};
ListNode* addTwoNumbers(ListNode* l1, ListNode* l2){
    ListNode* l1header = l1 ;
    ListNode* l2header = l2;
    ListNode* Result= new ListNode();;
    ListNode* current;
    ListNode* header ;

    int increment = 0 ;
    int value;
    int l1counter= 0 ;
    int l2counter= 0 ;
    int length = 0 ;
    current = l1header;
    while(current != NULL){
        l1counter++;
        current = current->next;
    }
    current = l2header;
    while(current != NULL){
        l2counter++;
        current = current->next;
    }
    length = (l1counter > l2counter) ? l1counter: l2counter;
    cout << "length " << length  << endl;
    current = Result;
    for(int i = 0 ; i < length ; i++){
        int lVal1 = (i >= l1counter) ? 0 : l1header->val;
        int lVal2 = (i >= l2counter) ? 0 : l2header->val;
        current->next = new ListNode();
        current = current->next;
        current->val = (lVal1 + lVal2);
        current->next = NULL;
        if(i < l1counter)
            l1header = l1header->next;
        if(i < l2counter)
            l2header = l2header->next;
    }
    header  = Result;
    current = header->next;
    do{
        cout << "result :: "<< current->val<< endl;
        current = current->next;
    }while(current !=NULL);

    header  = Result;
    current = header->next;
    cout <<"Result :";
    do{
        if (increment >= 1){
            if(current->next ==NULL && (current->val+1) >= 10){
                current->val = current->val +1;
                current->val = current->val %10;
                current->next = new ListNode();
                current =  current->next;
                current->val = 1;
                current->next = NULL;
            }else{
                if(current->val >10){
                    current->val = current->val + 1;
                    current->val = current->val %10;
                }else if (current->val == 10){
                    current->val = 0;
                }else{
                    current->val = current->val + 1;
                    if(current->val ==  10){
                        current->val = 0;
                    }else{
                        increment--;
                    }
                }
            }//end if else
        }else{
            if(current->val >= 10){
            current->val = current->val %10;
            increment++;

            }
        }
        current = current->next;
    }while(current !=NULL);
    
    header  = Result;
    current = header->next;
    do{
        cout << " "<< current->val;
        current = current->next;
    }while(current !=NULL);
    cout << endl;
    return Result;
}
int main() {
    //ListNode* l1header = new ListNode(2,new ListNode(4,new ListNode(3,NULL)));
    //ListNode* l2header = new ListNode(5,new ListNode(6,new ListNode(4,NULL)));
    ListNode* l1header = new ListNode(9,new ListNode(9,new ListNode(9,new ListNode(9, new ListNode(9,new ListNode(9,NULL))))));
    ListNode* l2header = new ListNode(9,new ListNode(9,new ListNode(9,NULL)));
    ListNode* HeaderP;
    ListNode* CurrentP;

    HeaderP =  addTwoNumbers(l1header, l2header);
    CurrentP = HeaderP->next;

    cout << "Return Result :" ;
    do{
        cout << " "<< CurrentP->val;
        CurrentP = CurrentP->next;
    }while(CurrentP !=NULL);
    cout << endl;

    //*************************
    //  To Delete ListNode
    //*************************
    CurrentP = l1header;
    while(CurrentP->next != NULL){
        l1header = CurrentP->next;
        delete CurrentP;
        CurrentP =l1header;
    }
    delete l1header;
    
    CurrentP = l2header;
    while(CurrentP->next !=NULL)
    {
        l2header = CurrentP->next;
        delete CurrentP;
        CurrentP = l2header;
    }
    delete l2header;
    

    CurrentP = HeaderP;
    while(CurrentP->next != NULL){
        HeaderP = CurrentP->next;
        delete CurrentP;
        CurrentP = HeaderP;
    }
    delete HeaderP;
    
    
    

   return 0 ;

}
