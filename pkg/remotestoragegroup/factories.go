package remotestoragegroup

import (
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/mergestrategy"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/queryfailurestrategy"
)

func QueryFailureStrategyFactory(strategyName string) OnQueryFailureStrategy {
	switch strategyName {
	case config.FailStrategyFailAll:
		return &queryfailurestrategy.FailAllStrategy{}
	case config.FailStrategyPartialResponse:
		return &queryfailurestrategy.PartialResponseStrategy{}
	default:
		panic("unrecognized failure strategy")
	}
}

func MergeStrategyFactory(conf config.MergeStrategyConfig) MergeStrategy {
	switch conf.Strategy {
	case config.MergeStrategyAlwaysMerge:
		return mergestrategy.NewAlwaysMergeStrategy()
	case config.MergeStrategyKeepBiggest:
		return mergestrategy.NewKeepBiggestMergeStrategy()
	default:
		panic("unrecognized merge strategy")
	}
}
