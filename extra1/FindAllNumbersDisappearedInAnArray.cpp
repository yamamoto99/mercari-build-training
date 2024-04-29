#include <vector>
#include <unordered_set>
#include <algorithm>

class Solution {
public:
	std::vector<int> findDisappearedNumbers(std::vector<int>& nums) {
		std::vector<int> res;
		for (int i : nums)
		{
			int index = abs(i) - 1;
			if (nums[index] > 0)
				nums[index] = -nums[index];
		}
		for (int i = 0; i < nums.size(); i++)
		{
			if (nums[i] > 0)
				res.push_back(i + 1);
		}
		return res;
	}
};