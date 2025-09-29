package main

import (
	"fmt"
	"time"
)

func main() {
	// fmt.Println("Hello World")
	// nums := []int{}
	// nums = append(nums, 1)
	// nums = append(nums, 2)
	// fmt.Println(nums[1])

	// m := map[string]string{
	// 	"a": "1",
	// 	"b": "2",
	// }

	// x, y := m["a"]
	// fmt.Println(y)
	// fmt.Println(x)

	// for i:=10; i < 15; i++ {
	//     fmt.Println(i)
	// }

	// i:= 1
	// for i < 10 {
	//     fmt.Println(i)
	//     i++
	// }

	// // var arr []int
	// arr := make([]uint8, 1, 100)
	// arr = append(arr, 10)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 20)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 30)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 40)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 50)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 60)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 60)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 60)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 60)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 60)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))
	// arr = append(arr, 60)
	// fmt.Printf("Len %v, cap %v \n", len(arr), cap(arr))

	// fmt.Printf("%v", arr[len(arr)-1])

	mymap := map[uint8]uint8{1: 10, 2: 20, 3: 30}

	for k, v := range mymap {
		fmt.Printf("k %v v %v \n", k, v)
	}

	timeLoop := func(slice []uint8, n int) time.Duration {
		var t0 = time.Now()
		for len(slice) < n {
			slice = append(slice, 1)
		}
		t1 := time.Now()
		return t1.Sub(t0)
	}

	v := 1000000000
	var ar1 []uint8
	ar2 := make([]uint8, 0, v)

	fmt.Printf("without make:: %v", timeLoop(ar1, v))
	fmt.Printf("with make: %v", timeLoop(ar2, v))

}
