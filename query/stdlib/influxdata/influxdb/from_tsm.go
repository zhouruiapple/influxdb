package influxdb

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/flux/values/objects"
)

const (
	streamArg   = "stream"
	FromTSMKind = "fromTSM"
)

var fromTSM = values.NewFunction(
	"fromTSM",
	semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			streamArg: semantic.Stream,
		},
		PipeArgument: streamArg,
		Required:     semantic.LabelSet{streamArg},
		Return:       semantic.NewArrayPolyType(objects.TableType),
	}),
	func(args values.Object) (values.Value, error) {
		return nil, nil
	}, false,
)

func init() {
	flux.RegisterPackageValue("influxdata/influxdb", FromTSMKind, fromTSM)
}
