package provider

import "fmt"

type isString struct{}

func (e isString) CheckValue(value any) error {
	_, ok := value.(string)

	if !ok {
		return fmt.Errorf("value is not a string: %v", value)
	}

	return nil
}

func (e isString) String() string {
	return "isString"
}
