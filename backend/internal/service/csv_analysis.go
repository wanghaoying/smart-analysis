package service

import (
	"fmt"
	"smart-analysis/internal/utils/llm"
)

// LLMAnalysisService 集成了新LLM封装的分析服务
type LLMAnalysisService struct {
	llmManager *llm.ClientManager
}

// NewLLMAnalysisService 创建集成LLM的分析服务
func NewLLMAnalysisService(llmManager *llm.ClientManager) *LLMAnalysisService {
	return &LLMAnalysisService{
		llmManager: llmManager,
	}
}

// GetAvailableLLMProviders 获取可用的LLM提供商
func (s *LLMAnalysisService) GetAvailableLLMProviders() []llm.LLMProvider {
	return s.llmManager.ListProviders()
}

// SetLLMProvider 设置默认LLM提供商（通过重新排序优先级）
func (s *LLMAnalysisService) SetLLMProvider(provider llm.LLMProvider) error {
	// 检查提供商是否可用
	providers := s.llmManager.ListProviders()
	for _, p := range providers {
		if p == provider {
			// 这里可以实现优先级逻辑
			return nil
		}
	}
	return fmt.Errorf("provider %s not available", provider)
}
