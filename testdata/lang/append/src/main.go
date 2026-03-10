package main

func main() {
	{
		// make, len, cap.
		nums := make([]int, 3)
		n := nums[1] // 0
		nums[1] = 42
		l := len(nums) // 3
		c := cap(nums)
		_ = n
		_ = l
		_ = c
	}

	{
		// Append values.
		nums := make([]int, 0, 3)
		nums = append(nums, 1)
		nums = append(nums, 2, 3)
		l := len(nums) // 3
		c := cap(nums) // 3
		_ = l
		_ = c

		// Resizing slices beyond their initial capacity with append() panics.
		// nums = append(nums, 4)
	}

	{
		// Append slices.
		nums := make([]int, 0, 8)
		numsa := []int{1, 2}
		nums = append(nums, numsa...)
		nums = append(nums, []int{3, 4}...)
		l := len(nums) // 4
		if l != 4 {
			panic("want l = 4")
		}
		if nums[3] != 4 {
			panic("want nums[3] = 4")
		}
	}
}
