package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// DataPreprocessingTool 数据预处理工具
type DataPreprocessingTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewDataPreprocessingTool 创建数据预处理工具
func NewDataPreprocessingTool(sandbox *sanbox.PythonSandbox) *DataPreprocessingTool {
	return &DataPreprocessingTool{
		sandbox: sandbox,
		name:    "data_preprocessing",
		desc:    "数据预处理和特征工程工具，支持数据标准化、归一化、特征编码、特征选择和降维等操作。",
	}
}

// Info 返回工具信息
func (t *DataPreprocessingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"operation": {
					Type:     schema.String,
					Desc:     "预处理操作类型: normalize（归一化）, standardize（标准化）, encode（特征编码）, select（特征选择）, reduce（降维）",
					Required: true,
				},
				"columns": {
					Type: schema.Array,
					Desc: "要处理的列名",
					ElemInfo: &schema.ParameterInfo{
						Type: schema.String,
					},
					Required: false,
				},
				"file_path": {
					Type:     schema.String,
					Desc:     "数据文件路径",
					Required: false,
				},
				"parameters": {
					Type:     schema.String,
					Desc:     "操作参数（JSON格式）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *DataPreprocessingTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Operation  string   `json:"operation"`
		Columns    []string `json:"columns,omitempty"`
		FilePath   string   `json:"file_path,omitempty"`
		Parameters string   `json:"parameters,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 生成预处理代码
	code := t.generatePreprocessingCode(args.Operation, args.Columns, args.FilePath, args.Parameters)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "数据预处理失败: " + result.Error, nil
	}

	// 格式化结果
	return t.formatResult(result), nil
}

// generatePreprocessingCode 生成数据预处理代码
func (t *DataPreprocessingTool) generatePreprocessingCode(operation string, columns []string, filePath, parameters string) string {
	code := `
import pandas as pd
import numpy as np
import json
from sklearn.preprocessing import StandardScaler, MinMaxScaler, LabelEncoder, OneHotEncoder
from sklearn.feature_selection import SelectKBest, f_classif, mutual_info_classif
from sklearn.decomposition import PCA
import warnings
warnings.filterwarnings('ignore')

`

	// 添加数据加载代码
	if filePath != "" {
		code += fmt.Sprintf(`
# 加载数据
try:
    if '%s'.endswith('.csv'):
        df = pd.read_csv('%s')
    elif '%s'.endswith(('.xlsx', '.xls')):
        df = pd.read_excel('%s')
    elif '%s'.endswith('.json'):
        df = pd.read_json('%s')
    else:
        print("不支持的文件格式")
        exit()
except Exception as e:
    print(f"数据加载失败: {e}")
    exit()

`, filePath, filePath, filePath, filePath, filePath, filePath)
	} else {
		code += `
# 假设数据已加载到df变量中
if 'df' not in locals():
    print("ERROR: 数据未加载")
    exit()

`
	}

	// 解析参数
	if parameters != "" {
		code += fmt.Sprintf(`
# 解析参数
try:
    params = json.loads('''%s''')
except:
    params = {}

`, parameters)
	} else {
		code += "params = {}\n"
	}

	// 根据操作类型生成代码
	switch operation {
	case "normalize":
		code += t.generateNormalizeCode(columns)
	case "standardize":
		code += t.generateStandardizeCode(columns)
	case "encode":
		code += t.generateEncodeCode(columns)
	case "select":
		code += t.generateFeatureSelectionCode(columns)
	case "reduce":
		code += t.generateDimensionReductionCode(columns)
	default:
		code += `
print("不支持的操作类型")
`
	}

	code += `
# 输出结果
result = {
    "operation": "` + operation + `",
    "shape_before": df.shape,
    "shape_after": df.shape if 'df' in locals() else None,
    "columns_before": df.columns.tolist() if 'df' in locals() else [],
    "columns_after": df.columns.tolist() if 'df' in locals() else [],
    "success": True
}

print("预处理操作完成:")
print(json.dumps(result, ensure_ascii=False, indent=2))
`

	return code
}

// generateNormalizeCode 生成归一化代码
func (t *DataPreprocessingTool) generateNormalizeCode(columns []string) string {
	if len(columns) == 0 {
		return `
# 归一化所有数值列
numeric_cols = df.select_dtypes(include=[np.number]).columns
scaler = MinMaxScaler()
df[numeric_cols] = scaler.fit_transform(df[numeric_cols])
print(f"已归一化 {len(numeric_cols)} 个数值列")
`
	}

	return fmt.Sprintf(`
# 归一化指定列
target_cols = %s
existing_cols = [col for col in target_cols if col in df.columns and df[col].dtype in ['int64', 'float64']]
if existing_cols:
    scaler = MinMaxScaler()
    df[existing_cols] = scaler.fit_transform(df[existing_cols])
    print(f"已归一化列: {existing_cols}")
else:
    print("没有找到可归一化的数值列")
`, fmt.Sprintf("%q", columns))
}

// generateStandardizeCode 生成标准化代码
func (t *DataPreprocessingTool) generateStandardizeCode(columns []string) string {
	if len(columns) == 0 {
		return `
# 标准化所有数值列
numeric_cols = df.select_dtypes(include=[np.number]).columns
scaler = StandardScaler()
df[numeric_cols] = scaler.fit_transform(df[numeric_cols])
print(f"已标准化 {len(numeric_cols)} 个数值列")
`
	}

	return fmt.Sprintf(`
# 标准化指定列
target_cols = %s
existing_cols = [col for col in target_cols if col in df.columns and df[col].dtype in ['int64', 'float64']]
if existing_cols:
    scaler = StandardScaler()
    df[existing_cols] = scaler.fit_transform(df[existing_cols])
    print(f"已标准化列: {existing_cols}")
else:
    print("没有找到可标准化的数值列")
`, fmt.Sprintf("%q", columns))
}

// generateEncodeCode 生成特征编码代码
func (t *DataPreprocessingTool) generateEncodeCode(columns []string) string {
	if len(columns) == 0 {
		return `
# 编码所有类别列
categorical_cols = df.select_dtypes(include=['object']).columns
encoding_method = params.get('method', 'label')  # label or onehot

if encoding_method == 'onehot':
    df_encoded = pd.get_dummies(df, columns=categorical_cols)
    df = df_encoded
    print(f"已进行独热编码的列: {categorical_cols.tolist()}")
else:
    for col in categorical_cols:
        le = LabelEncoder()
        df[col] = le.fit_transform(df[col].astype(str))
    print(f"已进行标签编码的列: {categorical_cols.tolist()}")
`
	}

	return fmt.Sprintf(`
# 编码指定列
target_cols = %s
existing_cols = [col for col in target_cols if col in df.columns and df[col].dtype == 'object']
encoding_method = params.get('method', 'label')  # label or onehot

if existing_cols:
    if encoding_method == 'onehot':
        df_encoded = pd.get_dummies(df, columns=existing_cols)
        df = df_encoded
        print(f"已进行独热编码的列: {existing_cols}")
    else:
        for col in existing_cols:
            le = LabelEncoder()
            df[col] = le.fit_transform(df[col].astype(str))
        print(f"已进行标签编码的列: {existing_cols}")
else:
    print("没有找到可编码的类别列")
`, fmt.Sprintf("%q", columns))
}

// generateFeatureSelectionCode 生成特征选择代码
func (t *DataPreprocessingTool) generateFeatureSelectionCode(columns []string) string {
	return `
# 特征选择
target_col = params.get('target_column')
k_features = params.get('k_features', 10)
method = params.get('method', 'f_classif')  # f_classif or mutual_info

if target_col and target_col in df.columns:
    X = df.drop(columns=[target_col])
    y = df[target_col]
    
    # 只选择数值特征
    numeric_features = X.select_dtypes(include=[np.number])
    
    if len(numeric_features.columns) > 0:
        if method == 'mutual_info':
            selector = SelectKBest(score_func=mutual_info_classif, k=min(k_features, len(numeric_features.columns)))
        else:
            selector = SelectKBest(score_func=f_classif, k=min(k_features, len(numeric_features.columns)))
        
        X_selected = selector.fit_transform(numeric_features, y)
        selected_features = numeric_features.columns[selector.get_support()]
        
        # 重构数据框
        df_selected = pd.DataFrame(X_selected, columns=selected_features)
        df_selected[target_col] = y.values
        df = df_selected
        
        print(f"特征选择完成，保留了 {len(selected_features)} 个特征: {selected_features.tolist()}")
    else:
        print("没有找到数值特征进行选择")
else:
    print("需要指定目标列进行特征选择")
`
}

// generateDimensionReductionCode 生成降维代码
func (t *DataPreprocessingTool) generateDimensionReductionCode(columns []string) string {
	return `
# 主成分分析降维
n_components = params.get('n_components', 2)
method = params.get('method', 'pca')  # 目前只支持PCA

if method == 'pca':
    # 只对数值列进行PCA
    numeric_cols = df.select_dtypes(include=[np.number]).columns
    
    if len(numeric_cols) > 1:
        pca = PCA(n_components=min(n_components, len(numeric_cols)))
        X_pca = pca.fit_transform(df[numeric_cols])
        
        # 创建新的数据框
        pca_columns = [f'PC{i+1}' for i in range(X_pca.shape[1])]
        df_pca = pd.DataFrame(X_pca, columns=pca_columns)
        
        # 保留非数值列
        non_numeric_cols = df.select_dtypes(exclude=[np.number]).columns
        for col in non_numeric_cols:
            df_pca[col] = df[col].values
        
        df = df_pca
        
        explained_variance = pca.explained_variance_ratio_
        print(f"PCA降维完成，保留 {len(pca_columns)} 个主成分")
        print(f"解释方差比: {explained_variance.tolist()}")
        print(f"累计解释方差比: {np.cumsum(explained_variance).tolist()}")
    else:
        print("数值特征不足，无法进行PCA降维")
else:
    print("目前只支持PCA降维方法")
`
}

// formatResult 格式化结果
func (t *DataPreprocessingTool) formatResult(result *sanbox.PythonExecutionResult) string {
	resultStr := "数据预处理执行成功:\n\n"

	if result.Stdout != "" {
		resultStr += "处理结果:\n" + result.Stdout + "\n\n"
	}

	if result.Output != nil {
		outputJSON, _ := json.Marshal(result.Output)
		resultStr += "详细信息:\n" + string(outputJSON) + "\n\n"
	}

	resultStr += "✅ 数据预处理完成"
	return resultStr
}
