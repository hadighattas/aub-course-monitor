package main

type course struct {
	crn string //crn of the monitored course

	telegramIds []int64 //telegramIds of users monitoring the course

}

// MonitorMessage --
type MonitorMessage struct {
	crn string

	spotsLeft int
}

// AddMessage --
type AddMessage struct {
	crn string

	telegramID int64
}

//Callback --
type Callback func(callbackValue int)

//indexOfInt64 --

func indexOfInt64(slice []int64, value int64) int {

	for i, v := range slice {

		if v == value {

			return i

		}

	}

	return -1

}

//indexOfString --

func indexOfString(slice []string, value string) int {

	for i, v := range slice {

		if v == value {

			return i

		}

	}

	return -1

}

//removeElementInt64 --

func removeElementInt64(slice []int64, index int) []int64 {

	return append(slice[:index], slice[index+1:]...)

}

//removeElementString --

func removeElementString(slice []string, index int) []string {

	return append(slice[:index], slice[index+1:]...)

}
