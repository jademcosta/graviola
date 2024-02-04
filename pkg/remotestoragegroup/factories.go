package remotestoragegroup

import (
	"github.com/jademcosta/graviola/pkg/config"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/mergestrategy"
	"github.com/jademcosta/graviola/pkg/remotestoragegroup/queryfailurestrategy"
)

func QueryFailureStrategyFactory(strategyName string) OnQueryFailureStrategy {
	switch strategyName {
	case config.StrategyFailAll:
		return &queryfailurestrategy.FailAllStrategy{}
	case config.StrategyPartialResponse:
		return &queryfailurestrategy.PartialResponseStrategy{}
	default:
		panic("unrecognized failure strategy")
	}
}

func MergeStrategyFactory(strategyName string) MergeStrategy {

	switch strategyName {
	case config.MergeStrategyAlwaysMerge:
		return mergestrategy.NewAlwaysMergeStrategy()
	case config.MergeStrategyKeepBiggest:
		return mergestrategy.NewKeepBiggestMergeStrategy()
	default:
		panic("unrecognized merge strategy")
	}
}
