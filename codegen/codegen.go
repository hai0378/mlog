package main

import (
	"github.com/mlogclub/simple"

	"github.com/mlogclub/mlog/model"
)

func main() {
	simple.Generate("./", "github.com/mlogclub/mlog", simple.GetGenerateStruct(&model.Project{}))
}
