package utils_test

import (
	"fmt"
	"github.com/helloh2o/lucky/utils"
)

func ExampleRandGroup() {
	i := utils.RandGroup(0, 0, 50, 50)
	switch i {
	case 2, 3:
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleRandInterval() {
	v := utils.RandInterval(-1, 1)
	switch v {
	case -1, 0, 1:
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleRandIntervalN() {
	r := utils.RandIntervalN(-1, 0, 2)
	if r[0] == -1 && r[1] == 0 ||
		r[0] == 0 && r[1] == -1 {
		fmt.Println("ok")
	}

	// Output:
	// ok
}

func ExampleDeepCopy() {
	src := []int{1, 2, 3}

	var dst []int
	utils.DeepCopy(&dst, &src)

	for _, v := range dst {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
}

func ExampleDeepClone() {
	src := []int{1, 2, 3}

	dst := utils.DeepClone(src).([]int)

	for _, v := range dst {
		fmt.Println(v)
	}

	// Output:
	// 1
	// 2
	// 3
}
