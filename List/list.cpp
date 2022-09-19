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

int main() {
    int listval[3]={2,4,3};
    int listval2[3] = {5,6,4};
   
   	//************************************
	//	To assign value to ListNode
	//************************************
    ListNode * Header1 = new ListNode();
    ListNode * CurrentP;
    CurrentP = Header1;
    for(int i = 0 ; i < 3 ; i++){
        CurrentP->next = new ListNode();
        CurrentP = CurrentP->next ;
        CurrentP->val = listval[i];
        CurrentP->next = NULL;
    }
    
    CurrentP = Header1->next;
    do
    {
        cout <<"list value : "<< CurrentP->val <<endl;
        CurrentP = CurrentP->next;
    }while(CurrentP !=NULL);


    ListNode * Header2 = new ListNode();
    CurrentP = Header2;
    for(int i = 0 ; i < 3 ;i++)
    {
        CurrentP->next = new ListNode();
        CurrentP = CurrentP->next ;
        CurrentP->val = listval2[i];
        CurrentP->next = NULL;
    }
    
    CurrentP = Header2;
    do{
        cout << "List 2 vale :" << CurrentP->val <<endl;
        CurrentP = CurrentP->next;
    }while(CurrentP != NULL);
    
	//************************************
	//	To Delete listvale2 and Header2
	//************************************
	CurrentP = Header1;
	while(CurrentP->next != NULL){
		Header1 = CurrentP->next;
		delete CurrentP;
		CurrentP = Header1;
	}
	delete Header1 ;
	CurrentP =NULL;

	//************************************
	//	To Delete listvale2 and Header2
	//************************************
	CurrentP = Header2;
	while(CurrentP->next!= NULL)
	{
		Header2 = CurrentP->next;
		delete CurrentP;
		CurrentP = Header2;
	}
	delete CurrentP ;
	

    return 0;
    
}
