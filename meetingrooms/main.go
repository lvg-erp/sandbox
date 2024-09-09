package main

import (
	"fmt"
	"meetingrooms/algCalc"
)

func main() {
	//1
	//inputArray := [][]int{{0, 30}, {32, 43}, {45, 50}}
	//2
	inputArray := [][]int{{0, 30}, {10, 15}, {15, 20}}
	s := algCalc.Solution{}
	result := s.CanAttendMeetings(inputArray)
	res := algCalc.MinCountMeetingRooms(inputArray)

	fmt.Println(result, "result")
	fmt.Println(res)

}
