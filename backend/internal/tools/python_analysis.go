package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"smart-analysis/internal/utils/sanbox"
)

// PythonAnalysisTool Python分析工具（统一的数据分析工具）
type PythonAnalysisTool struct {
	sandbox *sanbox.PythonSandbox
	name    string
	desc    string
}

// NewPythonAnalysisTool 创建Python分析工具
func NewPythonAnalysisTool(sandbox *sanbox.PythonSandbox) *PythonAnalysisTool {
	return &PythonAnalysisTool{
		sandbox: sandbox,
		name:    "python_analysis",
		desc:    "执行Python代码进行数据分析、统计计算、数据处理、机器学习和时间序列分析。支持pandas、numpy、scipy、scikit-learn等数据科学库。",
	}
}

// Info 返回工具信息
func (t *PythonAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"code": {
					Type:     schema.String,
					Desc:     "要执行的Python代码",
					Required: true,
				},
				"analysis_type": {
					Type:     schema.String,
					Desc:     "分析类型：general（通用）、statistical（统计分析）、cleaning（数据清洗）、ml（机器学习）、timeseries（时间序列）",
					Required: false,
				},
				"data_source": {
					Type:     schema.String,
					Desc:     "数据源文件路径（可选）",
					Required: false,
				},
			}),
	}, nil
}

// InvokableRun 执行工具
func (t *PythonAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Code         string `json:"code"`
		AnalysisType string `json:"analysis_type,omitempty"`
		DataSource   string `json:"data_source,omitempty"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	// 根据分析类型添加预处理代码
	finalCode := t.preprocessCode(args.Code, args.AnalysisType, args.DataSource)

	// 执行Python代码
	result, err := t.sandbox.ExecutePython(finalCode)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "执行失败: " + result.Error, nil
	}

	// 格式化结果
	return t.formatResult(result), nil
}

// preprocessCode 预处理代码，根据分析类型添加相应的工具函数
func (t *PythonAnalysisTool) preprocessCode(code, analysisType, dataSource string) string {
	prelude := `
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from scipy import stats
import warnings
warnings.filterwarnings('ignore')

# 设置中文字体
plt.rcParams['font.sans-serif'] = ['SimHei', 'Arial Unicode MS', 'DejaVu Sans']
plt.rcParams['axes.unicode_minus'] = False

`

	// 如果指定了数据源，添加数据加载代码
	if dataSource != "" {
		prelude += fmt.Sprintf(`
# 自动加载数据
try:
    if '%s'.endswith('.csv'):
        df = pd.read_csv('%s')
    elif '%s'.endswith(('.xlsx', '.xls')):
        df = pd.read_excel('%s')
    elif '%s'.endswith('.json'):
        df = pd.read_json('%s')
    else:
        print("不支持的文件格式，请手动加载数据")
except Exception as e:
    print(f"数据加载失败: {e}")

`, dataSource, dataSource, dataSource, dataSource, dataSource, dataSource)
	}

	switch analysisType {
	case "statistical":
		prelude += `
# 统计分析辅助函数
def describe_data(df):
    """数据描述性统计"""
    return {
        'shape': df.shape,
        'dtypes': df.dtypes.to_dict(),
        'null_counts': df.isnull().sum().to_dict(),
        'describe': df.describe().to_dict()
    }

def correlation_analysis(df, method='pearson'):
    """相关性分析"""
    numeric_cols = df.select_dtypes(include=[np.number]).columns
    if len(numeric_cols) > 1:
        return df[numeric_cols].corr(method=method).to_dict()
    return {}

def statistical_tests(df, col1, col2=None):
    """统计检验"""
    results = {}
    if col2 is None:
        # 单样本检验
        if df[col1].dtype in ['int64', 'float64']:
            stat, p_value = stats.normaltest(df[col1].dropna())
            results['normality_test'] = {'statistic': stat, 'p_value': p_value}
    else:
        # 双样本检验
        if df[col1].dtype in ['int64', 'float64'] and df[col2].dtype in ['int64', 'float64']:
            stat, p_value = stats.pearsonr(df[col1].dropna(), df[col2].dropna())
            results['correlation_test'] = {'statistic': stat, 'p_value': p_value}
    return results

`
	case "cleaning":
		prelude += `
# 数据清洗辅助函数
def clean_data(df):
    """基本数据清洗"""
    cleaned_df = df.copy()
    
    # 移除完全重复的行
    cleaned_df = cleaned_df.drop_duplicates()
    
    # 填充数值型列的缺失值（使用中位数）
    numeric_cols = cleaned_df.select_dtypes(include=[np.number]).columns
    for col in numeric_cols:
        cleaned_df[col].fillna(cleaned_df[col].median(), inplace=True)
    
    # 填充类别型列的缺失值（使用众数）
    categorical_cols = cleaned_df.select_dtypes(include=['object']).columns
    for col in categorical_cols:
        mode_val = cleaned_df[col].mode()
        if len(mode_val) > 0:
            cleaned_df[col].fillna(mode_val[0], inplace=True)
    
    return cleaned_df

def detect_outliers(df, column, method='iqr'):
    """异常值检测"""
    if method == 'iqr':
        Q1 = df[column].quantile(0.25)
        Q3 = df[column].quantile(0.75)
        IQR = Q3 - Q1
        lower_bound = Q1 - 1.5 * IQR
        upper_bound = Q3 + 1.5 * IQR
        return df[(df[column] < lower_bound) | (df[column] > upper_bound)]
    return pd.DataFrame()

`
	case "ml":
		prelude += `
# 机器学习辅助函数
try:
    from sklearn.model_selection import train_test_split, cross_val_score
    from sklearn.preprocessing import StandardScaler, LabelEncoder
    from sklearn.ensemble import RandomForestClassifier, RandomForestRegressor
    from sklearn.linear_model import LogisticRegression, LinearRegression
    from sklearn.svm import SVC, SVR
    from sklearn.cluster import KMeans
    from sklearn.metrics import accuracy_score, mean_squared_error, classification_report
    SKLEARN_AVAILABLE = True
except ImportError:
    print("scikit-learn不可用，部分机器学习功能将无法使用")
    SKLEARN_AVAILABLE = False

def prepare_ml_data(df, target_column, test_size=0.2):
    """准备机器学习数据"""
    if not SKLEARN_AVAILABLE:
        return None, None, None, None
    
    X = df.drop(columns=[target_column])
    y = df[target_column]
    
    # 处理类别型特征
    for col in X.select_dtypes(include=['object']).columns:
        le = LabelEncoder()
        X[col] = le.fit_transform(X[col].astype(str))
    
    return train_test_split(X, y, test_size=test_size, random_state=42)

def build_classifier(X_train, y_train, algorithm='rf'):
    """构建分类器"""
    if not SKLEARN_AVAILABLE:
        return None
    
    models = {
        'rf': RandomForestClassifier(random_state=42),
        'lr': LogisticRegression(random_state=42),
        'svm': SVC(random_state=42)
    }
    
    model = models.get(algorithm, RandomForestClassifier(random_state=42))
    model.fit(X_train, y_train)
    return model

def build_regressor(X_train, y_train, algorithm='rf'):
    """构建回归器"""
    if not SKLEARN_AVAILABLE:
        return None
    
    models = {
        'rf': RandomForestRegressor(random_state=42),
        'lr': LinearRegression(),
        'svm': SVR()
    }
    
    model = models.get(algorithm, RandomForestRegressor(random_state=42))
    model.fit(X_train, y_train)
    return model

def perform_clustering(df, n_clusters=3, algorithm='kmeans'):
    """执行聚类分析"""
    if not SKLEARN_AVAILABLE:
        return None
    
    numeric_cols = df.select_dtypes(include=[np.number]).columns
    X = df[numeric_cols]
    
    if algorithm == 'kmeans':
        model = KMeans(n_clusters=n_clusters, random_state=42)
        clusters = model.fit_predict(X)
        return clusters
    
    return None

`
	case "timeseries":
		prelude += `
# 时间序列分析辅助函数
def prepare_timeseries(df, date_column, value_column):
    """准备时间序列数据"""
    ts_df = df[[date_column, value_column]].copy()
    ts_df[date_column] = pd.to_datetime(ts_df[date_column])
    ts_df = ts_df.set_index(date_column)
    ts_df = ts_df.sort_index()
    return ts_df

def decompose_timeseries(ts_data, model='additive'):
    """时间序列分解"""
    try:
        from statsmodels.tsa.seasonal import seasonal_decompose
        decomposition = seasonal_decompose(ts_data, model=model)
        return decomposition
    except ImportError:
        print("statsmodels不可用，时间序列分解功能无法使用")
        return None

def calculate_moving_average(ts_data, window=7):
    """计算移动平均"""
    return ts_data.rolling(window=window).mean()

def detect_trend(ts_data):
    """趋势检测"""
    try:
        from scipy.stats import linregress
        x = np.arange(len(ts_data))
        y = ts_data.values.flatten()
        slope, intercept, r_value, p_value, std_err = linregress(x, y)
        return {
            'slope': slope,
            'r_squared': r_value**2,
            'p_value': p_value,
            'trend': 'increasing' if slope > 0 else 'decreasing' if slope < 0 else 'stable'
        }
    except Exception as e:
        print(f"趋势检测失败: {e}")
        return None

`
	}

	return prelude + "\n" + code
}

// formatResult 格式化执行结果
func (t *PythonAnalysisTool) formatResult(result *sanbox.PythonExecutionResult) string {
	resultStr := "分析执行成功:\n\n"

	if result.Stdout != "" {
		resultStr += "执行输出:\n" + result.Stdout + "\n\n"
	}

	if result.Output != nil {
		outputJSON, _ := json.Marshal(result.Output)
		resultStr += "结构化结果:\n" + string(outputJSON) + "\n\n"
	}

	if result.ImagePath != "" {
		resultStr += "生成图片: " + result.ImagePath + "\n\n"
	}

	// 添加执行统计信息
	if result.Success {
		resultStr += "✅ 代码执行成功完成"
	} else {
		resultStr += "❌ 代码执行失败: " + result.Error
	}

	return resultStr
}
