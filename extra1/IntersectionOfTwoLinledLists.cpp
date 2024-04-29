#include <cstddef>

class Solution {
struct ListNode {
	int val;
	ListNode *next;
	ListNode(int x) : val(x), next(NULL) {}
};

public:
	size_t lstlen(ListNode* head) {
		int res = 0;
		ListNode* currentNode = head;
		while (currentNode) {
			currentNode = currentNode->next;
			res++;
		}
		return res;
	}

public:
	ListNode* getIntersectionNode(ListNode* headA, ListNode* headB) {
		size_t a_len = lstlen(headA);
		size_t b_len = lstlen(headB);

		if (a_len > b_len)
		{
			while (a_len-- > b_len)
				headA = headA->next;
		}
		else 
		{
			while (b_len-- > a_len)
				headB = headB->next;
		}
		while (headA && headB)
		{
			if (headA == headB)
				return headA;
			headA = headA->next;
			headB = headB->next;
		}
		return NULL;
	}
};