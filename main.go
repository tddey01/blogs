package main

import "github.com/tddey01/blogs/InitRouter"

func main() {
	router := InitRouter.SetupRouter()
	_ = router.Run(":8080")
}
