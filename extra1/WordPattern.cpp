#include <vector>
#include <string>
#include <unordered_map>

class Solution {
public:
	bool wordPattern(std::string pattern, std::string s) {
		std::unordered_map<char, std::string> char_to_str_map;
		std::unordered_map<std::string, char> str_to_char_map;
		std::vector<std::string> s_splits;
		size_t start;
		size_t end;

		start = 0;
		end  = s.find_first_of(' ');
		while (end != std::string::npos)
		{
			s_splits.push_back(s.substr(start, end - start));
			start = end + 1;
			end = s.find_first_of(' ', start);
		}
		s_splits.push_back(s.substr(start));
		if (pattern.size() != s_splits.size())
			return false;
		for (int i = 0; i < pattern.size(); i++)
		{
			if (char_to_str_map.count(pattern[i]) && char_to_str_map.at(pattern[i]) != s_splits[i])
				return false;
			if (str_to_char_map.count(s_splits[i]) && str_to_char_map.at(s_splits[i]) != pattern[i])
				return false;
			char_to_str_map[pattern[i]] = s_splits[i];
			str_to_char_map[s_splits[i]] = pattern[i];
		}
		return true;
	}
};