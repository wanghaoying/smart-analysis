package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// MLAnalysisTool 机器学习分析工具
type MLAnalysisTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewMLAnalysisTool 创建机器学习分析工具
func NewMLAnalysisTool(sandbox *sanbox.PythonSandbox) *MLAnalysisTool {
	return &MLAnalysisTool{
		sandbox: sandbox,
		name:    "ml_analysis",
		desc:    "机器学习分析工具，支持分类、回归、聚类算法以及模型评估和交叉验证。包括随机森林、SVM、逻辑回归、线性回归、K-means聚类等算法。",
	}
}

// Info 返回工具信息
func (t *MLAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"task_type": {
					Type:     schema.String,
					Desc:     "任务类型: classification（分类）, regression（回归）, clustering（聚类）, evaluation（模型评估）",
					Required: true,
				},
				"algorithm": {
					Type:     schema.String,
					Desc:     "算法类型: rf（随机森林）, svm（支持向量机）, lr（逻辑回归/线性回归）, kmeans（K均值聚类）",
					Required: true,
				},
				"target_column": {
					Type:     schema.String,
					Desc:     "目标列名（分类和回归任务必需）",
					Required: false,
				},
				"feature_columns": {
					Type: schema.Array,
					Desc: "特征列名（可选，默认使用除目标列外的所有数值列）",
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
					Desc:     "算法参数（JSON格式）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *MLAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		TaskType       string   `json:"task_type"`
		Algorithm      string   `json:"algorithm"`
		TargetColumn   string   `json:"target_column,omitempty"`
		FeatureColumns []string `json:"feature_columns,omitempty"`
		FilePath       string   `json:"file_path,omitempty"`
		Parameters     string   `json:"parameters,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 生成机器学习代码
	code := t.generateMLCode(args.TaskType, args.Algorithm, args.TargetColumn, args.FeatureColumns, args.FilePath, args.Parameters)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(code)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "机器学习分析失败: " + result.Error, nil
	}

	// 格式化结果
	return t.formatResult(result), nil
}

// generateMLCode 生成机器学习代码
func (t *MLAnalysisTool) generateMLCode(taskType, algorithm, targetColumn string, featureColumns []string, filePath, parameters string) string {
	code := `
import pandas as pd
import numpy as np
import json
import warnings
warnings.filterwarnings('ignore')

# 机器学习库
from sklearn.model_selection import train_test_split, cross_val_score, GridSearchCV
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.ensemble import RandomForestClassifier, RandomForestRegressor
from sklearn.linear_model import LogisticRegression, LinearRegression
from sklearn.svm import SVC, SVR
from sklearn.cluster import KMeans
from sklearn.metrics import (
    accuracy_score, precision_score, recall_score, f1_score,
    mean_squared_error, mean_absolute_error, r2_score,
    classification_report, confusion_matrix, silhouette_score
)

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
    print(f"数据加载成功，形状: {df.shape}")
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

	// 根据任务类型生成代码
	switch taskType {
	case "classification":
		code += t.generateClassificationCode(algorithm, targetColumn, featureColumns)
	case "regression":
		code += t.generateRegressionCode(algorithm, targetColumn, featureColumns)
	case "clustering":
		code += t.generateClusteringCode(algorithm, featureColumns)
	case "evaluation":
		code += t.generateEvaluationCode(algorithm, targetColumn, featureColumns)
	default:
		code += `
print("不支持的任务类型")
`
	}

	return code
}

// generateClassificationCode 生成分类代码
func (t *MLAnalysisTool) generateClassificationCode(algorithm, targetColumn string, featureColumns []string) string {
	code := fmt.Sprintf(`
# 分类任务
target_col = "%s"
if target_col not in df.columns:
    print(f"目标列 {target_col} 不存在")
    exit()

`, targetColumn)

	if len(featureColumns) > 0 {
		code += fmt.Sprintf(`
# 使用指定的特征列
feature_cols = %s
missing_cols = [col for col in feature_cols if col not in df.columns]
if missing_cols:
    print(f"特征列 {missing_cols} 不存在")
    exit()
X = df[feature_cols]
`, fmt.Sprintf("%q", featureColumns))
	} else {
		code += `
# 使用所有数值列作为特征
X = df.select_dtypes(include=[np.number]).drop(columns=[target_col])
feature_cols = X.columns.tolist()
print(f"自动选择特征列: {feature_cols}")
`
	}

	code += `
y = df[target_col]

# 处理类别型目标变量
if y.dtype == 'object':
    le = LabelEncoder()
    y = le.fit_transform(y)
    print(f"目标变量类别: {le.classes_.tolist()}")

# 处理类别型特征
for col in X.select_dtypes(include=['object']).columns:
    le = LabelEncoder()
    X[col] = le.fit_transform(X[col].astype(str))

# 划分训练测试集
test_size = params.get('test_size', 0.2)
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=test_size, random_state=42)

# 特征标准化（SVM需要）
if "` + algorithm + `" == "svm":
    scaler = StandardScaler()
    X_train = scaler.fit_transform(X_train)
    X_test = scaler.transform(X_test)

`

	// 根据算法生成模型代码
	switch algorithm {
	case "rf":
		code += `
# 随机森林分类器
n_estimators = params.get('n_estimators', 100)
max_depth = params.get('max_depth', None)
model = RandomForestClassifier(n_estimators=n_estimators, max_depth=max_depth, random_state=42)
`
	case "svm":
		code += `
# 支持向量机分类器
C = params.get('C', 1.0)
kernel = params.get('kernel', 'rbf')
model = SVC(C=C, kernel=kernel, random_state=42)
`
	case "lr":
		code += `
# 逻辑回归分类器
C = params.get('C', 1.0)
max_iter = params.get('max_iter', 1000)
model = LogisticRegression(C=C, max_iter=max_iter, random_state=42)
`
	default:
		code += `
# 默认使用随机森林
model = RandomForestClassifier(random_state=42)
`
	}

	code += `
# 训练模型
model.fit(X_train, y_train)

# 预测
y_pred = model.predict(X_test)

# 评估模型
accuracy = accuracy_score(y_test, y_pred)
precision = precision_score(y_test, y_pred, average='weighted')
recall = recall_score(y_test, y_pred, average='weighted')
f1 = f1_score(y_test, y_pred, average='weighted')

# 交叉验证
cv_scores = cross_val_score(model, X, y, cv=5)

# 结果
result = {
    "task_type": "classification",
    "algorithm": "` + algorithm + `",
    "metrics": {
        "accuracy": float(accuracy),
        "precision": float(precision),
        "recall": float(recall),
        "f1_score": float(f1),
        "cv_mean": float(cv_scores.mean()),
        "cv_std": float(cv_scores.std())
    },
    "feature_importance": {},
    "confusion_matrix": confusion_matrix(y_test, y_pred).tolist()
}

# 特征重要性（随机森林）
if hasattr(model, 'feature_importances_'):
    importance_dict = dict(zip(feature_cols, model.feature_importances_))
    result["feature_importance"] = {k: float(v) for k, v in importance_dict.items()}

print("分类模型训练完成:")
print(json.dumps(result, ensure_ascii=False, indent=2))
`

	return code
}

// generateRegressionCode 生成回归代码
func (t *MLAnalysisTool) generateRegressionCode(algorithm, targetColumn string, featureColumns []string) string {
	code := fmt.Sprintf(`
# 回归任务
target_col = "%s"
if target_col not in df.columns:
    print(f"目标列 {target_col} 不存在")
    exit()

`, targetColumn)

	if len(featureColumns) > 0 {
		code += fmt.Sprintf(`
# 使用指定的特征列
feature_cols = %s
missing_cols = [col for col in feature_cols if col not in df.columns]
if missing_cols:
    print(f"特征列 {missing_cols} 不存在")
    exit()
X = df[feature_cols]
`, fmt.Sprintf("%q", featureColumns))
	} else {
		code += `
# 使用所有数值列作为特征
X = df.select_dtypes(include=[np.number]).drop(columns=[target_col])
feature_cols = X.columns.tolist()
print(f"自动选择特征列: {feature_cols}")
`
	}

	code += `
y = df[target_col]

# 处理类别型特征
for col in X.select_dtypes(include=['object']).columns:
    le = LabelEncoder()
    X[col] = le.fit_transform(X[col].astype(str))

# 划分训练测试集
test_size = params.get('test_size', 0.2)
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=test_size, random_state=42)

# 特征标准化（SVM需要）
if "` + algorithm + `" == "svm":
    scaler = StandardScaler()
    X_train = scaler.fit_transform(X_train)
    X_test = scaler.transform(X_test)

`

	// 根据算法生成模型代码
	switch algorithm {
	case "rf":
		code += `
# 随机森林回归器
n_estimators = params.get('n_estimators', 100)
max_depth = params.get('max_depth', None)
model = RandomForestRegressor(n_estimators=n_estimators, max_depth=max_depth, random_state=42)
`
	case "svm":
		code += `
# 支持向量机回归器
C = params.get('C', 1.0)
kernel = params.get('kernel', 'rbf')
model = SVR(C=C, kernel=kernel)
`
	case "lr":
		code += `
# 线性回归器
model = LinearRegression()
`
	default:
		code += `
# 默认使用随机森林
model = RandomForestRegressor(random_state=42)
`
	}

	code += `
# 训练模型
model.fit(X_train, y_train)

# 预测
y_pred = model.predict(X_test)

# 评估模型
mse = mean_squared_error(y_test, y_pred)
mae = mean_absolute_error(y_test, y_pred)
r2 = r2_score(y_test, y_pred)

# 交叉验证（负MSE）
cv_scores = cross_val_score(model, X, y, cv=5, scoring='neg_mean_squared_error')

# 结果
result = {
    "task_type": "regression",
    "algorithm": "` + algorithm + `",
    "metrics": {
        "mse": float(mse),
        "rmse": float(np.sqrt(mse)),
        "mae": float(mae),
        "r2_score": float(r2),
        "cv_mean": float(-cv_scores.mean()),
        "cv_std": float(cv_scores.std())
    },
    "feature_importance": {}
}

# 特征重要性（随机森林）
if hasattr(model, 'feature_importances_'):
    importance_dict = dict(zip(feature_cols, model.feature_importances_))
    result["feature_importance"] = {k: float(v) for k, v in importance_dict.items()}

print("回归模型训练完成:")
print(json.dumps(result, ensure_ascii=False, indent=2))
`

	return code
}

// generateClusteringCode 生成聚类代码
func (t *MLAnalysisTool) generateClusteringCode(algorithm string, featureColumns []string) string {
	code := ""

	if len(featureColumns) > 0 {
		code += fmt.Sprintf(`
# 使用指定的特征列进行聚类
feature_cols = %s
missing_cols = [col for col in feature_cols if col not in df.columns]
if missing_cols:
    print(f"特征列 {missing_cols} 不存在")
    exit()
X = df[feature_cols]
`, fmt.Sprintf("%q", featureColumns))
	} else {
		code += `
# 使用所有数值列进行聚类
X = df.select_dtypes(include=[np.number])
feature_cols = X.columns.tolist()
print(f"自动选择特征列: {feature_cols}")
`
	}

	code += `
# 处理类别型特征
for col in X.select_dtypes(include=['object']).columns:
    le = LabelEncoder()
    X[col] = le.fit_transform(X[col].astype(str))

# 特征标准化
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)

`

	switch algorithm {
	case "kmeans":
		code += `
# K-means聚类
n_clusters = params.get('n_clusters', 3)
random_state = params.get('random_state', 42)
model = KMeans(n_clusters=n_clusters, random_state=random_state)

# 聚类
clusters = model.fit_predict(X_scaled)

# 评估聚类效果
silhouette_avg = silhouette_score(X_scaled, clusters)

# 计算聚类中心
cluster_centers = model.cluster_centers_

# 结果
result = {
    "task_type": "clustering",
    "algorithm": "kmeans",
    "n_clusters": int(n_clusters),
    "metrics": {
        "silhouette_score": float(silhouette_avg),
        "inertia": float(model.inertia_)
    },
    "cluster_sizes": [int(x) for x in np.bincount(clusters)],
    "cluster_centers": cluster_centers.tolist()
}

# 将聚类结果添加到原数据框
df['cluster'] = clusters

print("聚类分析完成:")
print(json.dumps(result, ensure_ascii=False, indent=2))
`
	default:
		code += `
print("目前只支持K-means聚类算法")
`
	}

	return code
}

// generateEvaluationCode 生成模型评估代码
func (t *MLAnalysisTool) generateEvaluationCode(algorithm, targetColumn string, featureColumns []string) string {
	return `
# 模型评估和超参数调优
print("模型评估功能开发中...")
# 这里可以添加网格搜索、学习曲线等评估功能
`
}

// formatResult 格式化结果
func (t *MLAnalysisTool) formatResult(result *sanbox.PythonExecutionResult) string {
	resultStr := "机器学习分析执行成功:\n\n"

	if result.Stdout != "" {
		resultStr += "分析结果:\n" + result.Stdout + "\n\n"
	}

	if result.Output != nil {
		outputJSON, _ := json.Marshal(result.Output)
		resultStr += "详细信息:\n" + string(outputJSON) + "\n\n"
	}

	resultStr += "✅ 机器学习分析完成"
	return resultStr
}
