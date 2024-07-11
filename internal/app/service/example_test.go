package service

import "fmt"

func ExampleService_ValidateURL() {
	s := &service{}
	err := s.ValidateURL("https://practicum.yandex.ru/")
	fmt.Println(err)

	err = s.ValidateURL("1234567")
	fmt.Println(err)

	// Output:
	// <nil>
	// url is invalid
}
