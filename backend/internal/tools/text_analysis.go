package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// TextAnalysisTool 文本分析工具
type TextAnalysisTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewTextAnalysisTool 创建文本分析工具
func NewTextAnalysisTool(sandbox *sanbox.PythonSandbox) *TextAnalysisTool {
	return &TextAnalysisTool{
		sandbox: sandbox,
		name:    "text_analysis",
		desc:    "文本分析工具，支持情感分析、关键词提取、词频统计、文本清洗等功能。",
	}
}

// Info 返回工具信息
func (t *TextAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"operation": {
					Type:     schema.String,
					Desc:     "分析操作类型: sentiment（情感分析）, keywords（关键词提取）, wordfreq（词频统计）, clean（文本清洗）, summary（文本摘要）",
					Required: true,
				},
				"text_column": {
					Type:     schema.String,
					Desc:     "包含文本数据的列名",
					Required: true,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "数据文件路径",
					Required: false,
				},
				"language": {
					Type:     schema.String,
					Desc:     "文本语言（zh为中文，en为英文，默认自动检测）",
					Required: false,
				},
				"max_features": {
					Type:     schema.Number,
					Desc:     "最大特征数量（用于关键词提取等，默认100）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *TextAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Operation   string `json:"operation"`
		TextColumn  string `json:"text_column"`
		FilePath    string `json:"file_path,omitempty"`
		Language    string `json:"language,omitempty"`
		MaxFeatures int    `json:"max_features,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	if args.MaxFeatures <= 0 {
		args.MaxFeatures = 100
	}

	if args.Language == "" {
		args.Language = "auto"
	}

	code := t.generateTextAnalysisCode(args.Operation, args.TextColumn, args.FilePath, args.Language, args.MaxFeatures)

	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "文本分析失败: " + result.Error, nil
	}

	return result.Stdout, nil
}

// generateTextAnalysisCode 生成文本分析代码
func (t *TextAnalysisTool) generateTextAnalysisCode(operation, textColumn, filePath, language string, maxFeatures int) string {
	code := fmt.Sprintf(`
import pandas as pd
import numpy as np
import json
import re
from collections import Counter

# 数据加载
%s

# 检查文本列是否存在
if '%s' not in df.columns:
    print(json.dumps({"error": "指定的文本列不存在"}, ensure_ascii=False))
    exit()

text_data = df['%s'].dropna().astype(str)

`, t.getDataLoadCode(filePath), textColumn, textColumn)

	switch operation {
	case "sentiment":
		code += t.generateSentimentAnalysisCode(language)
	case "keywords":
		code += t.generateKeywordExtractionCode(language, maxFeatures)
	case "wordfreq":
		code += t.generateWordFrequencyCode(language, maxFeatures)
	case "clean":
		code += t.generateTextCleaningCode(language)
	case "summary":
		code += t.generateTextSummaryCode(language)
	default:
		code += t.generateWordFrequencyCode(language, maxFeatures) // 默认词频统计
	}

	return code
}

// generateSentimentAnalysisCode 生成情感分析代码
func (t *TextAnalysisTool) generateSentimentAnalysisCode(language string) string {
	return `
# 简单的情感分析（基于关键词）
def simple_sentiment_analysis(text, lang='zh'):
    """简单的情感分析"""
    if lang == 'zh':
        positive_words = ['好', '棒', '优秀', '满意', '喜欢', '推荐', '赞', '不错', '完美', '满意']
        negative_words = ['差', '糟糕', '失望', '不满', '讨厌', '垃圾', '烂', '问题', '故障', '投诉']
    else:
        positive_words = ['good', 'great', 'excellent', 'amazing', 'love', 'perfect', 'wonderful', 'fantastic']
        negative_words = ['bad', 'terrible', 'awful', 'hate', 'worst', 'horrible', 'disgusting', 'disappointing']
    
    text_lower = text.lower()
    pos_count = sum(1 for word in positive_words if word in text_lower)
    neg_count = sum(1 for word in negative_words if word in text_lower)
    
    if pos_count > neg_count:
        return 'positive'
    elif neg_count > pos_count:
        return 'negative'
    else:
        return 'neutral'

# 执行情感分析
sentiment_results = []
for text in text_data:
    sentiment = simple_sentiment_analysis(text, language)
    sentiment_results.append(sentiment)

# 统计结果
sentiment_counts = Counter(sentiment_results)
result = {
    "operation": "sentiment_analysis",
    "total_texts": len(sentiment_results),
    "sentiment_distribution": dict(sentiment_counts),
    "sentiment_percentage": {k: round(v/len(sentiment_results)*100, 2) for k, v in sentiment_counts.items()}
}

print(json.dumps(result, ensure_ascii=False, indent=2))
`
}

// generateKeywordExtractionCode 生成关键词提取代码
func (t *TextAnalysisTool) generateKeywordExtractionCode(language string, maxFeatures int) string {
	return fmt.Sprintf(`
# 关键词提取
import jieba
import jieba.analyse

def extract_keywords(texts, max_features=%d, lang='zh'):
    """提取关键词"""
    all_text = ' '.join(texts)
    
    if lang == 'zh':
        # 中文关键词提取
        keywords = jieba.analyse.extract_tags(all_text, topK=max_features, withWeight=True)
        return [(word, weight) for word, weight in keywords]
    else:
        # 英文关键词提取（简单版本）
        from collections import Counter
        import re
        
        # 简单的英文词汇提取
        words = re.findall(r'\b[a-zA-Z]{3,}\b', all_text.lower())
        # 移除常见停用词
        stop_words = {'the', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by', 'from', 'up', 'about', 'into', 'through', 'during', 'before', 'after', 'above', 'below', 'between', 'among', 'is', 'are', 'was', 'were', 'be', 'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could', 'should', 'may', 'might', 'must', 'can', 'this', 'that', 'these', 'those'}
        words = [word for word in words if word not in stop_words]
        
        word_counts = Counter(words)
        return [(word, count) for word, count in word_counts.most_common(max_features)]

# 执行关键词提取
keywords = extract_keywords(text_data.tolist(), %d, '%s')

result = {
    "operation": "keyword_extraction",
    "total_texts": len(text_data),
    "keywords": [{"word": word, "weight": weight} for word, weight in keywords[:50]]  # 返回前50个关键词
}

print(json.dumps(result, ensure_ascii=False, indent=2))
`, maxFeatures, maxFeatures, language)
}

// generateWordFrequencyCode 生成词频统计代码
func (t *TextAnalysisTool) generateWordFrequencyCode(language string, maxFeatures int) string {
	return fmt.Sprintf(`
# 词频统计
from collections import Counter
import re

def word_frequency_analysis(texts, max_features=%d, lang='zh'):
    """词频分析"""
    all_text = ' '.join(texts)
    
    if lang == 'zh':
        try:
            import jieba
            words = jieba.lcut(all_text)
            # 过滤长度小于2的词和标点符号
            words = [word for word in words if len(word) >= 2 and word.isalnum()]
        except ImportError:
            # 如果没有jieba，使用简单的字符分割
            words = re.findall(r'[\u4e00-\u9fff]+', all_text)
    else:
        # 英文词频分析
        words = re.findall(r'\b[a-zA-Z]{3,}\b', all_text.lower())
        # 移除停用词
        stop_words = {'the', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by'}
        words = [word for word in words if word not in stop_words]
    
    word_counts = Counter(words)
    return word_counts.most_common(max_features)

# 执行词频分析
word_freq = word_frequency_analysis(text_data.tolist(), %d, '%s')

result = {
    "operation": "word_frequency",
    "total_texts": len(text_data),
    "total_words": sum(count for _, count in word_freq),
    "unique_words": len(word_freq),
    "word_frequency": [{"word": word, "count": count} for word, count in word_freq]
}

print(json.dumps(result, ensure_ascii=False, indent=2))
`, maxFeatures, maxFeatures, language)
}

// generateTextCleaningCode 生成文本清洗代码
func (t *TextAnalysisTool) generateTextCleaningCode(language string) string {
	return `
# 文本清洗
import re

def clean_text(text, lang='zh'):
    """文本清洗"""
    # 移除HTML标签
    text = re.sub(r'<[^>]+>', '', text)
    
    # 移除URL
    text = re.sub(r'http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+', '', text)
    
    # 移除邮箱
    text = re.sub(r'\S+@\S+', '', text)
    
    # 移除多余的空白字符
    text = re.sub(r'\s+', ' ', text)
    
    if lang == 'zh':
        # 保留中文、英文、数字和基本标点
        text = re.sub(r'[^\u4e00-\u9fff\w\s.,!?;:()"""''【】「」（）]', '', text)
    else:
        # 保留英文、数字和基本标点
        text = re.sub(r'[^\w\s.,!?;:()"]', '', text)
    
    return text.strip()

# 执行文本清洗
cleaned_texts = []
for text in text_data:
    cleaned = clean_text(text, language)
    cleaned_texts.append(cleaned)

# 统计清洗效果
original_lengths = [len(text) for text in text_data]
cleaned_lengths = [len(text) for text in cleaned_texts]

result = {
    "operation": "text_cleaning",
    "total_texts": len(text_data),
    "average_original_length": round(np.mean(original_lengths), 2),
    "average_cleaned_length": round(np.mean(cleaned_lengths), 2),
    "cleaning_rate": round((1 - np.mean(cleaned_lengths)/np.mean(original_lengths)) * 100, 2),
    "sample_results": [
        {"original": orig[:100], "cleaned": clean[:100]} 
        for orig, clean in zip(text_data[:3].tolist(), cleaned_texts[:3])
    ]
}

print(json.dumps(result, ensure_ascii=False, indent=2))
`
}

// generateTextSummaryCode 生成文本摘要代码
func (t *TextAnalysisTool) generateTextSummaryCode(language string) string {
	return `
# 文本摘要（简单版本）
def simple_text_summary(texts, max_sentences=3):
    """简单的文本摘要"""
    all_text = ' '.join(texts)
    
    # 按句子分割
    import re
    sentences = re.split(r'[.!?。！？]', all_text)
    sentences = [s.strip() for s in sentences if len(s.strip()) > 10]
    
    # 计算句子长度分布
    lengths = [len(s) for s in sentences]
    avg_length = np.mean(lengths)
    
    # 选择接近平均长度的句子作为摘要
    summary_sentences = []
    for sentence in sentences:
        if abs(len(sentence) - avg_length) < avg_length * 0.3:
            summary_sentences.append(sentence)
            if len(summary_sentences) >= max_sentences:
                break
    
    return summary_sentences

# 执行文本摘要
summary = simple_text_summary(text_data.tolist())

result = {
    "operation": "text_summary",
    "total_texts": len(text_data),
    "total_characters": sum(len(text) for text in text_data),
    "summary_sentences": summary,
    "summary": ' '.join(summary)
}

print(json.dumps(result, ensure_ascii=False, indent=2))
`
}

// getDataLoadCode 获取数据加载代码
func (t *TextAnalysisTool) getDataLoadCode(filePath string) string {
	if filePath != "" {
		return fmt.Sprintf(`
try:
    if '%s'.endswith('.csv'):
        df = pd.read_csv('%s')
    elif '%s'.endswith(('.xlsx', '.xls')):
        df = pd.read_excel('%s')
    elif '%s'.endswith('.json'):
        df = pd.read_json('%s')
    else:
        print(json.dumps({"error": "不支持的文件格式"}xw, ensure_ascii=False))
        exit()
except Exception as e:
    print(json.dumps({"error": f"数据加载失败: {str(e)}"}, ensure_ascii=False))
    exit()
`, filePath, filePath, filePath, filePath, filePath, filePath)
	}
	return "# 假设数据已经加载到df变量中\nif 'df' not in locals():\n    print(json.dumps({'error': '数据未加载'}, ensure_ascii=False))\n    exit()"
}
