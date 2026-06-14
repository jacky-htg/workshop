package mockdata

import (
	"fmt"
	"workshop/pkg/app"
)

func Api() {
	privateRoutes := []app.RouteDefinition{
		{Method: "GET", Path: "/accesses", Group: "accesses", Alias: "accesses::list", HandlerFunc: nil},
	}

	fmt.Println(privateRoutes)
}
