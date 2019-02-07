package blueprint

import (
	"fmt"
	"strings"
)

func (o OverviewType) String() string {
	return fmt.Sprintf("%s", *o.Name)
}

func (t TreeStructureType) String() string {
	return "-"
}

func (mm MetricPropertyType) String() string {
	var builder strings.Builder

	if mm.Value != nil {

		switch u := mm.Unit; {
		case u == "tuple":
			fallthrough
		case u == "number":
			builder.WriteString(fmt.Sprintf("v:%0.0f, u:%s,", *mm.Value, u))
		default:
			builder.WriteString(fmt.Sprintf("b:%s, u:%s,", *mm.Value, u))
		}

	} else {
		builder.WriteString(fmt.Sprintf("u:%s,", mm.Unit))
	}

	if mm.Maximum != nil {
		builder.WriteString(fmt.Sprintf("max:%0.2f,", *mm.Maximum))
	}

	if mm.Minimum != nil {
		builder.WriteString(fmt.Sprintf("min:%0.2f,", *mm.Minimum))
	}

	return builder.String()

}
