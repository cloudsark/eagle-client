// +build !windows,cgo

package linux

import (
	"eagle-client/utils"
	"errors"
)

// #include <stdlib.h>
import "C"

// Reference: man 3 getloadavg
func get() (*Stats, error) {
	var loadavgs [3]C.double
	ret := C.getloadavg(&loadavgs[0], 3)
	if ret != 3 {
		return nil, errors.New("failed to get loadavg")
	}
	return &Stats{
		HostName:  utils.GetHostName(),
		Loadavg1:  float64(loadavgs[0]),
		Loadavg5:  float64(loadavgs[1]),
		Loadavg15: float64(loadavgs[2]),
	}, nil
}
