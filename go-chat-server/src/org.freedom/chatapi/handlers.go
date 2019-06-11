package chatapi

import "org.freedom/bootstrap"

var commandSetUserName bootstrap.CommandListener = func(data interface{}) interface{} {
	name, result := data.(string)
	if result {

	}
	return nil
}
