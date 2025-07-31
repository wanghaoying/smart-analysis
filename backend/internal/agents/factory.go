package agents

import (
	"context"
	"fmt"

	"smart-analysis/internal/types"
)

// AgentFactory 智能体工厂
type AgentFactory struct{}

// NewAgentFactory 创建智能体工厂
func NewAgentFactory() *AgentFactory {
	return &AgentFactory{}
}

// CreateAgent 创建智能体
func (f *AgentFactory) CreateAgent(ctx context.Context, agentType types.AgentType, config *types.AgentConfig) (types.Agent, error) {
	switch agentType {
	case types.AgentTypeMaster:
		return NewMasterAgent(ctx, config)
	case types.AgentTypePlanner:
		return NewPlannerAgent(ctx, config)
	case types.AgentTypeDataQuery:
		return NewDataQueryAgent(ctx, config)
	case types.AgentTypeDataAnalysis:
		return NewDataAnalysisAgent(ctx, config)
	case types.AgentTypeTrendForecast:
		return NewTrendForecastAgent(ctx, config)
	case types.AgentTypeAnomalyDetection:
		return NewAnomalyDetectionAgent(ctx, config)
	case types.AgentTypeAttributionAnalysis:
		return NewAttributionAnalysisAgent(ctx, config)
	case types.AgentTypeReact:
		return NewReactAgent(ctx, config)
	case types.AgentTypeAnalysis:
		return NewAnalysisAgent(ctx, config)
	case types.AgentTypeMulti:
		return NewMultiAgentManager(ctx, config)
	default:
		return nil, fmt.Errorf("不支持的智能体类型: %s", agentType)
	}
}

// CreateExpertAgent 创建专家智能体
func (f *AgentFactory) CreateExpertAgent(ctx context.Context, agentType types.AgentType, config *types.AgentConfig) (types.ExpertAgent, error) {
	switch agentType {
	case types.AgentTypeDataQuery:
		return NewDataQueryAgent(ctx, config)
	case types.AgentTypeDataAnalysis:
		return NewDataAnalysisAgent(ctx, config)
	case types.AgentTypeTrendForecast:
		return NewTrendForecastAgent(ctx, config)
	case types.AgentTypeAnomalyDetection:
		return NewAnomalyDetectionAgent(ctx, config)
	case types.AgentTypeAttributionAnalysis:
		return NewAttributionAnalysisAgent(ctx, config)
	default:
		return nil, fmt.Errorf("不支持的专家智能体类型: %s", agentType)
	}
}

// GetAvailableAgentTypes 获取可用的智能体类型
func (f *AgentFactory) GetAvailableAgentTypes() []types.AgentType {
	return []types.AgentType{
		types.AgentTypeMaster,
		types.AgentTypePlanner,
		types.AgentTypeDataQuery,
		types.AgentTypeDataAnalysis,
		types.AgentTypeTrendForecast,
		types.AgentTypeAnomalyDetection,
		types.AgentTypeAttributionAnalysis,
		types.AgentTypeReact,
		types.AgentTypeAnalysis,
		types.AgentTypeMulti,
	}
}

// GetExpertAgentTypes 获取专家智能体类型
func (f *AgentFactory) GetExpertAgentTypes() []types.AgentType {
	return []types.AgentType{
		types.AgentTypeDataQuery,
		types.AgentTypeDataAnalysis,
		types.AgentTypeTrendForecast,
		types.AgentTypeAnomalyDetection,
		types.AgentTypeAttributionAnalysis,
	}
}

// GetAgentCapabilities 获取智能体能力描述
func (f *AgentFactory) GetAgentCapabilities(agentType types.AgentType) []string {
	switch agentType {
	case types.AgentTypeMaster:
		return []string{
			"意图识别和分析",
			"查询重写和优化",
			"数据模式理解",
			"需求解析",
		}
	case types.AgentTypePlanner:
		return []string{
			"任务规划和分解",
			"执行计划创建",
			"任务调度和管理",
			"专家智能体协调",
		}
	case types.AgentTypeDataQuery:
		return []string{
			"数据查询与筛选",
			"SQL查询生成",
			"数据过滤和聚合",
			"多表关联查询",
			"数据预览和统计",
		}
	case types.AgentTypeDataAnalysis:
		return []string{
			"描述性统计分析",
			"相关性分析",
			"分布分析",
			"对比分析",
			"数据可视化",
			"数据预处理",
			"特征工程",
		}
	case types.AgentTypeTrendForecast:
		return []string{
			"时间序列分析",
			"趋势预测",
			"季节性分析",
			"周期性检测",
			"回归分析",
			"ARIMA建模",
			"指数平滑",
			"机器学习预测",
		}
	case types.AgentTypeAnomalyDetection:
		return []string{
			"异常值检测",
			"离群点分析",
			"时间序列异常检测",
			"统计异常检测",
			"机器学习异常检测",
			"异动根因分析",
			"异常模式识别",
			"实时异常监控",
		}
	case types.AgentTypeAttributionAnalysis:
		return []string{
			"因果关系分析",
			"根因分析",
			"贡献度分析",
			"影响因子识别",
			"特征重要性分析",
			"相关性与因果性分析",
			"回归分析",
			"变化归因分析",
		}
	case types.AgentTypeReact:
		return []string{
			"工具调用和执行",
			"反应式推理",
			"多步骤任务执行",
			"动态工具选择",
		}
	case types.AgentTypeAnalysis:
		return []string{
			"简化数据分析",
			"Python代码执行",
			"基础统计分析",
		}
	case types.AgentTypeMulti:
		return []string{
			"多智能体协作",
			"智能任务分配",
			"流程自动化",
			"综合分析能力",
		}
	default:
		return []string{"未知智能体类型"}
	}
}
